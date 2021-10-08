package logsevents

import (
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/tags"
)

type logsData struct {
	timestamp       uint64
	tokens          data.TokensHandler
	tagsCount       data.CountTags
	accounts        data.AlteredAccountsHandler
	txsMap          map[string]*data.Transaction
	scrsMap         map[string]*data.ScResult
	scDeploys       map[string]*data.ScDeployInfo
	pendingBalances *pendingBalancesProc
	tokensInfo      []*data.TokenInfo
}

func newLogsData(
	timestamp uint64,
	accounts data.AlteredAccountsHandler,
	txs []*data.Transaction,
	scrs []*data.ScResult,
) *logsData {
	ld := &logsData{}

	ld.txsMap = converters.ConvertTxsSliceIntoMap(txs)
	ld.scrsMap = converters.ConvertScrsSliceIntoMap(scrs)
	ld.tagsCount = tags.NewTagsCount()
	ld.tokens = data.NewTokensInfo()
	ld.accounts = accounts
	ld.timestamp = timestamp
	ld.scDeploys = make(map[string]*data.ScDeployInfo)
	ld.pendingBalances = newPendingBalancesProcessor()
	ld.tokensInfo = make([]*data.TokenInfo, 0)

	return ld
}
