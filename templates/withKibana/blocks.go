package withKibana

// Blocks will hold the configuration for the blocks index
var Blocks = Object{
	"index_patterns": Array{
		"blocks-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
		"opendistro.index_state_management.rollover_alias": "blocks",
		"index": Object{
			"sort.field": Array{
				"timestamp", "nonce",
			},
			"sort.order": Array{
				"desc", "desc",
			},
		},
	},
	"mappings": Object{
		"properties": Object{
			"nonce": Object{
				"type": "double",
			},
			"timestamp": Object{
				"type":   "date",
				"format": "epoch_second",
			},
			"searchOrder": Object{
				"type": "double",
			},
			"gasProvided": Object{
				"type": "double",
			},
			"gasRefunded": Object{
				"type": "double",
			},
			"gasPenalized": Object{
				"type": "double",
			},
			"maxGasLimit": Object{
				"type": "double",
			},
		},
	},
}
