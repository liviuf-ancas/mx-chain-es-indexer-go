package transactions

import datafield "github.com/ElrondNetwork/elrond-vm-common/parsers/dataField"

// DataFieldParser defines what a data field parser should be able to do
type DataFieldParser interface {
	Parse(dataField []byte, sender, receiver []byte, numOfShards uint32) *datafield.ResponseParseData
}