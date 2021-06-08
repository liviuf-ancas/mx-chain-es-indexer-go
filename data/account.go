package data

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

// AccountInfo holds (serializable) data about an account
type AccountInfo struct {
	Address                  string         `json:"address,omitempty"`
	Nonce                    uint64         `json:"nonce,omitempty"`
	Balance                  string         `json:"balance"`
	BalanceNum               float64        `json:"balanceNum"`
	Token                    string         `json:"token,omitempty"`
	Identifier               string         `json:"identifier,omitempty"`
	TokenNonce               uint64         `json:"tokenNonce,omitempty"`
	Properties               string         `json:"properties,omitempty"`
	IsSender                 bool           `json:"-"`
	IsSmartContract          bool           `json:"-"`
	TotalBalanceWithStake    string         `json:"totalBalanceWithStake,omitempty"`
	TotalBalanceWithStakeNum float64        `json:"totalBalanceWithStakeNum,omitempty"`
	MetaData                 *TokenMetaData `json:"metaData,omitempty"`
}

// TokenMetaData holds data about a token metadata
type TokenMetaData struct {
	Name       string      `json:"name,omitempty"`
	Creator    string      `json:"creator,omitempty"`
	Royalties  uint32      `json:"royalties,omitempty"`
	Hash       []byte      `json:"hash,omitempty"`
	URIs       [][]byte    `json:"uris,omitempty"`
	Attributes *Attributes `json:"attributes,omitempty"`
}

// AccountBalanceHistory represents an entry in the user accounts balances history
type AccountBalanceHistory struct {
	Address         string        `json:"address"`
	Timestamp       time.Duration `json:"timestamp"`
	Balance         string        `json:"balance"`
	Token           string        `json:"token,omitempty"`
	Identifier      string        `json:"identifier,omitempty"`
	TokenNonce      uint64        `json:"tokenNonce,omitempty"`
	IsSender        bool          `json:"isSender,omitempty"`
	IsSmartContract bool          `json:"isSmartContract,omitempty"`
}

// Account is a structure that is needed for regular accounts
type Account struct {
	UserAccount state.UserAccountHandler
	IsSender    bool
}

// AccountESDT is a structure that is needed for ESDT accounts
type AccountESDT struct {
	Account         state.UserAccountHandler
	TokenIdentifier string
	NFTNonce        uint64
	IsSender        bool
	IsNFTOperation  bool
}

// TokenInfo is a structure that is needed to store information about a token
type TokenInfo struct {
	Name      string        `json:"name"`
	Ticker    string        `json:"ticker"`
	Token     string        `json:"token"`
	Issuer    string        `json:"issuer"`
	Type      string        `json:"type"`
	Timestamp time.Duration `json:"timestamp"`
}
