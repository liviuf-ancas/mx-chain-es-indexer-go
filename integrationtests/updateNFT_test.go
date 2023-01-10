//go:build integrationtests

package integrationtests

import (
	"encoding/json"
	"math/big"
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestNFTUpdateMetadata(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	esdtCreateData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{
			URIs: [][]byte{[]byte("uri"), []byte("uri")},
		},
	}
	marshalizedCreate, _ := json.Marshal(esdtCreateData)

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
		ShardID:   1,
	}
	body := &dataBlock.Body{}

	// CREATE NFT data
	address := "erd1w7jyzuj6cv4ngw8luhlkakatjpmjh3ql95lmxphd3vssc4vpymks6k5th7"
	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(1).Bytes(), marshalizedCreate},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	ids := []string{"NFT-abcd-0e"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token.json"), string(genericResponse.Docs[0].Source))

	// Add URIS 1
	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTAddURI),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("uri1"), []byte("uri2")},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	// Add URIS 2 --- results should be the same
	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTAddURI),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("uri1"), []byte("uri2")},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	// Update attributes 1
	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-add-uris.json"), string(genericResponse.Docs[0].Source))

	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTUpdateAttributes),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("tags:test,free,fun;description:This is a test description for an awesome nft;metadata:metadata-test")},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-update-attributes.json"), string(genericResponse.Docs[0].Source))

	// Update attributes 2

	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    decodeAddress(address),
							Identifier: []byte(core.BuiltInFunctionESDTNFTUpdateAttributes),
							Topics:     [][]byte{[]byte("NFT-abcd"), big.NewInt(14).Bytes(), big.NewInt(0).Bytes(), []byte("something")},
						},
						nil,
					},
				},
				TxHash: "h1",
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool, nil, false, testNumOfShards)
	require.Nil(t, err)

	ids = []string{"NFT-abcd-0e"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TokensIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/updateNFT/token-after-update-attributes-second.json"), string(genericResponse.Docs[0].Source))
}
