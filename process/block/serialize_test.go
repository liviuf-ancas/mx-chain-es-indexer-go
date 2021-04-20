package block

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go/core"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/stretchr/testify/require"
)

func TestBlockProcessor_SerializeBlockNilElasticBlockErrors(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	_, err := bp.SerializeBlock(nil)
	require.True(t, errors.Is(err, indexer.ErrNilElasticBlock))
}

func TestBlockProcessor_SerializeBlock(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	buff, err := bp.SerializeBlock(&data.Block{Nonce: 1})
	require.Nil(t, err)
	require.Equal(t, `{"nonce":1,"round":0,"epoch":0,"miniBlocksHashes":null,"notarizedBlocksHashes":null,"proposer":0,"validators":null,"pubKeyBitmap":"","size":0,"sizeTxs":0,"timestamp":0,"stateRootHash":"","prevHash":"","shardId":0,"txCount":0,"accumulatedFees":"","developerFees":"","epochStartBlock":false,"searchOrder":0}`, buff.String())
}

func TestBlockProcessor_SerializeEpochInfoDataErrors(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	_, err := bp.SerializeEpochInfoData(nil)
	require.Equal(t, indexer.ErrNilHeaderHandler, err)

	_, err = bp.SerializeEpochInfoData(&dataBlock.Header{})
	require.True(t, errors.Is(err, indexer.ErrHeaderTypeAssertion))
}

func TestBlockProcessor_SerializeEpochInfoData(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	buff, err := bp.SerializeEpochInfoData(&dataBlock.MetaBlock{
		AccumulatedFeesInEpoch: big.NewInt(1),
		DevFeesInEpoch:         big.NewInt(2),
	})
	require.Nil(t, err)
	require.Equal(t, `{"accumulatedFees":"1","developerFees":"2"}`, buff.String())
}

func TestBlockProcessor_SerializeBlockEpochStartMeta(t *testing.T) {
	t.Parallel()

	bp, _ := NewBlockProcessor(&mock.HasherMock{}, &mock.MarshalizerMock{})

	buff, err := bp.SerializeBlock(&data.Block{
		Hash:            "11cb2a3a28522a11ae646a93aa4d50f87194cead7d6edeb333d502349407b61d",
		Size:            345,
		ShardID:         core.MetachainShardId,
		EpochStartBlock: true,
		SearchOrder:     0x3f2,
		EpochStartInfo: &data.EpochStartInfo{
			TotalSupply:                      "100",
			TotalToDistribute:                "55",
			TotalNewlyMinted:                 "20",
			RewardsPerBlock:                  "15",
			RewardsForProtocolSustainability: "2",
			NodePrice:                        "10",
			PrevEpochStartRound:              222,
			PrevEpochStartHash:               "7072657645706f6368",
		},
	})
	require.Nil(t, err)
	require.Equal(t, `{"nonce":0,"round":0,"epoch":0,"miniBlocksHashes":null,"notarizedBlocksHashes":null,"proposer":0,"validators":null,"pubKeyBitmap":"","size":345,"sizeTxs":0,"timestamp":0,"stateRootHash":"","prevHash":"","shardId":4294967295,"txCount":0,"accumulatedFees":"","developerFees":"","epochStartBlock":true,"searchOrder":1010,"epochStartInfo":{"totalSupply":"100","totalToDistribute":"55","totalNewlyMinted":"20","rewardsPerBlock":"15","rewardsForProtocolSustainability":"2","nodePrice":"10","prevEpochStartRound":222,"prevEpochStartHash":"7072657645706f6368"}}`, buff.String())
}
