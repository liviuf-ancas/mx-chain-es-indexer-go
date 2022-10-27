package logsevents

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/process/elasticproc/tokeninfo"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
)

type argsProcessEvent struct {
	txHashHexEncoded        string
	scDeploys               map[string]*data.ScDeployInfo
	txs                     map[string]*data.Transaction
	event                   coreData.EventHandler
	tokens                  data.TokensHandler
	tokensSupply            data.TokensHandler
	tokenRolesAndProperties *tokeninfo.TokenRolesAndProperties
	timestamp               uint64
	logAddress              []byte
	selfShardID             uint32
	numOfShards             uint32
}

type argOutputProcessEvent struct {
	tokenInfo     *data.TokenInfo
	delegator     *data.Delegator
	processed     bool
	updatePropNFT *data.NFTDataUpdate
}

type eventsProcessor interface {
	processEvent(args *argsProcessEvent) argOutputProcessEvent
}