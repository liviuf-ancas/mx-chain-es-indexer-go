package logsevents

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
)

type argsProcessEvent struct {
	txHashHexEncoded string
	scDeploys        map[string]*data.ScDeployInfo
	event            coreData.EventHandler
	accounts         data.AlteredAccountsHandler
	tokens           data.TokensHandler
	tagsCount        data.CountTags
	timestamp        uint64
	logAddress       []byte
}

type argOutputProcessEvent struct {
	identifier      string
	value           string
	receiver        string
	receiverShardID uint32
	tokenInfo       *data.TokenInfo
	delegator       *data.Delegator
	processed       bool
}

type eventsProcessor interface {
	processEvent(args *argsProcessEvent) argOutputProcessEvent
}