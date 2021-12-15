package withKibana

// Rounds will hold the configuration for the rounds index
var Rounds = Object{
	"index_patterns": Array{
		"rounds-*",
	},
	"settings": Object{
		"number_of_shards":   3,
		"number_of_replicas": 0,
		"opendistro.index_state_management.rollover_alias": "rounds",
		"index": Object{
			"sort.field": Array{
				"timestamp",
			},
			"sort.order": Array{
				"desc",
			},
		},
	},
	"mappings": Object{
		"properties": Object{
			"timestamp": Object{
				"type":   "date",
				"format": "epoch_second",
			},
		},
	},
}
