package accounts

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"testing"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/marshal"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/ElrondNetwork/elrond-vm-common/data/esdt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccountsProcessor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() (int, marshal.Marshalizer, core.PubkeyConverter, state.AccountsAdapter)
		exError  error
	}{
		{
			name: "NegativeDenomination",
			argsFunc: func() (int, marshal.Marshalizer, core.PubkeyConverter, state.AccountsAdapter) {
				return -1, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{}, &mock.AccountsStub{}
			},
			exError: indexer.ErrNegativeDenominationValue,
		},
		{
			name: "NilMarshalizer",
			argsFunc: func() (int, marshal.Marshalizer, core.PubkeyConverter, state.AccountsAdapter) {
				return 11, nil, &mock.PubkeyConverterMock{}, &mock.AccountsStub{}
			},
			exError: indexer.ErrNilMarshalizer,
		},
		{
			name: "NilPubKeyConverter",
			argsFunc: func() (int, marshal.Marshalizer, core.PubkeyConverter, state.AccountsAdapter) {
				return 11, &mock.MarshalizerMock{}, nil, &mock.AccountsStub{}
			},
			exError: indexer.ErrNilPubkeyConverter,
		},
		{
			name: "NilAccounts",
			argsFunc: func() (int, marshal.Marshalizer, core.PubkeyConverter, state.AccountsAdapter) {
				return 11, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{}, nil
			},
			exError: indexer.ErrNilAccountsDB,
		},
		{
			name: "ShouldWork",
			argsFunc: func() (int, marshal.Marshalizer, core.PubkeyConverter, state.AccountsAdapter) {
				return 11, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{}, &mock.AccountsStub{}
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAccountsProcessor(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
		})
	}
}

func TestAccountsProcessor_GetAccountsWithNil(t *testing.T) {
	t.Parallel()

	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})

	regularAccounts, esdtAccounts := ap.GetAccounts(nil)
	require.Len(t, regularAccounts, 0)
	require.Len(t, esdtAccounts, 0)
}

func TestAccountsProcessor_PrepareRegularAccountsMapWithNil(t *testing.T) {
	t.Parallel()

	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})

	accountsInfo := ap.PrepareRegularAccountsMap(nil)
	require.Len(t, accountsInfo, 0)
}

func TestAccountsProcessor_ComputeBalanceAsFloat(t *testing.T) {
	t.Parallel()

	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})
	require.NotNil(t, ap)

	tests := []struct {
		input  *big.Int
		output float64
	}{
		{
			input:  big.NewInt(200000000000000000),
			output: float64(20000000),
		},
		{
			input:  big.NewInt(57777777777),
			output: 5.7777777777,
		},
		{
			input:  big.NewInt(5777779),
			output: 0.0005777779,
		},
		{
			input:  big.NewInt(7),
			output: 0.0000000007,
		},
		{
			input:  big.NewInt(-7),
			output: 0.0,
		},

		{
			input:  big.NewInt(0),
			output: 0.0,
		},
	}

	for _, tt := range tests {
		out := ap.computeBalanceAsFloat(tt.input, ap.balancePrecision)
		assert.Equal(t, tt.output, out)
	}
}

func TestGetESDTInfo_CannotRetriveValueShoudError(t *testing.T) {
	t.Parallel()

	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})
	require.NotNil(t, ap)

	localErr := errors.New("local error")
	wrapAccount := &data.AccountESDT{
		Account: &mock.UserAccountStub{
			DataTrieTrackerCalled: func() state.DataTrieTracker {
				return &mock.DataTrieTrackerStub{
					RetrieveValueCalled: func(key []byte) ([]byte, error) {
						return nil, localErr
					},
				}
			},
		},
		TokenIdentifier: "token",
	}
	_, _, _, err := ap.getESDTInfo(wrapAccount)
	require.Equal(t, localErr, err)
}

func TestGetESDTInfo(t *testing.T) {
	t.Parallel()

	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})
	require.NotNil(t, ap)

	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
	}

	tokenIdentifier := "token-001"
	wrapAccount := &data.AccountESDT{
		Account: &mock.UserAccountStub{
			DataTrieTrackerCalled: func() state.DataTrieTracker {
				return &mock.DataTrieTrackerStub{
					RetrieveValueCalled: func(key []byte) ([]byte, error) {
						return json.Marshal(esdtToken)
					},
				}
			},
		},
		TokenIdentifier: tokenIdentifier,
	}
	balance, prop, _, err := ap.getESDTInfo(wrapAccount)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1000), balance)
	require.Equal(t, hex.EncodeToString([]byte("ok")), prop)
}

func TestGetESDTInfoNFT(t *testing.T) {
	t.Parallel()

	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})
	require.NotNil(t, ap)

	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1),
		Properties: []byte("ok"),
	}

	tokenIdentifier := "token-001"
	wrapAccount := &data.AccountESDT{
		Account: &mock.UserAccountStub{
			DataTrieTrackerCalled: func() state.DataTrieTracker {
				return &mock.DataTrieTrackerStub{
					RetrieveValueCalled: func(key []byte) ([]byte, error) {
						assert.Equal(t, append([]byte("ELRONDesdttoken-001"), 0xa), key)
						return json.Marshal(esdtToken)
					},
				}
			},
		},
		TokenIdentifier: tokenIdentifier,
		IsNFTOperation:  true,
		NFTNonce:        10,
	}
	balance, prop, _, err := ap.getESDTInfo(wrapAccount)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1), balance)
	require.Equal(t, hex.EncodeToString([]byte("ok")), prop)
}

func TestGetESDTInfoNFTWithMetaData(t *testing.T) {
	t.Parallel()

	pubKeyConverter := mock.NewPubkeyConverterMock(32)
	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, pubKeyConverter, &mock.AccountsStub{})
	require.NotNil(t, ap)

	nftName := "Test-nft"
	creator := []byte("010101")
	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1),
		Properties: []byte("ok"),
		TokenMetaData: &esdt.MetaData{
			Nonce:     1,
			Name:      []byte(nftName),
			Creator:   creator,
			Royalties: 2,
		},
	}

	tokenIdentifier := "token-001"
	wrapAccount := &data.AccountESDT{
		Account: &mock.UserAccountStub{
			DataTrieTrackerCalled: func() state.DataTrieTracker {
				return &mock.DataTrieTrackerStub{
					RetrieveValueCalled: func(key []byte) ([]byte, error) {
						assert.Equal(t, append([]byte("ELRONDesdttoken-001"), 0xa), key)
						return json.Marshal(esdtToken)
					},
				}
			},
		},
		TokenIdentifier: tokenIdentifier,
		IsNFTOperation:  true,
		NFTNonce:        10,
	}
	balance, prop, metaData, err := ap.getESDTInfo(wrapAccount)
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1), balance)
	require.Equal(t, hex.EncodeToString([]byte("ok")), prop)
	require.Equal(t, &data.TokenMetaData{
		Name:      nftName,
		Creator:   pubKeyConverter.Encode(creator),
		Royalties: 2,
	}, metaData)
}

func TestAccountsProcessor_GetAccountsEGLDAccounts(t *testing.T) {
	t.Parallel()

	addr := "aaaabbbb"
	mockAccount := &mock.UserAccountStub{}
	accountsStub := &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), accountsStub)
	require.NotNil(t, ap)

	alteredAccounts := data.NewAlteredAccounts()
	alteredAccounts.Add(addr, &data.AlteredAccount{
		IsESDTOperation: false,
		TokenIdentifier: "",
	})

	accounts, esdtAccounts := ap.GetAccounts(alteredAccounts)
	require.Equal(t, 0, len(esdtAccounts))
	require.Equal(t, []*data.Account{
		{UserAccount: mockAccount},
	}, accounts)
}

func TestAccountsProcessor_GetAccountsESDTAccount(t *testing.T) {
	t.Parallel()

	addr := "aaaabbbb"
	mockAccount := &mock.UserAccountStub{}
	accountsStub := &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), accountsStub)
	require.NotNil(t, ap)

	alteredAccounts := data.NewAlteredAccounts()
	alteredAccounts.Add(addr, &data.AlteredAccount{
		IsESDTOperation: true,
		TokenIdentifier: "token",
	})
	accounts, esdtAccounts := ap.GetAccounts(alteredAccounts)
	require.Equal(t, 0, len(accounts))
	require.Equal(t, []*data.AccountESDT{
		{Account: mockAccount, TokenIdentifier: "token"},
	}, esdtAccounts)
}

func TestAccountsProcessor_PrepareAccountsMapEGLD(t *testing.T) {
	t.Parallel()

	addr := string(make([]byte, 32))
	mockAccount := &mock.UserAccountStub{
		GetNonceCalled: func() uint64 {
			return 1
		},
		GetBalanceCalled: func() *big.Int {
			return big.NewInt(1000)
		},
		AddressBytesCalled: func() []byte {
			return []byte(addr)
		},
	}

	egldAccount := &data.Account{
		UserAccount: mockAccount,
		IsSender:    false,
	}

	accountsStub := &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), accountsStub)
	require.NotNil(t, ap)

	res := ap.PrepareRegularAccountsMap([]*data.Account{egldAccount})
	require.Equal(t, map[string]*data.AccountInfo{
		hex.EncodeToString([]byte(addr)): {
			Address:                  hex.EncodeToString([]byte(addr)),
			Nonce:                    1,
			Balance:                  "1000",
			BalanceNum:               ap.computeBalanceAsFloat(big.NewInt(1000), ap.balancePrecision),
			TotalBalanceWithStake:    "1000",
			TotalBalanceWithStakeNum: ap.computeBalanceAsFloat(big.NewInt(1000), ap.balancePrecision),
			IsSmartContract:          true,
		},
	}, res)
}

func TestAccountsProcessor_PrepareAccountsMapESDT(t *testing.T) {
	t.Parallel()

	esdtToken := &esdt.ESDigitalToken{
		Value:      big.NewInt(1000),
		Properties: []byte("ok"),
		TokenMetaData: &esdt.MetaData{
			Creator: []byte("creator"),
		},
	}

	addr := "aaaabbbb"
	mockAccount := &mock.UserAccountStub{
		DataTrieTrackerCalled: func() state.DataTrieTracker {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					return json.Marshal(esdtToken)
				},
			}
		},
		AddressBytesCalled: func() []byte {
			return []byte(addr)
		},
	}
	accountsStub := &mock.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return mockAccount, nil
		},
	}
	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), accountsStub)
	require.NotNil(t, ap)

	accountsESDT := []*data.AccountESDT{
		{Account: mockAccount, TokenIdentifier: "token", IsNFTOperation: true, NFTNonce: 15},
		{Account: mockAccount, TokenIdentifier: "token", IsNFTOperation: true, NFTNonce: 16},
	}
	res := ap.PrepareAccountsMapESDT(accountsESDT)
	require.Len(t, res, 2)

	require.Equal(t, &data.AccountInfo{
		Address:         hex.EncodeToString([]byte(addr)),
		Balance:         "1000",
		BalanceNum:      ap.computeBalanceAsFloat(big.NewInt(1000), ap.balancePrecision),
		TokenName:       "token",
		TokenIdentifier: "token-0f",
		Properties:      hex.EncodeToString([]byte("ok")),
		TokenNonce:      15,
		Data: &data.TokenMetaData{
			Creator: "63726561746f72",
		},
	}, res[hex.EncodeToString([]byte(addr))+"-token-15"])

	require.Equal(t, &data.AccountInfo{
		Address:         hex.EncodeToString([]byte(addr)),
		Balance:         "1000",
		BalanceNum:      ap.computeBalanceAsFloat(big.NewInt(1000), ap.balancePrecision),
		TokenName:       "token",
		TokenIdentifier: "token-10",
		Properties:      hex.EncodeToString([]byte("ok")),
		TokenNonce:      16,
		Data: &data.TokenMetaData{
			Creator: "63726561746f72",
		},
	}, res[hex.EncodeToString([]byte(addr))+"-token-16"])
}

func TestAccountsProcessor_PrepareAccountsHistory(t *testing.T) {
	t.Parallel()

	accounts := map[string]*data.AccountInfo{
		"addr1": {
			Address:    "addr1",
			Balance:    "112",
			TokenName:  "token-112",
			TokenNonce: 10,
			IsSender:   true,
		},
	}

	ap, _ := NewAccountsProcessor(10, &mock.MarshalizerMock{}, mock.NewPubkeyConverterMock(32), &mock.AccountsStub{})

	res := ap.PrepareAccountsHistory(100, accounts)
	accountBalanceHistory := res["addr1_100"]
	require.Equal(t, &data.AccountBalanceHistory{
		Address:    "addr1",
		Timestamp:  100,
		Balance:    "112",
		Token:      "token-112",
		IsSender:   true,
		TokenNonce: 10,
		Identifier: "token-112-0a",
	}, accountBalanceHistory)
}

func TestAccountsProcessor_GetUserAccountErrors(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local error")
	tests := []struct {
		name         string
		argsFunc     func() (int, marshal.Marshalizer, core.PubkeyConverter, state.AccountsAdapter)
		inputAddress string
		exError      error
	}{
		{
			name:    "InvalidAddress",
			exError: localErr,
			argsFunc: func() (int, marshal.Marshalizer, core.PubkeyConverter, state.AccountsAdapter) {
				return 10, &mock.MarshalizerMock{}, &mock.PubkeyConverterStub{
					DecodeCalled: func(humanReadable string) ([]byte, error) {
						return nil, localErr
					}}, &mock.AccountsStub{}
			},
		},
		{
			name:    "CannotLoadAccount",
			exError: localErr,
			argsFunc: func() (int, marshal.Marshalizer, core.PubkeyConverter, state.AccountsAdapter) {
				return 10, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{}, &mock.AccountsStub{
					LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
						return nil, localErr
					},
				}
			},
		},
		{
			name:    "CannotCastAccount",
			exError: indexer.ErrCannotCastAccountHandlerToUserAccount,
			argsFunc: func() (int, marshal.Marshalizer, core.PubkeyConverter, state.AccountsAdapter) {
				return 10, &mock.MarshalizerMock{}, &mock.PubkeyConverterMock{}, &mock.AccountsStub{
					LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
						return nil, nil
					},
				}
			},
		},
	}

	for _, tt := range tests {
		ap, err := NewAccountsProcessor(tt.argsFunc())
		require.Nil(t, err)

		_, err = ap.getUserAccount(tt.inputAddress)
		require.Equal(t, tt.exError, err)
	}
}

func TestAccountsProcessor_ComputeBalanceFloatBigNumber(t *testing.T) {
	t.Parallel()

	processor := &accountsProcessor{
		balancePrecisionESDT:   math.Pow(10, float64(core.MaxInt(numDecimalsInFloatBalanceESDT, 0))),
		dividerForDenomination: math.Pow(10, float64(core.MaxInt(18, 0))),
	}

	floatValue := processor.computeBalanceAsFloat(big.NewInt(1), processor.balancePrecisionESDT)
	require.Equal(t, 1e-18, floatValue)
}
