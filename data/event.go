package data

import "time"

// LogEvent is the dto for the log event structure
type LogEvent struct {
	ID             string        `json:"-"`
	TxHash         string        `json:"txHash"`
	OriginalTxHash string        `json:"originalTxHash,omitempty"`
	LogAddress     string        `json:"logAddress"`
	Address        string        `json:"address"`
	Identifier     string        `json:"identifier"`
	Data           string        `json:"data"`
	AdditionalData []string      `json:"additionalData,omitempty"`
	Topics         []string      `json:"topics"`
	Order          int           `json:"order"`
	ShardID        uint32        `json:"shardID"`
	Timestamp      time.Duration `json:"timestamp,omitempty"`
}
