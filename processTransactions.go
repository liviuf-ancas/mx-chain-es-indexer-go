package indexer

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/receipt"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const (
	// A smart contract action (deploy, call, ...) should have minimum 2 smart contract results
	// exception to this rule are smart contract calls to ESDT contract
	minimumNumberOfSmartContractResults = 2
)

type txDatabaseProcessor struct {
	*commonProcessor
	hasher           hashing.Hasher
	marshalizer      marshal.Marshalizer
	isInImportMode   bool
	shardCoordinator Coordinator
	txFeeCalculator  FeesProcessorHandler
}

func newTxDatabaseProcessor(
	hasher hashing.Hasher,
	marshalizer marshal.Marshalizer,
	addressPubkeyConverter core.PubkeyConverter,
	validatorPubkeyConverter core.PubkeyConverter,
	txFeeCalculator FeesProcessorHandler,
	isInImportMode bool,
	shardCoordinator Coordinator,
) *txDatabaseProcessor {
	return &txDatabaseProcessor{
		hasher:      hasher,
		marshalizer: marshalizer,
		commonProcessor: &commonProcessor{
			addressPubkeyConverter:   addressPubkeyConverter,
			validatorPubkeyConverter: validatorPubkeyConverter,
			txFeeCalculator:          txFeeCalculator,
			shardCoordinator:         shardCoordinator,
		},
		isInImportMode:   isInImportMode,
		shardCoordinator: shardCoordinator,
		txFeeCalculator:  txFeeCalculator,
	}
}

func (tdp *txDatabaseProcessor) prepareTransactionsForDatabase(
	body *block.Body,
	header coreData.HeaderHandler,
	txPool map[string]coreData.TransactionHandler,
	selfShardID uint32,
	logs map[string]coreData.LogHandler,
) ([]*data.Transaction, map[string]struct{}) {
	transactions, rewardsTxs, alteredAddresses := tdp.groupNormalTxsAndRewards(body, txPool, header, selfShardID)
	//we can not iterate smart contract results directly on the miniblocks contained in the block body
	// as some miniblocks might be missing. Example: intra-shard miniblock that holds smart contract results
	scResults := groupSmartContractResults(txPool)
	tdp.addScrsReceiverToAlteredAccounts(alteredAddresses, scResults)

	transactions = tdp.setTransactionSearchOrder(transactions)

	countScResults := make(map[string]int)
	for scHash, scResult := range scResults {
		tx, ok := transactions[string(scResult.OriginalTxHash)]
		if !ok {
			continue
		}

		tx = tdp.addScResultInfoInTx(scHash, scResult, tx)
		countScResults[string(scResult.OriginalTxHash)]++
		delete(scResults, scHash)

		// append child smart contract results
		scrs := findAllChildScrResults(scHash, scResults)
		for childScHash, sc := range scrs {
			tx = tdp.addScResultInfoInTx(childScHash, sc, tx)
			countScResults[string(scResult.OriginalTxHash)]++
		}
	}

	for hash, nrScResult := range countScResults {
		tx, ok := transactions[hash]
		if !ok {
			continue
		}

		if isRelayedTx(tx) || isESDTNFTTransfer(tx) {
			tx.GasUsed = tx.GasLimit
			fee := tdp.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasUsed)
			tx.Fee = fee.String()

			continue
		}

		// ignore invalid transaction because status and gas fields were already set
		if tx.Status == transaction.TxStatusInvalid.String() {
			continue
		}

		if nrScResult < minimumNumberOfSmartContractResults {
			if len(tx.SmartContractResults) > 0 {
				scResultData := tx.SmartContractResults[0].Data
				if isScResultOrLogSuccessful(scResultData) {
					// ESDT contract calls generate just one smart contract result
					continue
				}
			}

			tx.Status = transaction.TxStatusFail.String()

			tx.GasUsed = tx.GasLimit
			fee := tdp.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasUsed)
			tx.Fee = fee.String()
		}
	}

	tdp.processTransactionsLogs(transactions, logs)

	return append(convertMapTxsToSlice(transactions), rewardsTxs...), alteredAddresses
}

func (tdp *txDatabaseProcessor) processTransactionsLogs(
	txs map[string]*data.Transaction,
	logs map[string]coreData.LogHandler,
) {
	for txHash, txLog := range logs {
		tx, ok := txs[txHash]
		if !ok {
			continue
		}

		tdp.processLogEvents(tx, txLog.GetLogEvents())
	}
}

func (tdp *txDatabaseProcessor) processLogEvents(tx *data.Transaction, events []coreData.EventHandler) {
	for _, event := range events {
		identifier := string(event.GetIdentifier())
		if identifier == "writeLog" {
			tx.GasUsed = tx.GasLimit
			fee := tdp.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasLimit)
			tx.Fee = fee.String()
			tx.Status = transaction.TxStatusSuccess.String()
		}

		if identifier == "signalError" {
			tx.GasUsed = tx.GasLimit
			fee := tdp.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasLimit)
			tx.Fee = fee.String()

			tx.Status = transaction.TxStatusFail.String()
		}
	}
}

func (tdp *txDatabaseProcessor) addScrsReceiverToAlteredAccounts(
	alteredAddress map[string]struct{},
	scrs map[string]*smartContractResult.SmartContractResult,
) {
	for _, scr := range scrs {
		shardID := tdp.shardCoordinator.ComputeId(scr.RcvAddr)
		if shardID == tdp.shardCoordinator.SelfId() {
			encodedReceiverAddress := tdp.addressPubkeyConverter.Encode(scr.RcvAddr)
			alteredAddress[encodedReceiverAddress] = struct{}{}
		}
	}
}

func getGasUsedFromReceipt(rec *receipt.Receipt, tx *data.Transaction) uint64 {
	if rec.Data != nil && string(rec.Data) == data.RefundGasMessage {
		// in this gas receipt contains the refunded value
		gasUsed := big.NewInt(0).SetUint64(tx.GasPrice)
		gasUsed.Mul(gasUsed, big.NewInt(0).SetUint64(tx.GasLimit))
		gasUsed.Sub(gasUsed, rec.Value)
		gasUsed.Div(gasUsed, big.NewInt(0).SetUint64(tx.GasPrice))

		return gasUsed.Uint64()
	}

	gasUsed := big.NewInt(0)
	gasUsed = gasUsed.Div(rec.Value, big.NewInt(0).SetUint64(tx.GasPrice))

	return gasUsed.Uint64()
}

func isScResultOrLogSuccessful(scResultData []byte) bool {
	okReturnDataNewVersion := []byte("@" + hex.EncodeToString([]byte(vmcommon.Ok.String())))
	okReturnDataOldVersion := []byte("@" + vmcommon.Ok.String()) // backwards compatible
	return bytes.Contains(scResultData, okReturnDataNewVersion) || bytes.Contains(scResultData, okReturnDataOldVersion)
}

func findAllChildScrResults(hash string, scrs map[string]*smartContractResult.SmartContractResult) map[string]*smartContractResult.SmartContractResult {
	scrResults := make(map[string]*smartContractResult.SmartContractResult)
	for scrHash, scr := range scrs {
		if string(scr.OriginalTxHash) == hash {
			scrResults[scrHash] = scr
			delete(scrs, scrHash)
		}
	}

	return scrResults
}

func (tdp *txDatabaseProcessor) addScResultInfoInTx(scHash string, scr *smartContractResult.SmartContractResult, tx *data.Transaction) *data.Transaction {
	dbScResult := tdp.commonProcessor.convertScResultInDatabaseScr(scHash, scr)
	tx.SmartContractResults = append(tx.SmartContractResults, dbScResult)

	// ignore invalid transaction because status and gas fields was already set
	if tx.Status == transaction.TxStatusInvalid.String() {
		return tx
	}

	if isSCRForSenderWithRefund(dbScResult, tx) {
		refundValue := stringValueToBigInt(dbScResult.Value)
		gasUsed, fee := tdp.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
		tx.GasUsed = gasUsed
		tx.Fee = fee.String()
	}

	return tx
}

func isSCRForSenderWithRefund(dbScResult data.ScResult, tx *data.Transaction) bool {
	isForSender := dbScResult.Receiver == tx.Sender
	isRightNonce := dbScResult.Nonce == tx.Nonce+1
	isFromCurrentTx := dbScResult.PreTxHash == tx.Hash
	isScrDataOk := isDataOk(dbScResult.Data)

	return isFromCurrentTx && isForSender && isRightNonce && isScrDataOk
}

func isDataOk(data []byte) bool {
	okEncoded := hex.EncodeToString([]byte("ok"))
	dataFieldStr := "@" + okEncoded

	return strings.HasPrefix(string(data), dataFieldStr)
}

func (tdp *txDatabaseProcessor) prepareTxLog(log coreData.LogHandler) data.TxLog {
	scAddr := tdp.addressPubkeyConverter.Encode(log.GetAddress())
	events := log.GetLogEvents()

	txLogEvents := make([]data.Event, len(events))
	for i, event := range events {
		txLogEvents[i].Address = hex.EncodeToString(event.GetAddress())
		txLogEvents[i].Data = hex.EncodeToString(event.GetData())
		txLogEvents[i].Identifier = hex.EncodeToString(event.GetIdentifier())

		topics := event.GetTopics()
		txLogEvents[i].Topics = make([]string, len(topics))
		for j, topic := range topics {
			txLogEvents[i].Topics[j] = hex.EncodeToString(topic)
		}
	}

	return data.TxLog{
		Address: scAddr,
		Events:  txLogEvents,
	}
}

func convertMapTxsToSlice(txs map[string]*data.Transaction) []*data.Transaction {
	transactions := make([]*data.Transaction, len(txs))
	i := 0
	for _, tx := range txs {
		transactions[i] = tx
		i++
	}
	return transactions
}

func (tdp *txDatabaseProcessor) groupNormalTxsAndRewards(
	body *block.Body,
	txPool map[string]coreData.TransactionHandler,
	header coreData.HeaderHandler,
	selfShardID uint32,
) (
	map[string]*data.Transaction,
	[]*data.Transaction,
	map[string]struct{},
) {
	alteredAddresses := make(map[string]struct{})
	transactions := make(map[string]*data.Transaction)
	rewardsTxs := make([]*data.Transaction, 0)

	for _, mb := range body.MiniBlocks {
		mbHash, err := core.CalculateHash(tdp.marshalizer, tdp.hasher, mb)
		if err != nil {
			continue
		}

		mbTxStatus := transaction.TxStatusPending.String()
		if selfShardID == mb.ReceiverShardID {
			mbTxStatus = transaction.TxStatusSuccess.String()
		}

		switch mb.Type {
		case block.TxBlock:
			txs := getTransactions(txPool, mb.TxHashes)
			for hash, tx := range txs {
				dbTx := tdp.commonProcessor.buildTransaction(tx, []byte(hash), mbHash, mb, header, mbTxStatus)
				addToAlteredAddresses(dbTx, alteredAddresses, mb, selfShardID, false)
				if tdp.shouldIndex(selfShardID, mb.ReceiverShardID) {
					transactions[hash] = dbTx
				}
				delete(txPool, hash)
			}
		case block.InvalidBlock:
			txs := getTransactions(txPool, mb.TxHashes)
			for hash, tx := range txs {
				dbTx := tdp.commonProcessor.buildTransaction(tx, []byte(hash), mbHash, mb, header, transaction.TxStatusInvalid.String())
				addToAlteredAddresses(dbTx, alteredAddresses, mb, selfShardID, false)

				dbTx.GasUsed = dbTx.GasLimit
				fee := tdp.commonProcessor.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, dbTx.GasUsed)
				dbTx.Fee = fee.String()

				transactions[hash] = dbTx
				delete(txPool, hash)
			}
		case block.RewardsBlock:
			rTxs := getRewardsTransaction(txPool, mb.TxHashes)
			for hash, rtx := range rTxs {
				dbTx := tdp.commonProcessor.buildRewardTransaction(rtx, []byte(hash), mbHash, mb, header, mbTxStatus)
				addToAlteredAddresses(dbTx, alteredAddresses, mb, selfShardID, true)
				if tdp.shouldIndex(selfShardID, mb.ReceiverShardID) {
					rewardsTxs = append(rewardsTxs, dbTx)
				}
				delete(txPool, hash)
			}
		default:
			continue
		}
	}

	return transactions, rewardsTxs, alteredAddresses
}

func (tdp *txDatabaseProcessor) shouldIndex(selfShardID uint32, destinationShardID uint32) bool {
	if !tdp.isInImportMode {
		return true
	}

	return selfShardID == destinationShardID
}

func (tdp *txDatabaseProcessor) setTransactionSearchOrder(transactions map[string]*data.Transaction) map[string]*data.Transaction {
	currentOrder := uint32(0)
	for _, tx := range transactions {
		tx.SearchOrder = currentOrder
		currentOrder++
	}

	return transactions
}

func addToAlteredAddresses(
	tx *data.Transaction,
	alteredAddresses map[string]struct{},
	miniBlock *block.MiniBlock,
	selfShardID uint32,
	isRewardTx bool,
) {
	if selfShardID == miniBlock.SenderShardID && !isRewardTx {
		alteredAddresses[tx.Sender] = struct{}{}
	}

	if tx.Status == transaction.TxStatusInvalid.String() {
		return
	}

	if selfShardID == miniBlock.ReceiverShardID || miniBlock.ReceiverShardID == core.AllShardId {
		alteredAddresses[tx.Receiver] = struct{}{}
	}
}

func groupSmartContractResults(txPool map[string]coreData.TransactionHandler) map[string]*smartContractResult.SmartContractResult {
	scResults := make(map[string]*smartContractResult.SmartContractResult)
	for hash, tx := range txPool {
		scResult, ok := tx.(*smartContractResult.SmartContractResult)
		if !ok {
			continue
		}
		scResults[hash] = scResult
	}

	return scResults
}

func getTransactions(txPool map[string]coreData.TransactionHandler,
	txHashes [][]byte,
) map[string]*transaction.Transaction {
	transactions := make(map[string]*transaction.Transaction)
	for _, txHash := range txHashes {
		txHandler, ok := txPool[string(txHash)]
		if !ok {
			continue
		}

		tx, ok := txHandler.(*transaction.Transaction)
		if !ok {
			continue
		}
		transactions[string(txHash)] = tx
	}
	return transactions
}

func getRewardsTransaction(txPool map[string]coreData.TransactionHandler,
	txHashes [][]byte,
) map[string]*rewardTx.RewardTx {
	rewardsTxs := make(map[string]*rewardTx.RewardTx)
	for _, txHash := range txHashes {
		txHandler, ok := txPool[string(txHash)]
		if !ok {
			continue
		}

		reward, ok := txHandler.(*rewardTx.RewardTx)
		if !ok {
			continue
		}
		rewardsTxs[string(txHash)] = reward
	}
	return rewardsTxs
}
