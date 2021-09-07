package transactions

import (
	"testing"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/stretchr/testify/require"
)

func TestSerializeTokens(t *testing.T) {
	t.Parallel()

	tok1 := &data.TokenInfo{
		Name:      "TokenName",
		Ticker:    "TKN",
		Token:     "TKN-01234",
		Timestamp: 50000,
		Issuer:    "erd123",
		Type:      core.SemiFungibleESDT,
	}
	tok2 := &data.TokenInfo{
		Name:      "Token2",
		Ticker:    "TKN2",
		Token:     "TKN2-51234",
		Issuer:    "erd1231213123",
		Timestamp: 60000,
		Type:      core.NonFungibleESDT,
	}
	tokens := []*data.TokenInfo{tok1, tok2}

	res, err := (&txsDatabaseProcessor{}).SerializeTokens(tokens)
	require.Nil(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "TKN-01234" } }
{"name":"TokenName","ticker":"TKN","token":"TKN-01234","issuer":"erd123","type":"SemiFungibleESDT","timestamp":50000}
{ "index" : { "_id" : "TKN2-51234" } }
{"name":"Token2","ticker":"TKN2","token":"TKN2-51234","issuer":"erd1231213123","type":"NonFungibleESDT","timestamp":60000}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeScResults(t *testing.T) {
	t.Parallel()

	scResult1 := &data.ScResult{
		Hash:          "hash1",
		Nonce:         1,
		GasPrice:      10,
		GasLimit:      50,
		SenderShard:   0,
		ReceiverShard: 1,
	}
	scResult2 := &data.ScResult{
		Hash:          "hash2",
		Nonce:         2,
		GasPrice:      10,
		GasLimit:      50,
		SenderShard:   2,
		ReceiverShard: 3,
	}
	scrs := []*data.ScResult{scResult1, scResult2}

	res, err := (&txsDatabaseProcessor{}).SerializeScResults(scrs)
	require.Nil(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "hash1" } }
{"nonce":1,"gasLimit":50,"gasPrice":10,"value":"","sender":"","receiver":"","senderShard":0,"receiverShard":1,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}
{ "index" : { "_id" : "hash2" } }
{"nonce":2,"gasLimit":50,"gasPrice":10,"value":"","sender":"","receiver":"","senderShard":2,"receiverShard":3,"prevTxHash":"","originalTxHash":"","callType":"","timestamp":0}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeReceipts(t *testing.T) {
	t.Parallel()

	rec1 := &data.Receipt{
		Hash:   "recHash1",
		Sender: "sender1",
		TxHash: "txHash1",
	}
	rec2 := &data.Receipt{
		Hash:   "recHash2",
		Sender: "sender2",
		TxHash: "txHash2",
	}

	recs := []*data.Receipt{rec1, rec2}

	res, err := (&txsDatabaseProcessor{}).SerializeReceipts(recs)
	require.Nil(t, err)
	require.Equal(t, 1, len(res))

	expectedRes := `{ "index" : { "_id" : "recHash1" } }
{"value":"","sender":"sender1","txHash":"txHash1","timestamp":0}
{ "index" : { "_id" : "recHash2" } }
{"value":"","sender":"sender2","txHash":"txHash2","timestamp":0}
`
	require.Equal(t, expectedRes, res[0].String())
}

func TestSerializeTransactionsIntraShardTx(t *testing.T) {
	t.Parallel()

	buffers, err := (&txsDatabaseProcessor{}).SerializeTransactions([]*data.Transaction{{
		Hash:                 "txHash",
		SmartContractResults: []*data.ScResult{{}},
	}}, map[string]string{}, 0)
	require.Nil(t, err)

	expectedBuff := `{ "index" : { "_id" : "txHash", "_type" : "_doc" } }
{"miniBlockHash":"","nonce":0,"round":0,"value":"","receiver":"","sender":"","receiverShard":0,"senderShard":0,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","data":null,"signature":"","timestamp":0,"status":"","searchOrder":0}
`
	require.Equal(t, expectedBuff, buffers[0].String())
}

func TestSerializeTransactionCrossShardTxSource(t *testing.T) {
	t.Parallel()

	buffers, err := (&txsDatabaseProcessor{}).SerializeTransactions([]*data.Transaction{{
		Hash:                 "txHash",
		SenderShard:          0,
		ReceiverShard:        1,
		SmartContractResults: []*data.ScResult{{}},
	}}, map[string]string{}, 0)
	require.Nil(t, err)

	expectedBuff := `{"update":{"_id":"txHash", "_type": "_doc"}}
{"script":{"source":"return"},"upsert":{"miniBlockHash":"","nonce":0,"round":0,"value":"","receiver":"","sender":"","receiverShard":1,"senderShard":0,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","data":null,"signature":"","timestamp":0,"status":"","searchOrder":0}}
`
	require.Equal(t, expectedBuff, buffers[0].String())
}

func TestSerializeTransactionsCrossShardTxDestination(t *testing.T) {
	t.Parallel()

	buffers, err := (&txsDatabaseProcessor{}).SerializeTransactions([]*data.Transaction{{
		Hash:                 "txHash",
		SenderShard:          1,
		ReceiverShard:        0,
		SmartContractResults: []*data.ScResult{{}},
	}}, map[string]string{}, 0)
	require.Nil(t, err)

	expectedBuff := `{ "index" : { "_id" : "txHash", "_type" : "_doc" } }
{"miniBlockHash":"","nonce":0,"round":0,"value":"","receiver":"","sender":"","receiverShard":0,"senderShard":1,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","data":null,"signature":"","timestamp":0,"status":"","searchOrder":0}
`
	require.Equal(t, expectedBuff, buffers[0].String())
}
