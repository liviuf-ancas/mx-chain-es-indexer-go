package elasticproc

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	elasticIndexer "github.com/multiversx/mx-chain-es-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tags"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/tokeninfo"
	logger "github.com/multiversx/mx-chain-logger-go"
)

const (
	docsKey  = "docs"
	errorKey = "error"
	idKey    = "_id"
	foundKey = "found"
)

var (
	log = logger.GetOrCreate("indexer/process")

	indexes = []string{
		elasticIndexer.TransactionsIndex, elasticIndexer.BlockIndex, elasticIndexer.MiniblocksIndex, elasticIndexer.RatingIndex, elasticIndexer.RoundsIndex, elasticIndexer.ValidatorsIndex,
		elasticIndexer.AccountsIndex, elasticIndexer.AccountsHistoryIndex, elasticIndexer.ReceiptsIndex, elasticIndexer.ScResultsIndex, elasticIndexer.AccountsESDTHistoryIndex, elasticIndexer.AccountsESDTIndex,
		elasticIndexer.EpochInfoIndex, elasticIndexer.SCDeploysIndex, elasticIndexer.TokensIndex, elasticIndexer.TagsIndex, elasticIndexer.LogsIndex, elasticIndexer.DelegatorsIndex, elasticIndexer.OperationsIndex,
		elasticIndexer.ESDTsIndex,
	}
)

type objectsMap = map[string]interface{}

// ArgElasticProcessor holds all dependencies required by the elasticProcessor in order to create
// new instances
type ArgElasticProcessor struct {
	BulkRequestMaxSize int
	UseKibana          bool
	IndexTemplates     map[string]*bytes.Buffer
	IndexPolicies      map[string]*bytes.Buffer
	EnabledIndexes     map[string]struct{}
	TransactionsProc   DBTransactionsHandler
	AccountsProc       DBAccountHandler
	BlockProc          DBBlockHandler
	MiniblocksProc     DBMiniblocksHandler
	StatisticsProc     DBStatisticsHandler
	ValidatorsProc     DBValidatorsHandler
	DBClient           DatabaseClientHandler
	LogsAndEventsProc  DBLogsAndEventsHandler
	OperationsProc     OperationsHandler
}

type elasticProcessor struct {
	bulkRequestMaxSize int
	enabledIndexes     map[string]struct{}
	elasticClient      DatabaseClientHandler
	accountsProc       DBAccountHandler
	blockProc          DBBlockHandler
	transactionsProc   DBTransactionsHandler
	miniblocksProc     DBMiniblocksHandler
	statisticsProc     DBStatisticsHandler
	validatorsProc     DBValidatorsHandler
	logsAndEventsProc  DBLogsAndEventsHandler
	operationsProc     OperationsHandler
}

// NewElasticProcessor handles Elasticsearch operations such as initialization, adding, modifying or removing data
func NewElasticProcessor(arguments *ArgElasticProcessor) (*elasticProcessor, error) {
	err := checkArguments(arguments)
	if err != nil {
		return nil, err
	}

	ei := &elasticProcessor{
		elasticClient:      arguments.DBClient,
		enabledIndexes:     arguments.EnabledIndexes,
		accountsProc:       arguments.AccountsProc,
		blockProc:          arguments.BlockProc,
		miniblocksProc:     arguments.MiniblocksProc,
		transactionsProc:   arguments.TransactionsProc,
		statisticsProc:     arguments.StatisticsProc,
		validatorsProc:     arguments.ValidatorsProc,
		logsAndEventsProc:  arguments.LogsAndEventsProc,
		operationsProc:     arguments.OperationsProc,
		bulkRequestMaxSize: arguments.BulkRequestMaxSize,
	}

	err = ei.init(arguments.UseKibana, arguments.IndexTemplates, arguments.IndexPolicies)
	if err != nil {
		return nil, err
	}

	return ei, nil
}

// TODO move all the index create part in a new component
func (ei *elasticProcessor) init(useKibana bool, indexTemplates, _ map[string]*bytes.Buffer) error {
	err := ei.createOpenDistroTemplates(indexTemplates)
	if err != nil {
		return err
	}

	//nolint
	if useKibana {
		// TODO: Re-activate after we think of a solid way to handle forks+rotating indexes
		// err = ei.createIndexPolicies(indexPolicies)
		// if err != nil {
		//	return err
		// }
	}

	err = ei.createIndexTemplates(indexTemplates)
	if err != nil {
		return err
	}

	err = ei.createIndexes()
	if err != nil {
		return err
	}

	err = ei.createAliases()
	if err != nil {
		return err
	}

	return nil
}

// nolint
func (ei *elasticProcessor) createIndexPolicies(indexPolicies map[string]*bytes.Buffer) error {
	indexesPolicies := []string{elasticIndexer.TransactionsPolicy, elasticIndexer.BlockPolicy, elasticIndexer.MiniblocksPolicy, elasticIndexer.RatingPolicy, elasticIndexer.RoundsPolicy, elasticIndexer.ValidatorsPolicy,
		elasticIndexer.AccountsPolicy, elasticIndexer.AccountsESDTPolicy, elasticIndexer.AccountsHistoryPolicy, elasticIndexer.AccountsESDTHistoryPolicy, elasticIndexer.AccountsESDTIndex, elasticIndexer.ReceiptsPolicy, elasticIndexer.ScResultsPolicy}
	for _, indexPolicyName := range indexesPolicies {
		indexPolicy := getTemplateByName(indexPolicyName, indexPolicies)
		if indexPolicy != nil {
			err := ei.elasticClient.CheckAndCreatePolicy(indexPolicyName, indexPolicy)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (ei *elasticProcessor) createOpenDistroTemplates(indexTemplates map[string]*bytes.Buffer) error {
	opendistroTemplate := getTemplateByName(elasticIndexer.OpenDistroIndex, indexTemplates)
	if opendistroTemplate != nil {
		err := ei.elasticClient.CheckAndCreateTemplate(elasticIndexer.OpenDistroIndex, opendistroTemplate)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) createIndexTemplates(indexTemplates map[string]*bytes.Buffer) error {
	for _, index := range indexes {
		indexTemplate := getTemplateByName(index, indexTemplates)
		if indexTemplate != nil {
			err := ei.elasticClient.CheckAndCreateTemplate(index, indexTemplate)
			if err != nil {
				return fmt.Errorf("index: %s, error: %w", index, err)
			}
		}
	}
	return nil
}

func (ei *elasticProcessor) createIndexes() error {

	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-%s", index, elasticIndexer.IndexSuffix)
		err := ei.elasticClient.CheckAndCreateIndex(indexName)
		if err != nil {
			return fmt.Errorf("index: %s, error: %w", index, err)
		}
	}
	return nil
}

func (ei *elasticProcessor) createAliases() error {
	for _, index := range indexes {
		indexName := fmt.Sprintf("%s-%s", index, elasticIndexer.IndexSuffix)
		err := ei.elasticClient.CheckAndCreateAlias(index, indexName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ei *elasticProcessor) getExistingObjMap(hashes []string, index string) (map[string]bool, error) {
	if len(hashes) == 0 {
		return make(map[string]bool), nil
	}

	response := make(objectsMap)
	err := ei.elasticClient.DoMultiGet(hashes, index, false, &response)
	if err != nil {
		return make(map[string]bool), err
	}

	return getDecodedResponseMultiGet(response), nil
}

func getDecodedResponseMultiGet(response objectsMap) map[string]bool {
	founded := make(map[string]bool)
	interfaceSlice, ok := response[docsKey].([]interface{})
	if !ok {
		return founded
	}

	for _, element := range interfaceSlice {
		obj := element.(objectsMap)
		_, ok = obj[errorKey]
		if ok {
			continue
		}
		founded[obj[idKey].(string)] = obj[foundKey].(bool)
	}

	return founded
}

func getTemplateByName(templateName string, templateList map[string]*bytes.Buffer) *bytes.Buffer {
	if template, ok := templateList[templateName]; ok {
		return template
	}

	log.Debug("elasticProcessor.getTemplateByName", "could not find template", templateName)
	return nil
}

// SaveHeader will prepare and save information about a header in elasticsearch server
func (ei *elasticProcessor) SaveHeader(outportBlockWithHeader *outport.OutportBlockWithHeader) error {
	if !ei.isIndexEnabled(elasticIndexer.BlockIndex) {
		return nil
	}

	elasticBlock, err := ei.blockProc.PrepareBlockForDB(outportBlockWithHeader)
	if err != nil {
		return err
	}

	buffSlice := data.NewBufferSlice(ei.bulkRequestMaxSize)
	err = ei.blockProc.SerializeBlock(elasticBlock, buffSlice, elasticIndexer.BlockIndex)
	if err != nil {
		return err
	}

	err = ei.indexEpochInfoData(outportBlockWithHeader.Header, buffSlice)
	if err != nil {
		return err
	}

	return ei.doBulkRequests("", buffSlice.Buffers())
}

func (ei *elasticProcessor) indexEpochInfoData(header coreData.HeaderHandler, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.EpochInfoIndex) ||
		header.GetShardID() != core.MetachainShardId {
		return nil
	}

	return ei.blockProc.SerializeEpochInfoData(header, buffSlice, elasticIndexer.EpochInfoIndex)
}

// RemoveHeader will remove a block from elasticsearch server
func (ei *elasticProcessor) RemoveHeader(header coreData.HeaderHandler) error {
	headerHash, err := ei.blockProc.ComputeHeaderHash(header)
	if err != nil {
		return err
	}

	return ei.elasticClient.DoQueryRemove(
		elasticIndexer.BlockIndex,
		converters.PrepareHashesForQueryRemove([]string{hex.EncodeToString(headerHash)}),
	)
}

// RemoveMiniblocks will remove all miniblocks that are in header from elasticsearch server
func (ei *elasticProcessor) RemoveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	encodedMiniblocksHashes := ei.miniblocksProc.GetMiniblocksHashesHexEncoded(header, body)
	if len(encodedMiniblocksHashes) == 0 {
		return nil
	}

	return ei.elasticClient.DoQueryRemove(
		elasticIndexer.MiniblocksIndex,
		converters.PrepareHashesForQueryRemove(encodedMiniblocksHashes),
	)
}

// RemoveTransactions will remove transaction that are in miniblock from the elasticsearch server
func (ei *elasticProcessor) RemoveTransactions(header coreData.HeaderHandler, body *block.Body) error {
	encodedTxsHashes, encodedScrsHashes := ei.transactionsProc.GetHexEncodedHashesForRemove(header, body)

	err := ei.removeIfHashesNotEmpty(elasticIndexer.TransactionsIndex, encodedTxsHashes)
	if err != nil {
		return err
	}

	err = ei.removeIfHashesNotEmpty(elasticIndexer.ScResultsIndex, encodedScrsHashes)
	if err != nil {
		return err
	}

	err = ei.removeIfHashesNotEmpty(elasticIndexer.OperationsIndex, append(encodedTxsHashes, encodedScrsHashes...))
	if err != nil {
		return err
	}

	err = ei.removeIfHashesNotEmpty(elasticIndexer.LogsIndex, append(encodedTxsHashes, encodedScrsHashes...))
	if err != nil {
		return err
	}

	return ei.updateDelegatorsInCaseOfRevert(header, body)
}

func (ei *elasticProcessor) updateDelegatorsInCaseOfRevert(header coreData.HeaderHandler, body *block.Body) error {
	// delegators index should be updated in case of revert only if the observer is in Metachain and the reverted block has miniblocks
	isMeta := header.GetShardID() == core.MetachainShardId
	hasMiniblocks := len(body.MiniBlocks) > 0
	shouldUpdate := isMeta && hasMiniblocks
	if !shouldUpdate {
		return nil
	}

	delegatorsQuery := ei.logsAndEventsProc.PrepareDelegatorsQueryInCaseOfRevert(header.GetTimeStamp())
	return ei.elasticClient.UpdateByQuery(elasticIndexer.DelegatorsIndex, delegatorsQuery)
}

func (ei *elasticProcessor) removeIfHashesNotEmpty(index string, hashes []string) error {
	if len(hashes) == 0 {
		return nil
	}

	return ei.elasticClient.DoQueryRemove(
		index,
		converters.PrepareHashesForQueryRemove(hashes),
	)
}

// RemoveAccountsESDT will remove data from accountsesdt index and accountsesdthistory
func (ei *elasticProcessor) RemoveAccountsESDT(headerTimestamp uint64, shardID uint32) error {
	query := fmt.Sprintf(`{"query": {"bool": {"must": [{"match": {"shardID": {"query": %d,"operator": "AND"}}},{"match": {"timestamp": {"query": "%d","operator": "AND"}}}]}}}`, shardID, headerTimestamp)
	err := ei.elasticClient.DoQueryRemove(
		elasticIndexer.AccountsESDTIndex,
		bytes.NewBuffer([]byte(query)),
	)
	if err != nil {
		return err
	}

	return ei.elasticClient.DoQueryRemove(
		elasticIndexer.AccountsESDTHistoryIndex,
		bytes.NewBuffer([]byte(query)),
	)
}

// SaveMiniblocks will prepare and save information about miniblocks in elasticsearch server
func (ei *elasticProcessor) SaveMiniblocks(header coreData.HeaderHandler, body *block.Body) error {
	if !ei.isIndexEnabled(elasticIndexer.MiniblocksIndex) {
		return nil
	}

	mbs := ei.miniblocksProc.PrepareDBMiniblocks(header, body)
	if len(mbs) == 0 {
		return nil
	}

	miniblocksInDBMap, err := ei.miniblocksInDBMap(mbs)
	if err != nil {
		log.Warn("elasticProcessor.SaveMiniblocks cannot get indexed miniblocks", "error", err)
	}

	buffSlice := data.NewBufferSlice(ei.bulkRequestMaxSize)
	ei.miniblocksProc.SerializeBulkMiniBlocks(mbs, miniblocksInDBMap, buffSlice, elasticIndexer.MiniblocksIndex, header.GetShardID())

	return ei.doBulkRequests("", buffSlice.Buffers())
}

func (ei *elasticProcessor) miniblocksInDBMap(mbs []*data.Miniblock) (map[string]bool, error) {
	mbsHashes := make([]string, len(mbs))
	for idx := range mbs {
		mbsHashes[idx] = mbs[idx].Hash
	}

	return ei.getExistingObjMap(mbsHashes, elasticIndexer.MiniblocksIndex)
}

// SaveTransactions will prepare and save information about a transactions in elasticsearch server
func (ei *elasticProcessor) SaveTransactions(obh *outport.OutportBlockWithHeader) error {
	headerTimestamp := obh.Header.GetTimeStamp()

	preparedResults := ei.transactionsProc.PrepareTransactionsForDatabase(obh.BlockData.Body, obh.Header, obh.TransactionPool, obh.IsImportDB, obh.NumberOfShards)
	logsData := ei.logsAndEventsProc.ExtractDataFromLogs(obh.TransactionPool.Logs, preparedResults, headerTimestamp, obh.Header.GetShardID(), obh.NumberOfShards)

	buffers := data.NewBufferSlice(ei.bulkRequestMaxSize)
	err := ei.indexTransactions(preparedResults.Transactions, preparedResults.TxHashStatus, obh.Header, buffers)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexOperations(preparedResults.Transactions, preparedResults.TxHashStatus, obh.Header, preparedResults.ScResults, buffers, obh.IsImportDB)
	if err != nil {
		return err
	}

	err = ei.indexTransactionsFeeData(preparedResults.TxHashFee, buffers)
	if err != nil {
		return err
	}

	err = ei.indexNFTCreateInfo(logsData.Tokens, obh.AlteredAccounts, buffers)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexLogs(obh.TransactionPool.Logs, headerTimestamp, buffers)
	if err != nil {
		return err
	}

	err = ei.indexScResults(preparedResults.ScResults, buffers)
	if err != nil {
		return err
	}

	err = ei.indexReceipts(preparedResults.Receipts, buffers)
	if err != nil {
		return err
	}

	tagsCount := tags.NewTagsCount()
	err = ei.indexAlteredAccounts(headerTimestamp, logsData.NFTsDataUpdates, obh.AlteredAccounts, buffers, tagsCount, obh.Header.GetShardID())
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexTagsCount(tagsCount, buffers)
	if err != nil {
		return err
	}

	err = ei.indexTokens(logsData.TokensInfo, logsData.NFTsDataUpdates, buffers)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexDelegators(logsData.Delegators, buffers)
	if err != nil {
		return err
	}

	err = ei.indexNFTBurnInfo(logsData.TokensSupply, buffers)
	if err != nil {
		return err
	}

	err = ei.prepareAndIndexRolesData(logsData.TokenRolesAndProperties, buffers, elasticIndexer.TokensIndex)
	if err != nil {
		return err
	}
	err = ei.prepareAndIndexRolesData(logsData.TokenRolesAndProperties, buffers, elasticIndexer.ESDTsIndex)
	if err != nil {
		return err
	}

	err = ei.indexScDeploys(logsData.ScDeploys, buffers)
	if err != nil {
		return err
	}

	return ei.doBulkRequests("", buffers.Buffers())
}

func (ei *elasticProcessor) prepareAndIndexRolesData(tokenRolesAndProperties *tokeninfo.TokenRolesAndProperties, buffSlice *data.BufferSlice, index string) error {
	if !ei.isIndexEnabled(index) {
		return nil
	}

	return ei.logsAndEventsProc.SerializeRolesData(tokenRolesAndProperties, buffSlice, index)
}

func (ei *elasticProcessor) prepareAndIndexDelegators(delegators map[string]*data.Delegator, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.DelegatorsIndex) {
		return nil
	}

	return ei.logsAndEventsProc.SerializeDelegators(delegators, buffSlice, elasticIndexer.DelegatorsIndex)
}

func (ei *elasticProcessor) indexTransactionsFeeData(txsHashFeeData map[string]*data.FeeData, buffSlice *data.BufferSlice) error {
	if len(txsHashFeeData) == 0 {
		return nil
	}

	err := ei.transactionsProc.SerializeTransactionsFeeData(txsHashFeeData, buffSlice, elasticIndexer.TransactionsIndex)
	if err != nil {
		return nil
	}

	return ei.transactionsProc.SerializeTransactionsFeeData(txsHashFeeData, buffSlice, elasticIndexer.OperationsIndex)
}

func (ei *elasticProcessor) prepareAndIndexLogs(logsAndEvents map[string]*transaction.Log, timestamp uint64, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.LogsIndex) {
		return nil
	}

	logsDB := ei.logsAndEventsProc.PrepareLogsForDB(logsAndEvents, timestamp)

	return ei.logsAndEventsProc.SerializeLogs(logsDB, buffSlice, elasticIndexer.LogsIndex)
}

func (ei *elasticProcessor) indexScDeploys(deployData map[string]*data.ScDeployInfo, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.SCDeploysIndex) {
		return nil
	}

	return ei.logsAndEventsProc.SerializeSCDeploys(deployData, buffSlice, elasticIndexer.SCDeploysIndex)
}

func (ei *elasticProcessor) indexTransactions(txs []*data.Transaction, txHashStatus map[string]string, header coreData.HeaderHandler, bytesBuff *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.TransactionsIndex) {
		return nil
	}

	return ei.transactionsProc.SerializeTransactions(txs, txHashStatus, header.GetShardID(), bytesBuff, elasticIndexer.TransactionsIndex)
}

func (ei *elasticProcessor) prepareAndIndexOperations(
	txs []*data.Transaction,
	txHashStatus map[string]string,
	header coreData.HeaderHandler,
	scrs []*data.ScResult,
	buffSlice *data.BufferSlice,
	isImportDB bool,
) error {
	if !ei.isIndexEnabled(elasticIndexer.OperationsIndex) {
		return nil
	}

	processedTxs, processedSCRs := ei.operationsProc.ProcessTransactionsAndSCRs(txs, scrs, isImportDB, header.GetShardID())

	err := ei.transactionsProc.SerializeTransactions(processedTxs, txHashStatus, header.GetShardID(), buffSlice, elasticIndexer.OperationsIndex)
	if err != nil {
		return err
	}

	return ei.operationsProc.SerializeSCRs(processedSCRs, buffSlice, elasticIndexer.OperationsIndex, header.GetShardID())
}

// SaveValidatorsRating will save validators rating
func (ei *elasticProcessor) SaveValidatorsRating(ratingData *outport.ValidatorsRating) error {
	if !ei.isIndexEnabled(elasticIndexer.RatingIndex) {
		return nil
	}

	buffSlice, err := ei.validatorsProc.SerializeValidatorsRating(ratingData)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.RatingIndex, buffSlice)
}

// SaveShardValidatorsPubKeys will prepare and save information about a shard validators public keys in elasticsearch server
func (ei *elasticProcessor) SaveShardValidatorsPubKeys(validatorsPubKeys *outport.ValidatorsPubKeys) error {
	if !ei.isIndexEnabled(elasticIndexer.ValidatorsIndex) {
		return nil
	}

	buffSlice, err := ei.validatorsProc.PrepareAnSerializeValidatorsPubKeys(validatorsPubKeys)
	if err != nil {
		return err
	}

	return ei.doBulkRequests(elasticIndexer.ValidatorsIndex, buffSlice)
}

// SaveRoundsInfo will prepare and save information about a slice of rounds in elasticsearch server
func (ei *elasticProcessor) SaveRoundsInfo(rounds *outport.RoundsInfo) error {
	if !ei.isIndexEnabled(elasticIndexer.RoundsIndex) {
		return nil
	}

	buff := ei.statisticsProc.SerializeRoundsInfo(rounds)

	return ei.elasticClient.DoBulkRequest(buff, elasticIndexer.RoundsIndex)
}

func (ei *elasticProcessor) indexAlteredAccounts(
	timestamp uint64,
	updatesNFTsData []*data.NFTDataUpdate,
	coreAlteredAccounts map[string]*alteredAccount.AlteredAccount,
	buffSlice *data.BufferSlice,
	tagsCount data.CountTags,
	shardID uint32,
) error {
	regularAccountsToIndex, accountsToIndexESDT := ei.accountsProc.GetAccounts(coreAlteredAccounts)

	err := ei.saveAccounts(timestamp, regularAccountsToIndex, buffSlice, shardID)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDT(timestamp, accountsToIndexESDT, updatesNFTsData, buffSlice, tagsCount, shardID)
}

func (ei *elasticProcessor) saveAccountsESDT(
	timestamp uint64,
	wrappedAccounts []*data.AccountESDT,
	updatesNFTsData []*data.NFTDataUpdate,
	buffSlice *data.BufferSlice,
	tagsCount data.CountTags,
	shardID uint32,
) error {
	accountsESDTMap, tokensData := ei.accountsProc.PrepareAccountsMapESDT(timestamp, wrappedAccounts, tagsCount, shardID)
	err := ei.addTokenTypeAndCurrentOwnerInAccountsESDT(tokensData, accountsESDTMap)
	if err != nil {
		return err
	}

	err = ei.indexAccountsESDT(accountsESDTMap, updatesNFTsData, buffSlice)
	if err != nil {
		return err
	}

	return ei.saveAccountsESDTHistory(timestamp, accountsESDTMap, buffSlice, shardID)
}

func (ei *elasticProcessor) addTokenTypeAndCurrentOwnerInAccountsESDT(tokensData data.TokensHandler, accountsESDTMap map[string]*data.AccountInfo) error {
	if check.IfNil(tokensData) || tokensData.Len() == 0 {
		return nil
	}

	responseTokens := &data.ResponseTokens{}
	err := ei.elasticClient.DoMultiGet(tokensData.GetAllTokens(), elasticIndexer.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	tokensData.AddTypeAndOwnerFromResponse(responseTokens)
	tokensData.PutTypeAndOwnerInAccountsESDT(accountsESDTMap)

	return nil
}

func (ei *elasticProcessor) prepareAndIndexTagsCount(tagsCount data.CountTags, buffSlice *data.BufferSlice) error {
	shouldSkipIndex := !ei.isIndexEnabled(elasticIndexer.TagsIndex) || tagsCount.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	return tagsCount.Serialize(buffSlice, elasticIndexer.TagsIndex)
}

func (ei *elasticProcessor) indexAccountsESDT(
	accountsESDTMap map[string]*data.AccountInfo,
	updatesNFTsData []*data.NFTDataUpdate,
	buffSlice *data.BufferSlice,
) error {
	if !ei.isIndexEnabled(elasticIndexer.AccountsESDTIndex) {
		return nil
	}

	return ei.accountsProc.SerializeAccountsESDT(accountsESDTMap, updatesNFTsData, buffSlice, elasticIndexer.AccountsESDTIndex)
}

func (ei *elasticProcessor) indexNFTCreateInfo(tokensData data.TokensHandler, coreAlteredAccounts map[string]*alteredAccount.AlteredAccount, buffSlice *data.BufferSlice) error {
	shouldSkipIndex := !ei.isIndexEnabled(elasticIndexer.TokensIndex) || tokensData.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	responseTokens := &data.ResponseTokens{}
	err := ei.elasticClient.DoMultiGet(tokensData.GetAllTokens(), elasticIndexer.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	tokensData.AddTypeAndOwnerFromResponse(responseTokens)

	tokens := tokensData.GetAllWithoutMetaESDT()
	ei.accountsProc.PutTokenMedataDataInTokens(tokens, coreAlteredAccounts)

	return ei.accountsProc.SerializeNFTCreateInfo(tokens, buffSlice, elasticIndexer.TokensIndex)
}

func (ei *elasticProcessor) indexNFTBurnInfo(tokensData data.TokensHandler, buffSlice *data.BufferSlice) error {
	shouldSkipIndex := !ei.isIndexEnabled(elasticIndexer.TokensIndex) || tokensData.Len() == 0
	if shouldSkipIndex {
		return nil
	}

	responseTokens := &data.ResponseTokens{}
	err := ei.elasticClient.DoMultiGet(tokensData.GetAllTokens(), elasticIndexer.TokensIndex, true, responseTokens)
	if err != nil {
		return err
	}

	// TODO implement to keep in tokens also the supply
	tokensData.AddTypeAndOwnerFromResponse(responseTokens)
	return ei.logsAndEventsProc.SerializeSupplyData(tokensData, buffSlice, elasticIndexer.TokensIndex)
}

// SaveAccounts will prepare and save information about provided accounts in elasticsearch server
func (ei *elasticProcessor) SaveAccounts(timestamp uint64, accts []*data.Account, shardID uint32) error {
	buffSlice := data.NewBufferSlice(ei.bulkRequestMaxSize)
	return ei.saveAccounts(timestamp, accts, buffSlice, shardID)
}

func (ei *elasticProcessor) saveAccounts(timestamp uint64, accts []*data.Account, buffSlice *data.BufferSlice, shardID uint32) error {
	accountsMap := ei.accountsProc.PrepareRegularAccountsMap(timestamp, accts, shardID)
	err := ei.indexAccounts(accountsMap, elasticIndexer.AccountsIndex, buffSlice)
	if err != nil {
		return err
	}

	return ei.saveAccountsHistory(timestamp, accountsMap, buffSlice, shardID)
}

func (ei *elasticProcessor) indexAccounts(accountsMap map[string]*data.AccountInfo, index string, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(index) {
		return nil
	}

	return ei.serializeAndIndexAccounts(accountsMap, index, buffSlice)
}

func (ei *elasticProcessor) serializeAndIndexAccounts(accountsMap map[string]*data.AccountInfo, index string, buffSlice *data.BufferSlice) error {
	return ei.accountsProc.SerializeAccounts(accountsMap, buffSlice, index)
}

func (ei *elasticProcessor) saveAccountsESDTHistory(timestamp uint64, accountsInfoMap map[string]*data.AccountInfo, buffSlice *data.BufferSlice, shardID uint32) error {
	if !ei.isIndexEnabled(elasticIndexer.AccountsESDTHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(timestamp, accountsInfoMap, shardID)

	return ei.serializeAndIndexAccountsHistory(accountsMap, elasticIndexer.AccountsESDTHistoryIndex, buffSlice)
}

func (ei *elasticProcessor) saveAccountsHistory(timestamp uint64, accountsInfoMap map[string]*data.AccountInfo, buffSlice *data.BufferSlice, shardID uint32) error {
	if !ei.isIndexEnabled(elasticIndexer.AccountsHistoryIndex) {
		return nil
	}

	accountsMap := ei.accountsProc.PrepareAccountsHistory(timestamp, accountsInfoMap, shardID)

	return ei.serializeAndIndexAccountsHistory(accountsMap, elasticIndexer.AccountsHistoryIndex, buffSlice)
}

func (ei *elasticProcessor) serializeAndIndexAccountsHistory(accountsMap map[string]*data.AccountBalanceHistory, index string, buffSlice *data.BufferSlice) error {
	return ei.accountsProc.SerializeAccountsHistory(accountsMap, buffSlice, index)
}

func (ei *elasticProcessor) indexScResults(scrs []*data.ScResult, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.ScResultsIndex) {
		return nil
	}

	return ei.transactionsProc.SerializeScResults(scrs, buffSlice, elasticIndexer.ScResultsIndex)
}

func (ei *elasticProcessor) indexReceipts(receipts []*data.Receipt, buffSlice *data.BufferSlice) error {
	if !ei.isIndexEnabled(elasticIndexer.ReceiptsIndex) {
		return nil
	}

	return ei.transactionsProc.SerializeReceipts(receipts, buffSlice, elasticIndexer.ReceiptsIndex)
}

func (ei *elasticProcessor) isIndexEnabled(index string) bool {
	_, isEnabled := ei.enabledIndexes[index]
	return isEnabled
}

func (ei *elasticProcessor) doBulkRequests(index string, buffSlice []*bytes.Buffer) error {
	var err error
	for idx := range buffSlice {
		err = ei.elasticClient.DoBulkRequest(buffSlice[idx], index)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ei *elasticProcessor) IsInterfaceNil() bool {
	return ei == nil
}
