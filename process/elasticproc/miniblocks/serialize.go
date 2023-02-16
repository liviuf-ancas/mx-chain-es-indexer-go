package miniblocks

import (
	"encoding/json"
	"fmt"

	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-es-indexer-go/data"
	"github.com/multiversx/mx-chain-es-indexer-go/process/elasticproc/converters"
)

// SerializeBulkMiniBlocks will serialize the provided miniblocks slice in a way that Elasticsearch expects a bulk request
func (mp *miniblocksProcessor) SerializeBulkMiniBlocks(
	bulkMbs []*data.Miniblock,
	existsInDb map[string]bool,
	buffSlice *data.BufferSlice,
	index string,
	shardID uint32,
) {
	for _, mb := range bulkMbs {
		meta, serializedData, err := mp.prepareMiniblockData(mb, existsInDb[mb.Hash], index, shardID)
		if err != nil {
			log.Warn("miniblocksProcessor.prepareMiniblockData cannot prepare miniblock data", "error", err)
			continue
		}

		err = buffSlice.PutData(meta, serializedData)
		if err != nil {
			log.Warn("miniblocksProcessor.putInBufferMiniblockData cannot prepare miniblock data", "error", err)
			continue
		}
	}
}

func (mp *miniblocksProcessor) prepareMiniblockData(miniblockDB *data.Miniblock, isInDB bool, index string, shardID uint32) ([]byte, []byte, error) {
	mbHash := miniblockDB.Hash
	miniblockDB.Hash = ""

	if !isInDB {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index":"%s", "_id" : "%s"} }%s`, index, converters.JsonEscape(mbHash), "\n"))
		serializedData, err := json.Marshal(miniblockDB)

		return meta, serializedData, err
	}

	// prepare data for update operation
	meta := []byte(fmt.Sprintf(`{ "update" : {"_index":"%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(mbHash), "\n"))
	if shardID == miniblockDB.SenderShardID && miniblockDB.ProcessingTypeOnDestination != block.Processed.String() {
		// prepare for update sender block hash
		serializedData := []byte(fmt.Sprintf(`{ "doc" : { "senderBlockHash" : "%s", "procTypeS": "%s" } }`, converters.JsonEscape(miniblockDB.SenderBlockHash), converters.JsonEscape(miniblockDB.ProcessingTypeOnSource)))

		return meta, serializedData, nil
	}

	// prepare for update receiver block hash
	serializedData := []byte(fmt.Sprintf(`{ "doc" : { "receiverBlockHash" : "%s", "procTypeD": "%s" } }`, converters.JsonEscape(miniblockDB.ReceiverBlockHash), converters.JsonEscape(miniblockDB.ProcessingTypeOnDestination)))

	return meta, serializedData, nil
}
