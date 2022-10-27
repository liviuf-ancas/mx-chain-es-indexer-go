//go:build integrationtests

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer"
	"github.com/ElrondNetwork/elrond-go-core/core"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

func TestAccountBalanceNFTTransfer(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	// ################ CREATE NFT ##########################
	body := &dataBlock.Body{}

	addr := "test-address-balance-1"
	addrHex := hex.EncodeToString([]byte(addr))

	esProc, err := CreateElasticProcessor(esClient)
	require.Nil(t, err)

	header := &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
		ShardID:   1,
	}

	pool := &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("test-address-balance-1"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics:     [][]byte{[]byte("NFT-abcdef"), big.NewInt(7440483).Bytes(), big.NewInt(1).Bytes()},
						},
						nil,
					},
				},
			},
		},
	}

	coreAlteredAccounts := map[string]*outport.AlteredAccount{
		addrHex: {
			Address: addrHex,
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "NFT-abcdef",
					Nonce:      7440483,
					Balance:    "1000",
				},
			},
		},
	}

	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)

	ids := []string{"746573742d616464726573732d62616c616e63652d31-NFT-abcdef-718863"}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceNftTransfer/balance-nft-after-create.json"), string(genericResponse.Docs[0].Source))

	// ################ TRANSFER NFT ##########################

	header = &dataBlock.Header{
		Round:     51,
		TimeStamp: 5600,
		ShardID:   1,
	}

	pool = &outport.Pool{
		Logs: []*coreData.LogData{
			{
				TxHash: "h1",
				LogHandler: &transaction.Log{
					Events: []*transaction.Event{
						{
							Address:    []byte("test-address-balance-1"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
							Topics:     [][]byte{[]byte("NFT-abcdef"), big.NewInt(7440483).Bytes(), big.NewInt(1).Bytes(), []byte("new-address")},
						},
						nil,
					},
				},
			},
		},
	}

	addrReceiver := "new-address"
	addrReceiverHex := hex.EncodeToString([]byte(addrReceiver))

	esProc, err = CreateElasticProcessor(esClient)
	require.Nil(t, err)

	coreAlteredAccounts = map[string]*outport.AlteredAccount{
		addrHex: {
			Address: addrHex,
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "NFT-abcdef",
					Nonce:      7440483,
					Balance:    "0",
				},
			},
		},
		addrReceiverHex: {
			Address: addrReceiverHex,
			Tokens: []*outport.AccountTokenData{
				{
					Identifier: "NFT-abcdef",
					Nonce:      7440483,
					Balance:    "1000",
				},
			},
		},
	}
	err = esProc.SaveTransactions(body, header, pool, coreAlteredAccounts, false, testNumOfShards)
	require.Nil(t, err)

	ids = []string{"746573742d616464726573732d62616c616e63652d31-NFT-abcdef-718863"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.False(t, genericResponse.Docs[0].Found)

	ids = []string{"6e65772d61646472657373-NFT-abcdef-718863"}
	genericResponse = &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.AccountsESDTIndex, true, genericResponse)
	require.Nil(t, err)
	require.JSONEq(t, readExpectedResult("./testdata/accountsBalanceNftTransfer/balance-nft-after-transfer.json"), string(genericResponse.Docs[0].Source))
}