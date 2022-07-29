package block

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	nodeBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("indexer/process/block")

type blockProcessor struct {
	hasher      hashing.Hasher
	marshalizer marshal.Marshalizer
}

// NewBlockProcessor will create a new instance of block processor
func NewBlockProcessor(hasher hashing.Hasher, marshalizer marshal.Marshalizer) (*blockProcessor, error) {
	if check.IfNil(hasher) {
		return nil, indexer.ErrNilHasher
	}
	if check.IfNil(marshalizer) {
		return nil, indexer.ErrNilMarshalizer
	}

	return &blockProcessor{
		hasher:      hasher,
		marshalizer: marshalizer,
	}, nil
}

// PrepareBlockForDB will prepare a database block and serialize it for database
func (bp *blockProcessor) PrepareBlockForDB(
	header coreData.HeaderHandler,
	signersIndexes []uint64,
	body *block.Body,
	notarizedHeadersHashes []string,
	gasConsumptionData outport.HeaderGasConsumption,
	sizeTxs int,
) (*data.Block, error) {
	if check.IfNil(header) {
		return nil, indexer.ErrNilHeaderHandler
	}
	if body == nil {
		return nil, indexer.ErrNilBlockBody
	}

	blockSizeInBytes, headerHash, err := bp.computeBlockSizeAndHeaderHash(header, body)
	if err != nil {
		return nil, err
	}

	miniblocksHashes := bp.getEncodedMBSHashes(body)
	leaderIndex := bp.getLeaderIndex(signersIndexes)

	numTxs, notarizedTxs := getTxsCount(header)
	elasticBlock := &data.Block{
		Nonce:                 header.GetNonce(),
		Round:                 header.GetRound(),
		Epoch:                 header.GetEpoch(),
		ShardID:               header.GetShardID(),
		Hash:                  hex.EncodeToString(headerHash),
		MiniBlocksHashes:      miniblocksHashes,
		NotarizedBlocksHashes: notarizedHeadersHashes,
		Proposer:              leaderIndex,
		Validators:            signersIndexes,
		PubKeyBitmap:          hex.EncodeToString(header.GetPubKeysBitmap()),
		Size:                  int64(blockSizeInBytes),
		SizeTxs:               int64(sizeTxs),
		Timestamp:             time.Duration(header.GetTimeStamp()),
		TxCount:               numTxs,
		NotarizedTxsCount:     notarizedTxs,
		StateRootHash:         hex.EncodeToString(header.GetRootHash()),
		PrevHash:              hex.EncodeToString(header.GetPrevHash()),
		SearchOrder:           computeBlockSearchOrder(header),
		EpochStartBlock:       header.IsStartOfEpochBlock(),
		GasProvided:           gasConsumptionData.GasProvided,
		GasRefunded:           gasConsumptionData.GasRefunded,
		GasPenalized:          gasConsumptionData.GasPenalized,
		MaxGasLimit:           gasConsumptionData.MaxGasPerBlock,
		AccumulatedFees:       converters.BigIntToString(header.GetAccumulatedFees()),
		DeveloperFees:         converters.BigIntToString(header.GetDeveloperFees()),
	}

	additionalData := header.GetAdditionalData()
	if header.GetAdditionalData() != nil {
		elasticBlock.ScheduledData = &data.ScheduledData{
			ScheduledRootHash:        hex.EncodeToString(additionalData.GetScheduledRootHash()),
			ScheduledAccumulatedFees: converters.BigIntToString(additionalData.GetScheduledAccumulatedFees()),
			ScheduledDeveloperFees:   converters.BigIntToString(additionalData.GetScheduledDeveloperFees()),
			ScheduledGasProvided:     additionalData.GetScheduledGasProvided(),
			ScheduledGasPenalized:    additionalData.GetScheduledGasPenalized(),
			ScheduledGasRefunded:     additionalData.GetScheduledGasRefunded(),
		}
	}

	bp.addEpochStartInfoForMeta(header, elasticBlock)

	return elasticBlock, nil
}

func getTxsCount(header coreData.HeaderHandler) (numTxs, notarizedTxs uint32) {
	numTxs = header.GetTxCount()

	if core.MetachainShardId != header.GetShardID() {
		return numTxs, notarizedTxs
	}

	metaHeader, ok := header.(*nodeBlock.MetaBlock)
	if !ok {
		return 0, 0
	}

	notarizedTxs = metaHeader.TxCount
	numTxs = 0
	for _, mb := range metaHeader.MiniBlockHeaders {
		numTxs += mb.TxCount
	}

	notarizedTxs = notarizedTxs - numTxs

	return numTxs, notarizedTxs
}

func (bp *blockProcessor) addEpochStartInfoForMeta(header coreData.HeaderHandler, block *data.Block) {
	if header.GetShardID() != core.MetachainShardId {
		return
	}

	metaHeader, ok := header.(*nodeBlock.MetaBlock)
	if !ok {
		return
	}

	if !metaHeader.IsStartOfEpochBlock() {
		return
	}

	metaHeaderEconomics := metaHeader.EpochStart.Economics

	block.EpochStartInfo = &data.EpochStartInfo{
		TotalSupply:                      metaHeaderEconomics.TotalSupply.String(),
		TotalToDistribute:                metaHeaderEconomics.TotalToDistribute.String(),
		TotalNewlyMinted:                 metaHeaderEconomics.TotalNewlyMinted.String(),
		RewardsPerBlock:                  metaHeaderEconomics.RewardsPerBlock.String(),
		RewardsForProtocolSustainability: metaHeaderEconomics.RewardsForProtocolSustainability.String(),
		NodePrice:                        metaHeaderEconomics.NodePrice.String(),
		PrevEpochStartRound:              metaHeaderEconomics.PrevEpochStartRound,
		PrevEpochStartHash:               hex.EncodeToString(metaHeaderEconomics.PrevEpochStartHash),
	}
}

func (bp *blockProcessor) getEncodedMBSHashes(body *block.Body) []string {
	miniblocksHashes := make([]string, 0)
	for _, miniblock := range body.MiniBlocks {
		mbHash, errComputeHash := core.CalculateHash(bp.marshalizer, bp.hasher, miniblock)
		if errComputeHash != nil {
			log.Warn("internal error computing hash", "error", errComputeHash)

			continue
		}

		encodedMbHash := hex.EncodeToString(mbHash)
		miniblocksHashes = append(miniblocksHashes, encodedMbHash)
	}

	return miniblocksHashes
}

func (bp *blockProcessor) computeBlockSizeAndHeaderHash(header coreData.HeaderHandler, body *block.Body) (int, []byte, error) {
	headerBytes, err := bp.marshalizer.Marshal(header)
	if err != nil {
		return 0, nil, err
	}
	bodyBytes, err := bp.marshalizer.Marshal(body)
	if err != nil {
		return 0, nil, err
	}

	blockSize := len(headerBytes) + len(bodyBytes)

	headerHash := bp.hasher.Compute(string(headerBytes))

	return blockSize, headerHash, nil
}

func (bp *blockProcessor) getLeaderIndex(signersIndexes []uint64) uint64 {
	if len(signersIndexes) > 0 {
		return signersIndexes[0]
	}

	return 0
}

func computeBlockSearchOrder(header coreData.HeaderHandler) uint64 {
	shardIdentifier := createShardIdentifier(header.GetShardID())
	stringOrder := fmt.Sprintf("1%02d%d", shardIdentifier, header.GetNonce())

	order, err := strconv.ParseUint(stringOrder, 10, 64)
	if err != nil {
		log.Debug("elasticsearchDatabase.computeBlockSearchOrder",
			"could not set uint32 search order", err.Error())
		return 0
	}

	return order
}

func createShardIdentifier(shardID uint32) uint32 {
	shardIdentifier := shardID + 2
	if shardID == core.MetachainShardId {
		shardIdentifier = 1
	}

	return shardIdentifier
}

// ComputeHeaderHash will compute the hash of a provided header
func (bp *blockProcessor) ComputeHeaderHash(header coreData.HeaderHandler) ([]byte, error) {
	return core.CalculateHash(bp.marshalizer, bp.hasher, header)
}
