package converters

import (
	"strings"

	"github.com/ElrondNetwork/elastic-indexer-go/data"
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
)

const (
	ipfsURL = "https://ipfs.io/ipfs/"
)

// PrepareTokenMetaData will prepare the token metadata in a friendly format for database
func PrepareTokenMetaData(pubKeyConverter core.PubkeyConverter, esdtInfo *esdt.ESDigitalToken) *data.TokenMetaData {
	if check.IfNil(pubKeyConverter) {
		return nil
	}

	if esdtInfo == nil || esdtInfo.TokenMetaData == nil {
		return nil
	}

	creatorStr := ""
	if esdtInfo.TokenMetaData.Creator != nil {
		creatorStr = pubKeyConverter.Encode(esdtInfo.TokenMetaData.Creator)
	}

	return &data.TokenMetaData{
		Name:               string(esdtInfo.TokenMetaData.Name),
		Creator:            creatorStr,
		Royalties:          esdtInfo.TokenMetaData.Royalties,
		Hash:               esdtInfo.TokenMetaData.Hash,
		URIs:               esdtInfo.TokenMetaData.URIs,
		Attributes:         esdtInfo.TokenMetaData.Attributes,
		Tags:               ExtractTagsFromAttributes(esdtInfo.TokenMetaData.Attributes),
		MetaData:           ExtractMetaDataFromAttributes(esdtInfo.TokenMetaData.Attributes),
		NonEmptyURIs:       nonEmptyURIs(esdtInfo.TokenMetaData.URIs),
		WhiteListedStorage: whiteListedStorage(esdtInfo.TokenMetaData.URIs),
	}
}

func nonEmptyURIs(uris [][]byte) bool {
	for _, uri := range uris {
		if len(uri) > 0 {
			return true
		}
	}

	return false
}

func whiteListedStorage(uris [][]byte) bool {
	if len(uris) == 0 {
		return false
	}

	return strings.HasPrefix(string(uris[0]), ipfsURL)
}
