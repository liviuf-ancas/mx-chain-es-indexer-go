package workItems_test

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elastic-indexer-go/process/dataindexer/workItems"
	"github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTxs(numTxs int) map[string]data.TransactionHandlerWithGasUsedAndFee {
	txs := make(map[string]data.TransactionHandlerWithGasUsedAndFee, numTxs)
	for i := 0; i < numTxs; i++ {
		tx := &transaction.Transaction{
			Nonce:     uint64(i),
			Value:     big.NewInt(int64(i)),
			RcvAddr:   []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
			SndAddr:   []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
			GasPrice:  10000000,
			GasLimit:  1000,
			Data:      []byte("dasjdksakjdksajdjksajkdjkasjdksajkdasjdksakjdksajdjksajkdjkasjdksajkdasjdksakjdksajdjksajkdjkasjdksajk"),
			Signature: []byte("randomSignatureasdasldkasdsahjgdlhjaskldsjkaldjklasjkdjskladjkl;sajkl"),
		}
		txs[fmt.Sprintf("%d", i)] = outport.NewTransactionHandlerWithGasAndFee(tx, 0, big.NewInt(0))
	}

	return txs
}

func TestItemBlock_SaveNilHeaderShouldRetNil(t *testing.T) {
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{},
		&mock.MarshalizerMock{},
		&outport.ArgsSaveBlockData{},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	assert.Nil(t, err)
}

func TestItemBlock_SaveHeaderShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveHeaderCalled: func(headerHash []byte, header data.HeaderHandler, signersIndexes []uint64, body *dataBlock.Body, notarizedHeadersHashes []string, gasConsumptionData outport.HeaderGasConsumption, txsSize int) error {
				return localErr
			},
		},
		&mock.MarshalizerMock{},
		&outport.ArgsSaveBlockData{
			Header:           &dataBlock.Header{},
			Body:             &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{{}}},
			TransactionsPool: &outport.Pool{},
		},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.True(t, errors.Is(err, localErr))
}

func TestItemBlock_SaveNoMiniblocksShoulCallSaveHeader(t *testing.T) {
	countCalled := 0
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveHeaderCalled: func(headerHash []byte, header data.HeaderHandler, signersIndexes []uint64, body *dataBlock.Body, notarizedHeadersHashes []string, gasConsumptionData outport.HeaderGasConsumption, txsSize int) error {
				countCalled++
				return nil
			},
			SaveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				countCalled++
				return nil
			},
			SaveTransactionsCalled: func(body *dataBlock.Body, header data.HeaderHandler, pool *outport.Pool, coreAlteredAccounts map[string]*outport.AlteredAccount) error {
				countCalled++
				return nil
			},
		},
		&mock.MarshalizerMock{},
		&outport.ArgsSaveBlockData{
			Body:             &dataBlock.Body{},
			Header:           &dataBlock.Header{},
			TransactionsPool: &outport.Pool{},
		},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.NoError(t, err)
	require.Equal(t, 1, countCalled)
}

func TestItemBlock_SaveMiniblocksShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				return localErr
			},
		},
		&mock.MarshalizerMock{},
		&outport.ArgsSaveBlockData{
			Header:           &dataBlock.Header{},
			Body:             &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{{}}},
			TransactionsPool: &outport.Pool{},
		},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.True(t, errors.Is(err, localErr))
}

func TestItemBlock_SaveTransactionsShouldErr(t *testing.T) {
	localErr := errors.New("local err")
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveTransactionsCalled: func(body *dataBlock.Body, header data.HeaderHandler, pool *outport.Pool, coreAlteredAccounts map[string]*outport.AlteredAccount) error {
				return localErr
			},
		},
		&mock.MarshalizerMock{},
		&outport.ArgsSaveBlockData{
			Header:           &dataBlock.Header{},
			Body:             &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{{}}},
			TransactionsPool: &outport.Pool{},
		},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.True(t, errors.Is(err, localErr))
}

func TestItemBlock_SaveShouldWork(t *testing.T) {
	countCalled := 0
	itemBlock := workItems.NewItemBlock(
		&mock.ElasticProcessorStub{
			SaveHeaderCalled: func(headerHash []byte, header data.HeaderHandler, signersIndexes []uint64, body *dataBlock.Body, notarizedHeadersHashes []string, gasConsumptionData outport.HeaderGasConsumption, txsSize int) error {
				countCalled++
				return nil
			},
			SaveMiniblocksCalled: func(header data.HeaderHandler, body *dataBlock.Body) error {
				countCalled++
				return nil
			},
			SaveTransactionsCalled: func(body *dataBlock.Body, header data.HeaderHandler, pool *outport.Pool, coreAlteredAccounts map[string]*outport.AlteredAccount) error {
				countCalled++
				return nil
			},
		},
		&mock.MarshalizerMock{},
		&outport.ArgsSaveBlockData{
			Header:           &dataBlock.Header{},
			Body:             &dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{{}}},
			TransactionsPool: &outport.Pool{},
		},
	)
	require.False(t, itemBlock.IsInterfaceNil())

	err := itemBlock.Save()
	require.NoError(t, err)
	require.Equal(t, 3, countCalled)
}

func TestComputeSizeOfTxsDuration(t *testing.T) {
	res := testing.Benchmark(benchmarkComputeSizeOfTxsDuration)

	fmt.Println("Time to calculate size of txs :", time.Duration(res.NsPerOp()))
}

func benchmarkComputeSizeOfTxsDuration(b *testing.B) {
	numTxs := 20000
	txs := generateTxs(numTxs)
	gogoMarsh := &marshal.GogoProtoMarshalizer{}

	for i := 0; i < b.N; i++ {
		workItems.ComputeSizeOfTxs(gogoMarsh, &outport.Pool{Txs: txs})
	}
}

func TestComputeSizeOfTxs(t *testing.T) {
	const kb = 1024
	numTxs := 20000

	txs := generateTxs(numTxs)
	gogoMarsh := &marshal.GogoProtoMarshalizer{}
	lenTxs := workItems.ComputeSizeOfTxs(gogoMarsh, &outport.Pool{Txs: txs})

	keys := reflect.ValueOf(txs).MapKeys()
	oneTxBytes, _ := gogoMarsh.Marshal(txs[keys[0].String()].GetTxHandler())
	oneTxSize := len(oneTxBytes)
	expectedSize := numTxs * oneTxSize
	expectedSizeDeltaPlus := expectedSize + int(0.01*float64(expectedSize))
	expectedSizeDeltaMinus := expectedSize - int(0.01*float64(expectedSize))

	require.Greater(t, lenTxs, expectedSizeDeltaMinus)
	require.Less(t, lenTxs, expectedSizeDeltaPlus)
	fmt.Printf("Size of %d transactions : %d Kbs \n", numTxs, lenTxs/kb)
}
