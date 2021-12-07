package accounts

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	indexer "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/converters"
	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("indexer/process/accounts")

// accountsProcessor a is structure responsible for processing accounts
type accountsProcessor struct {
	internalMarshalizer    marshal.Marshalizer
	addressPubkeyConverter core.PubkeyConverter
	accountsDB             indexer.AccountsAdapter
	balanceConverter       indexer.BalanceConverter
}

// NewAccountsProcessor will create a new instance of accounts processor
func NewAccountsProcessor(
	marshalizer marshal.Marshalizer,
	addressPubkeyConverter core.PubkeyConverter,
	accountsDB indexer.AccountsAdapter,
	balanceConverter indexer.BalanceConverter,
) (*accountsProcessor, error) {
	if check.IfNil(marshalizer) {
		return nil, indexer.ErrNilMarshalizer
	}
	if check.IfNil(addressPubkeyConverter) {
		return nil, indexer.ErrNilPubkeyConverter
	}
	if check.IfNil(accountsDB) {
		return nil, indexer.ErrNilAccountsDB
	}
	if check.IfNil(balanceConverter) {
		return nil, indexer.ErrNilBalanceConverter
	}

	return &accountsProcessor{
		internalMarshalizer:    marshalizer,
		addressPubkeyConverter: addressPubkeyConverter,
		accountsDB:             accountsDB,
		balanceConverter:       balanceConverter,
	}, nil
}

// GetAccounts will get accounts for regular operations and esdt operations
func (ap *accountsProcessor) GetAccounts(alteredAccounts data.AlteredAccountsHandler) ([]*data.Account, []*data.AccountESDT) {
	regularAccountsToIndex := make([]*data.Account, 0)
	accountsToIndexESDT := make([]*data.AccountESDT, 0)

	if check.IfNil(alteredAccounts) {
		return regularAccountsToIndex, accountsToIndexESDT
	}

	allAlteredAccounts := alteredAccounts.GetAll()
	for address, altered := range allAlteredAccounts {
		userAccount, err := ap.getUserAccount(address)
		if err != nil || check.IfNil(userAccount) {
			log.Warn("cannot get user account", "address", address, "error", err)
			continue
		}

		regularAccounts, esdtAccounts := splitAlteredAccounts(userAccount, altered)

		regularAccountsToIndex = append(regularAccountsToIndex, regularAccounts...)
		accountsToIndexESDT = append(accountsToIndexESDT, esdtAccounts...)
	}

	return regularAccountsToIndex, accountsToIndexESDT
}

func splitAlteredAccounts(userAccount coreData.UserAccountHandler, altered []*data.AlteredAccount) ([]*data.Account, []*data.AccountESDT) {
	regularAccountsToIndex := make([]*data.Account, 0)
	accountsToIndexESDT := make([]*data.AccountESDT, 0)
	for _, info := range altered {
		if info.IsESDTOperation || info.IsNFTOperation {
			accountsToIndexESDT = append(accountsToIndexESDT, &data.AccountESDT{
				Account:         userAccount,
				TokenIdentifier: info.TokenIdentifier,
				IsSender:        info.IsSender,
				IsNFTOperation:  info.IsNFTOperation,
				NFTNonce:        info.NFTNonce,
				Type:            info.Type,
			})
		}

		// if the balance of the ESDT receiver is 0 the receiver is a new account most probably, and we should index it
		ignoreReceiver := !info.BalanceChange && notZeroBalance(userAccount) && !info.IsSender
		if ignoreReceiver {
			continue
		}

		regularAccountsToIndex = append(regularAccountsToIndex, &data.Account{
			UserAccount: userAccount,
			IsSender:    info.IsSender,
		})
	}

	return regularAccountsToIndex, accountsToIndexESDT
}

func notZeroBalance(userAccount coreData.UserAccountHandler) bool {
	if userAccount.GetBalance() == nil {
		return false
	}

	return userAccount.GetBalance().Cmp(big.NewInt(0)) > 0
}

func (ap *accountsProcessor) getUserAccount(address string) (coreData.UserAccountHandler, error) {
	addressBytes, err := ap.addressPubkeyConverter.Decode(address)
	if err != nil {
		return nil, err
	}

	account, err := ap.accountsDB.LoadAccount(addressBytes)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(coreData.UserAccountHandler)
	if !ok {
		return nil, indexer.ErrCannotCastAccountHandlerToUserAccount
	}

	return userAccount, nil
}

// PrepareRegularAccountsMap will prepare a map of regular accounts
func (ap *accountsProcessor) PrepareRegularAccountsMap(accounts []*data.Account) map[string]*data.AccountInfo {
	accountsMap := make(map[string]*data.AccountInfo)
	for _, userAccount := range accounts {
		address := ap.addressPubkeyConverter.Encode(userAccount.UserAccount.AddressBytes())
		balance := userAccount.UserAccount.GetBalance()
		balanceAsFloat := ap.balanceConverter.ComputeBalanceAsFloat(balance)
		acc := &data.AccountInfo{
			Address:                  address,
			Nonce:                    userAccount.UserAccount.GetNonce(),
			Balance:                  balance.String(),
			BalanceNum:               balanceAsFloat,
			IsSender:                 userAccount.IsSender,
			IsSmartContract:          core.IsSmartContractAddress(userAccount.UserAccount.AddressBytes()),
			TotalBalanceWithStake:    balance.String(),
			TotalBalanceWithStakeNum: balanceAsFloat,
		}

		accountsMap[address] = acc
	}

	return accountsMap
}

// PrepareAccountsMapESDT will prepare a map of accounts with ESDT tokens
func (ap *accountsProcessor) PrepareAccountsMapESDT(
	accounts []*data.AccountESDT,
) map[string]*data.AccountInfo {
	accountsESDTMap := make(map[string]*data.AccountInfo)
	for _, accountESDT := range accounts {
		address := ap.addressPubkeyConverter.Encode(accountESDT.Account.AddressBytes())
		balance, properties, tokenMetaData, err := ap.getESDTInfo(accountESDT)
		if err != nil {
			log.Warn("cannot get esdt info from account",
				"address", address,
				"error", err.Error())
			continue
		}

		acc := &data.AccountInfo{
			Address:         address,
			TokenName:       accountESDT.TokenIdentifier,
			TokenIdentifier: converters.ComputeTokenIdentifier(accountESDT.TokenIdentifier, accountESDT.NFTNonce),
			TokenNonce:      accountESDT.NFTNonce,
			Balance:         balance.String(),
			BalanceNum:      ap.balanceConverter.ComputeESDTBalanceAsFloat(balance),
			Properties:      properties,
			IsSender:        accountESDT.IsSender,
			IsSmartContract: core.IsSmartContractAddress(accountESDT.Account.AddressBytes()),
			Data:            tokenMetaData,
		}

		keyInMap := fmt.Sprintf("%s-%s-%d", acc.Address, acc.TokenName, accountESDT.NFTNonce)
		accountsESDTMap[keyInMap] = acc
	}

	return accountsESDTMap
}

// PrepareAccountsHistory will prepare a map of accounts history balance from a map of accounts
func (ap *accountsProcessor) PrepareAccountsHistory(
	timestamp uint64,
	accounts map[string]*data.AccountInfo,
) map[string]*data.AccountBalanceHistory {
	accountsMap := make(map[string]*data.AccountBalanceHistory)
	for _, userAccount := range accounts {
		acc := &data.AccountBalanceHistory{
			Address:         userAccount.Address,
			Balance:         userAccount.Balance,
			Timestamp:       time.Duration(timestamp),
			Token:           userAccount.TokenName,
			TokenNonce:      userAccount.TokenNonce,
			IsSender:        userAccount.IsSender,
			IsSmartContract: userAccount.IsSmartContract,
			Identifier:      converters.ComputeTokenIdentifier(userAccount.TokenName, userAccount.TokenNonce),
		}
		keyInMap := fmt.Sprintf("%s-%s-%d", acc.Address, acc.Token, acc.TokenNonce)
		accountsMap[keyInMap] = acc
	}

	return accountsMap
}

func (ap *accountsProcessor) getESDTInfo(accountESDT *data.AccountESDT) (*big.Int, string, *data.TokenMetaData, error) {
	if accountESDT.TokenIdentifier == "" {
		return big.NewInt(0), "", nil, nil
	}
	if accountESDT.IsNFTOperation && accountESDT.NFTNonce == 0 {
		return big.NewInt(0), "", nil, nil
	}

	tokenKey := []byte(core.ElrondProtectedKeyPrefix + core.ESDTKeyIdentifier + accountESDT.TokenIdentifier)
	if accountESDT.IsNFTOperation {
		nonceBig := big.NewInt(0).SetUint64(accountESDT.NFTNonce)
		tokenKey = append(tokenKey, nonceBig.Bytes()...)
	}

	valueBytes, err := accountESDT.Account.RetrieveValueFromDataTrieTracker(tokenKey)
	if err != nil {
		return nil, "", nil, err
	}

	esdtToken := &esdt.ESDigitalToken{}
	err = ap.internalMarshalizer.Unmarshal(esdtToken, valueBytes)
	if err != nil {
		return nil, "", nil, err
	}

	if esdtToken.Value == nil {
		return big.NewInt(0), "", nil, nil
	}

	tokenMetaData := converters.PrepareTokenMetaData(ap.addressPubkeyConverter, esdtToken)

	return esdtToken.Value, hex.EncodeToString(esdtToken.Properties), tokenMetaData, nil
}