package transactions

import (
	"encoding/hex"
	"strings"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const minNumOfArgumentsNFTTransferORMultiTransfer = 4

type scrsDataToTransactions struct {
	txFeeCalculator indexer.FeesProcessorHandler
}

func newScrsDataToTransactions(txFeeCalculator indexer.FeesProcessorHandler) *scrsDataToTransactions {
	return &scrsDataToTransactions{
		txFeeCalculator: txFeeCalculator,
	}
}

func (st *scrsDataToTransactions) attachSCRsToTransactionsAndReturnSCRsWithoutTx(txs map[string]*data.Transaction, scrs []*data.ScResult) []*data.ScResult {
	scrsWithoutTx := make([]*data.ScResult, 0)
	for _, scr := range scrs {
		decodedOriginalTxHash, err := hex.DecodeString(scr.OriginalTxHash)
		if err != nil {
			continue
		}

		tx, ok := txs[string(decodedOriginalTxHash)]
		if !ok {
			scrsWithoutTx = append(scrsWithoutTx, scr)
			continue
		}

		st.addScResultInfoIntoTx(scr, tx)
	}

	return scrsWithoutTx
}

func (st *scrsDataToTransactions) addScResultInfoIntoTx(dbScResult *data.ScResult, tx *data.Transaction) {
	tx.SmartContractResults = append(tx.SmartContractResults, dbScResult)

	// ignore invalid transaction because status and gas fields was already set
	if tx.Status == transaction.TxStatusInvalid.String() {
		return
	}

	if isSCRForSenderWithRefund(dbScResult, tx) {
		refundValue := stringValueToBigInt(dbScResult.Value)
		gasUsed, fee := st.txFeeCalculator.ComputeGasUsedAndFeeBasedOnRefundValue(tx, refundValue)
		tx.GasUsed = gasUsed
		tx.Fee = fee.String()
	}

	return
}

func (st *scrsDataToTransactions) processTransactionsAfterSCRsWereAttached(transactions map[string]*data.Transaction) {
	for _, tx := range transactions {
		if len(tx.SmartContractResults) == 0 {
			continue
		}

		st.fillTxWithSCRsFields(tx)
	}
}

func (st *scrsDataToTransactions) fillTxWithSCRsFields(tx *data.Transaction) {
	tx.HasSCR = true

	if isRelayedTx(tx) {
		tx.GasUsed = tx.GasLimit
		fee := st.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasUsed)
		tx.Fee = fee.String()

		return
	}

	// ignore invalid transaction because status and gas fields were already set
	if tx.Status == transaction.TxStatusInvalid.String() {
		return
	}

	if hasSuccessfulSCRs(tx) {
		return
	}

	if hasCrossShardPendingTransfer(tx) {
		tx.GasUsed = tx.GasLimit
		fee := st.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasUsed)
		tx.Fee = fee.String()
		return
	}

	tx.Status = transaction.TxStatusFail.String()
	tx.GasUsed = tx.GasLimit
	fee := st.txFeeCalculator.ComputeTxFeeBasedOnGasUsed(tx, tx.GasUsed)
	tx.Fee = fee.String()
}

func hasSuccessfulSCRs(tx *data.Transaction) bool {
	for _, scr := range tx.SmartContractResults {
		if isScResultSuccessful(scr.Data) {
			return true
		}
	}

	return false
}

func hasCrossShardPendingTransfer(tx *data.Transaction) bool {
	for _, scr := range tx.SmartContractResults {
		splitData := strings.Split(string(scr.Data), atSeparator)
		if len(splitData) < 2 {
			return false
		}

		isMultiTransferOrNFTTransfer := splitData[0] == core.BuiltInFunctionESDTNFTTransfer || splitData[0] == core.BuiltInFunctionMultiESDTNFTTransfer
		if !isMultiTransferOrNFTTransfer {
			return false
		}

		if scr.SenderShard != scr.ReceiverShard {
			return true
		}
	}

	return false
}

func (st *scrsDataToTransactions) processSCRsWithoutTx(scrs []*data.ScResult) map[string]string {
	txHashStatus := make(map[string]string)
	for _, scr := range scrs {
		if !isESDTNFTTransferWithUserError(string(scr.Data)) {
			continue
		}

		txHashStatus[scr.OriginalTxHash] = transaction.TxStatusFail.String()
	}

	return txHashStatus
}

func isESDTNFTTransferWithUserError(scrData string) bool {
	splitData := strings.Split(scrData, atSeparator)
	isMultiTransferOrNFTTransfer := splitData[0] == core.BuiltInFunctionESDTNFTTransfer || splitData[0] == core.BuiltInFunctionMultiESDTNFTTransfer
	if !isMultiTransferOrNFTTransfer || len(splitData) < minNumOfArgumentsNFTTransferORMultiTransfer {
		return false
	}

	isUserErr := splitData[len(splitData)-1] == hex.EncodeToString([]byte(vmcommon.UserError.String()))

	return isUserErr
}
