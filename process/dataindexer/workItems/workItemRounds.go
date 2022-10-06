package workItems

import (
	"github.com/ElrondNetwork/elastic-indexer-go/data"
)

type itemRounds struct {
	indexer    saveRounds
	roundsInfo []*data.RoundInfo
}

// NewItemRounds will create a new instance of itemRounds
func NewItemRounds(indexer saveRounds, roundsInfo []*data.RoundInfo) WorkItemHandler {
	return &itemRounds{
		indexer:    indexer,
		roundsInfo: roundsInfo,
	}
}

// Save will save in elasticsearch database information about rounds
func (wir *itemRounds) Save() error {
	err := wir.indexer.SaveRoundsInfo(wir.roundsInfo)
	if err != nil {
		log.Warn("itemRounds.Save", "could not index rounds info", err.Error())
		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (wir *itemRounds) IsInterfaceNil() bool {
	return wir == nil
}
