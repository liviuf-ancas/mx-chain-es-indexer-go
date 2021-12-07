//go:build integration

package integrationtests

import (
	"encoding/hex"
	"math/big"
	"testing"

	indexerdata "github.com/ElrondNetwork/elastic-indexer-go"
	"github.com/ElrondNetwork/elastic-indexer-go/mock"
	coreData "github.com/ElrondNetwork/elrond-go-core/data"
	dataBlock "github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/stretchr/testify/require"
)

const (
	esURL = "http://localhost:9200"

	expectedTxNFTTransfer                  = `{"receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwMDAwMDAwMDAwMDAwMDAwMDUwMGE3YTAyNzcxYWEwNzA5MGU2MDdmMDJiMjVmNGQ2ZDI0MWJmZjMyYjk5MGEy","signature":"","fee":"232880000000000","nonce":79,"gasLimit":5000000,"gasUsed":963500,"miniBlockHash":"83c60064098aa89220b5adc9d71f22b489bfc78cb3dcb516381102d7fec959e8","round":50,"hasScResults":true,"sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"scresults":{"7363724861736831":{"prevTxHash":"6e66745472616e73666572","receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","data":"QDZmNmI=","nonce":80,"callType":"0","gasLimit":0,"originalTxHash":"6e66745472616e73666572","sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"value":"40365000000000","gasPrice":1000000000,"timestamp":5040},"7363724861736832":{"prevTxHash":"6e66745472616e73666572","receiver":"65726431717171717171717171717171717067713537737a77756432717579737563726c71326539376e74647973646c377634656a7a33716e336e6a7134","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwODAxMTIwMjAwMDEyMmU1MDEwODAxMTIwNDQzNGY0ZjRjMWEyMGUwZjNlY2Y1NTVmNjNmMmQxMDEyNDFkZmM5OGI0NjE0YWZmOTI4NGVkZDUwYjQ2YTFjNmUzNmI4MzU1ODc0NGQyMGM0MTMyYTJlNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzMjQzNjg3NDc0NzA3MzNhMmYyZjY5NzA2NjczMmU2OTZmMmY2OTcwNjY3MzJmNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzYTQxNzQ2MTY3NzMzYTQzNmY2ZjZjM2I2ZDY1NzQ2MTY0NjE3NDYxM2E1MTZkNTg2OTQxNzg1MDM5NmU1MzU5NDg1MTU5NTQ1NDYxNDMzNTczNTg3MTdhNGQzMjY0NTg1NjMzNDE0MjUxNjE0NTM1NTI0MTcyNTkzMjc3N2E0ZTY4NjE3OQ==","nonce":0,"callType":"0","gasLimit":0,"originalTxHash":"6e66745472616e73666572","miniBlockHash":"fabb8b32bbafbb7a1f85b12a8892143fbce35ebe14ab67c455d69c9007082d5c","sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":1,"senderShard":0,"value":"<nil>","gasPrice":1000000000,"timestamp":5040}},"value":"0","gasPrice":1000000000,"timestamp":5040,"status":"success","searchOrder":0}`
	expectedTxNFTTransferFailOnDestination = `{"receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwMDAwMDAwMDAwMDAwMDAwMDUwMGE3YTAyNzcxYWEwNzA5MGU2MDdmMDJiMjVmNGQ2ZDI0MWJmZjMyYjk5MGEy","signature":"","fee":"232880000000000","nonce":79,"gasLimit":5000000,"gasUsed":963500,"miniBlockHash":"83c60064098aa89220b5adc9d71f22b489bfc78cb3dcb516381102d7fec959e8","round":50,"hasScResults":true,"sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"scresults":{"7363724861736831":{"prevTxHash":"6e66745472616e73666572","receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","data":"QDZmNmI=","nonce":80,"callType":"0","gasLimit":0,"originalTxHash":"6e66745472616e73666572","sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"value":"40365000000000","gasPrice":1000000000,"timestamp":5040},"7363724861736832":{"prevTxHash":"6e66745472616e73666572","receiver":"65726431717171717171717171717171717067713537737a77756432717579737563726c71326539376e74647973646c377634656a7a33716e336e6a7134","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwODAxMTIwMjAwMDEyMmU1MDEwODAxMTIwNDQzNGY0ZjRjMWEyMGUwZjNlY2Y1NTVmNjNmMmQxMDEyNDFkZmM5OGI0NjE0YWZmOTI4NGVkZDUwYjQ2YTFjNmUzNmI4MzU1ODc0NGQyMGM0MTMyYTJlNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzMjQzNjg3NDc0NzA3MzNhMmYyZjY5NzA2NjczMmU2OTZmMmY2OTcwNjY3MzJmNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzYTQxNzQ2MTY3NzMzYTQzNmY2ZjZjM2I2ZDY1NzQ2MTY0NjE3NDYxM2E1MTZkNTg2OTQxNzg1MDM5NmU1MzU5NDg1MTU5NTQ1NDYxNDMzNTczNTg3MTdhNGQzMjY0NTg1NjMzNDE0MjUxNjE0NTM1NTI0MTcyNTkzMjc3N2E0ZTY4NjE3OQ==","nonce":0,"callType":"0","gasLimit":0,"originalTxHash":"6e66745472616e73666572","miniBlockHash":"fabb8b32bbafbb7a1f85b12a8892143fbce35ebe14ab67c455d69c9007082d5c","sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":1,"senderShard":0,"value":"<nil>","gasPrice":1000000000,"timestamp":5040},"736372576974684572726f72":{"prevTxHash":"31353436656239393730613664633137313062363532383237346537356435303935633133343937303666346666373066353261316635386531313536333136","receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwODAxMTIwMjAwMDEyMmU1MDEwODAxMTIwNDQzNGY0ZjRjMWEyMGUwZjNlY2Y1NTVmNjNmMmQxMDEyNDFkZmM5OGI0NjE0YWZmOTI4NGVkZDUwYjQ2YTFjNmUzNmI4MzU1ODc0NGQyMGM0MTMyYTJlNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzMjQzNjg3NDc0NzA3MzNhMmYyZjY5NzA2NjczMmU2OTZmMmY2OTcwNjY3MzJmNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzYTQxNzQ2MTY3NzMzYTQzNmY2ZjZjM2I2ZDY1NzQ2MTY0NjE3NDYxM2E1MTZkNTg2OTQxNzg1MDM5NmU1MzU5NDg1MTU5NTQ1NDYxNDMzNTczNTg3MTdhNGQzMjY0NTg1NjMzNDE0MjUxNjE0NTM1NTI0MTcyNTkzMjc3N2E0ZTY4NjE3OUA3NTczNjU3MjIwNjU3MjcyNmY3Mg==","nonce":0,"callType":"0","gasLimit":0,"originalTxHash":"6e66745472616e73666572","sender":"65726431717171717171717171717171717067713537737a77756432717579737563726c71326539376e74647973646c377634656a7a33716e336e6a7134","receiverShard":0,"senderShard":0,"value":"<nil>","gasPrice":0,"timestamp":5040}},"value":"0","gasPrice":1000000000,"timestamp":5040,"status":"fail","searchOrder":0}`

	txWithStatusOnly     = `{"miniBlockHash":"","nonce":0,"round":0,"value":"","receiver":"","sender":"","receiverShard":0,"senderShard":0,"gasPrice":0,"gasLimit":0,"gasUsed":0,"fee":"","data":null,"signature":"","timestamp":0,"status":"fail","searchOrder":0,"scresults":{"7363724861736843726f737332":{"prevTxHash":"6e66745472616e7366657243726f7373","receiver":"65726431717171717171717171717171717067713537737a77756432717579737563726c71326539376e74647973646c377634656a7a33716e336e6a7134","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwODAxMTIwMjAwMDEyMmU1MDEwODAxMTIwNDQzNGY0ZjRjMWEyMGUwZjNlY2Y1NTVmNjNmMmQxMDEyNDFkZmM5OGI0NjE0YWZmOTI4NGVkZDUwYjQ2YTFjNmUzNmI4MzU1ODc0NGQyMGM0MTMyYTJlNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzMjQzNjg3NDc0NzA3MzNhMmYyZjY5NzA2NjczMmU2OTZmMmY2OTcwNjY3MzJmNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzYTQxNzQ2MTY3NzMzYTQzNmY2ZjZjM2I2ZDY1NzQ2MTY0NjE3NDYxM2E1MTZkNTg2OTQxNzg1MDM5NmU1MzU5NDg1MTU5NTQ1NDYxNDMzNTczNTg3MTdhNGQzMjY0NTg1NjMzNDE0MjUxNjE0NTM1NTI0MTcyNTkzMjc3N2E0ZTY4NjE3OQ==","nonce":0,"callType":"0","gasLimit":0,"originalTxHash":"6e66745472616e7366657243726f7373","miniBlockHash":"a02d5676ae2a2e7abbcc8f3fce730fd5311df5aa5c0503559fbdd0ae86cc16c9","sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":1,"senderShard":0,"value":"<nil>","gasPrice":1000000000,"timestamp":5040},"736372576974684572726f72":{"prevTxHash":"31353436656239393730613664633137313062363532383237346537356435303935633133343937303666346666373066353261316635386531313536333136","receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwODAxMTIwMjAwMDEyMmU1MDEwODAxMTIwNDQzNGY0ZjRjMWEyMGUwZjNlY2Y1NTVmNjNmMmQxMDEyNDFkZmM5OGI0NjE0YWZmOTI4NGVkZDUwYjQ2YTFjNmUzNmI4MzU1ODc0NGQyMGM0MTMyYTJlNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzMjQzNjg3NDc0NzA3MzNhMmYyZjY5NzA2NjczMmU2OTZmMmY2OTcwNjY3MzJmNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzYTQxNzQ2MTY3NzMzYTQzNmY2ZjZjM2I2ZDY1NzQ2MTY0NjE3NDYxM2E1MTZkNTg2OTQxNzg1MDM5NmU1MzU5NDg1MTU5NTQ1NDYxNDMzNTczNTg3MTdhNGQzMjY0NTg1NjMzNDE0MjUxNjE0NTM1NTI0MTcyNTkzMjc3N2E0ZTY4NjE3OUA3NTczNjU3MjIwNjU3MjcyNmY3Mg==","nonce":0,"callType":"0","gasLimit":0,"originalTxHash":"6e66745472616e7366657243726f7373","sender":"65726431717171717171717171717171717067713537737a77756432717579737563726c71326539376e74647973646c377634656a7a33716e336e6a7134","receiverShard":0,"senderShard":0,"value":"<nil>","gasPrice":0,"timestamp":5040}}}`
	completeTxWithStatus = `{"receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwMDAwMDAwMDAwMDAwMDAwMDUwMGE3YTAyNzcxYWEwNzA5MGU2MDdmMDJiMjVmNGQ2ZDI0MWJmZjMyYjk5MGEy","signature":"","fee":"232880000000000","nonce":79,"gasLimit":5000000,"gasUsed":963500,"miniBlockHash":"db7161a83f08489cba131e55f042536ee49116b622e33e70335a13e51a6c268c","round":50,"hasScResults":true,"sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"scresults":{"7363724861736843726f737332":{"prevTxHash":"6e66745472616e7366657243726f7373","receiver":"65726431717171717171717171717171717067713537737a77756432717579737563726c71326539376e74647973646c377634656a7a33716e336e6a7134","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwODAxMTIwMjAwMDEyMmU1MDEwODAxMTIwNDQzNGY0ZjRjMWEyMGUwZjNlY2Y1NTVmNjNmMmQxMDEyNDFkZmM5OGI0NjE0YWZmOTI4NGVkZDUwYjQ2YTFjNmUzNmI4MzU1ODc0NGQyMGM0MTMyYTJlNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzMjQzNjg3NDc0NzA3MzNhMmYyZjY5NzA2NjczMmU2OTZmMmY2OTcwNjY3MzJmNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzYTQxNzQ2MTY3NzMzYTQzNmY2ZjZjM2I2ZDY1NzQ2MTY0NjE3NDYxM2E1MTZkNTg2OTQxNzg1MDM5NmU1MzU5NDg1MTU5NTQ1NDYxNDMzNTczNTg3MTdhNGQzMjY0NTg1NjMzNDE0MjUxNjE0NTM1NTI0MTcyNTkzMjc3N2E0ZTY4NjE3OQ==","nonce":0,"callType":"0","gasLimit":0,"originalTxHash":"6e66745472616e7366657243726f7373","miniBlockHash":"a02d5676ae2a2e7abbcc8f3fce730fd5311df5aa5c0503559fbdd0ae86cc16c9","sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":1,"senderShard":0,"value":"<nil>","gasPrice":1000000000,"timestamp":5040},"736372576974684572726f72":{"prevTxHash":"31353436656239393730613664633137313062363532383237346537356435303935633133343937303666346666373066353261316635386531313536333136","receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","data":"RVNEVE5GVFRyYW5zZmVyQDQzNGY0YzQ1NDM1NDQ5NDUyZDMyMzY2MzMxMzgzOEAwMUAwMUAwODAxMTIwMjAwMDEyMmU1MDEwODAxMTIwNDQzNGY0ZjRjMWEyMGUwZjNlY2Y1NTVmNjNmMmQxMDEyNDFkZmM5OGI0NjE0YWZmOTI4NGVkZDUwYjQ2YTFjNmUzNmI4MzU1ODc0NGQyMGM0MTMyYTJlNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzMjQzNjg3NDc0NzA3MzNhMmYyZjY5NzA2NjczMmU2OTZmMmY2OTcwNjY3MzJmNTE2ZDVhNzk2MTU2NTYzMTc4NmE3ODY2NDQ2MjU1NTc1YTUwMzE3ODY1NWE3Njc2NTQ0ZDMxNTY2ODZmNjEzNDZmNTk0NzUyNDQ0ZDcwNmQ0YTcyN2E1MjQzNWEzYTQxNzQ2MTY3NzMzYTQzNmY2ZjZjM2I2ZDY1NzQ2MTY0NjE3NDYxM2E1MTZkNTg2OTQxNzg1MDM5NmU1MzU5NDg1MTU5NTQ1NDYxNDMzNTczNTg3MTdhNGQzMjY0NTg1NjMzNDE0MjUxNjE0NTM1NTI0MTcyNTkzMjc3N2E0ZTY4NjE3OUA3NTczNjU3MjIwNjU3MjcyNmY3Mg==","nonce":0,"callType":"0","gasLimit":0,"originalTxHash":"6e66745472616e7366657243726f7373","sender":"65726431717171717171717171717171717067713537737a77756432717579737563726c71326539376e74647973646c377634656a7a33716e336e6a7134","receiverShard":0,"senderShard":0,"value":"<nil>","gasPrice":0,"timestamp":5040},"7363724861736843726f737331":{"prevTxHash":"6e66745472616e7366657243726f7373","receiver":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","data":"QDZmNmI=","nonce":80,"callType":"0","gasLimit":0,"originalTxHash":"6e66745472616e7366657243726f7373","sender":"65726431757265376561323437636c6a3679716a673830756e7a36787a6a686c6a327a776d3467746736737564636d747364326377337873373468617376","receiverShard":0,"senderShard":0,"value":"40365000000000","gasPrice":1000000000,"timestamp":5040}},"value":"0","gasPrice":1000000000,"timestamp":5040,"status":"fail","searchOrder":0}`

	expectedTxNFTTransferSCCallSource      = `{"receiver":"657264316566397878336b336d3839617a6634633478633938777063646e78356830636e787936656d34377236646334616c756430757771783234663530","data":"RVNEVE5GVFRyYW5zZmVyQDRjNGI0NjQxNTI0ZDJkMzM2NjM0NjYzOTYyQDAxNjUzNEA2ZjFlNmYwMWJjNzYyN2Y1YWVAMDAwMDAwMDAwMDAwMDAwMDA1MDBmMWM4ZjJmZGM1OGE2M2M2YjIwMWZjMmVkNjI5OTYyZDNkZmEzM2ZlN2NlYkA2MzZmNmQ3MDZmNzU2ZTY0NTI2NTc3NjE3MjY0NzM1MDcyNmY3ODc5QDAwMDAwMDAwMDAwMDAwMDAwNTAwNGY3OWVjNDRiYjEzMzcyYjVhYzlkOTk2ZDc0OTEyMGY0NzY0Mjc2MjdjZWI=","signature":"","fee":"1904415000000000","nonce":79,"gasLimit":150000000,"gasUsed":150000000,"miniBlockHash":"b30aaa656bf101a7fb87f6c02a9da9e70cd053a79de24f5d14276232757d9766","round":50,"hasScResults":true,"sender":"657264316566397878336b336d3839617a6634633478633938777063646e78356830636e787936656d34377236646334616c756430757771783234663530","receiverShard":0,"senderShard":0,"scresults":{"7363724861736832":{"prevTxHash":"6e66745472616e7366657257697468534343616c6c","receiver":"657264317171717171717171717171717170677137387930396c773933663375647673706c7368647632766b3935376c35766c37306e3473706c72616432","data":"RVNEVE5GVFRyYW5zZmVyQDRjNGI0NjQxNTI0ZDJkMzM2NjM0NjYzOTYyQDAxNjUzNEA2ZjFlNmYwMWJjNzYyN2Y1YWVAMDgwMTEyMGEwMDZmMWU2ZjAxYmM3NjI3ZjVhZTIyNzYwOGI0Y2EwNTFhMjAwMDAwMDAwMDAwMDAwMDAwMDUwMGYxYzhmMmZkYzU4YTYzYzZiMjAxZmMyZWQ2Mjk5NjJkM2RmYTMzZmU3Y2ViMzIwMDNhNGMwMDAwMDAwZTRkNDU1ODQ2NDE1MjRkMmQ2MjM5MzM2NTM2MzAwMDAwMDAwMDAwMDE2YWI1MDAwMDAwMDk2ZjFlNmYwMWJjNzYyN2Y1YWUwMDAwMDAwYzRjNGI0ZDQ1NTgyZDY0MzE2MzM0NjE2MjAwMDAwMDAwMDAwMDczM2IwMDAwMDAwOTZlOTAxOGExZjBjYzlhOWFlZkA2MzZmNmQ3MDZmNzU2ZTY0NTI2NTc3NjE3MjY0NzM1MDcyNmY3ODc5QDAwMDAwMDAwMDAwMDAwMDAwNTAwNGY3OWVjNDRiYjEzMzcyYjVhYzlkOTk2ZDc0OTEyMGY0NzY0Mjc2MjdjZWI=","nonce":0,"callType":"0","gasLimit":148957500,"originalTxHash":"6e66745472616e7366657257697468534343616c6c","miniBlockHash":"fabb8b32bbafbb7a1f85b12a8892143fbce35ebe14ab67c455d69c9007082d5c","sender":"657264316566397878336b336d3839617a6634633478633938777063646e78356830636e787936656d34377236646334616c756430757771783234663530","receiverShard":1,"senderShard":0,"value":"<nil>","gasPrice":1000000000,"timestamp":5040}},"value":"0","gasPrice":1000000000,"timestamp":5040,"status":"success","searchOrder":0}`
	expectedTxNFTTransferSCCallAfterRefund = `{"miniBlockHash":"b30aaa656bf101a7fb87f6c02a9da9e70cd053a79de24f5d14276232757d9766","nonce":79,"round":50,"value":"0","receiver":"657264316566397878336b336d3839617a6634633478633938777063646e78356830636e787936656d34377236646334616c756430757771783234663530","sender":"657264316566397878336b336d3839617a6634633478633938777063646e78356830636e787936656d34377236646334616c756430757771783234663530","receiverShard":0,"senderShard":0,"gasPrice":1000000000,"gasLimit":150000000,"gasUsed":139832352,"fee":"1802738520000000","data":"RVNEVE5GVFRyYW5zZmVyQDRjNGI0NjQxNTI0ZDJkMzM2NjM0NjYzOTYyQDAxNjUzNEA2ZjFlNmYwMWJjNzYyN2Y1YWVAMDAwMDAwMDAwMDAwMDAwMDA1MDBmMWM4ZjJmZGM1OGE2M2M2YjIwMWZjMmVkNjI5OTYyZDNkZmEzM2ZlN2NlYkA2MzZmNmQ3MDZmNzU2ZTY0NTI2NTc3NjE3MjY0NzM1MDcyNmY3ODc5QDAwMDAwMDAwMDAwMDAwMDAwNTAwNGY3OWVjNDRiYjEzMzcyYjVhYzlkOTk2ZDc0OTEyMGY0NzY0Mjc2MjdjZWI=","signature":"","timestamp":5040,"status":"success","searchOrder":0,"hasScResults":true,"scresults":{"7363724861736832":{"miniBlockHash":"fabb8b32bbafbb7a1f85b12a8892143fbce35ebe14ab67c455d69c9007082d5c","nonce":0,"gasLimit":148957500,"gasPrice":1000000000,"value":"\u003cnil\u003e","sender":"657264316566397878336b336d3839617a6634633478633938777063646e78356830636e787936656d34377236646334616c756430757771783234663530","receiver":"657264317171717171717171717171717170677137387930396c773933663375647673706c7368647632766b3935376c35766c37306e3473706c72616432","senderShard":0,"receiverShard":1,"data":"RVNEVE5GVFRyYW5zZmVyQDRjNGI0NjQxNTI0ZDJkMzM2NjM0NjYzOTYyQDAxNjUzNEA2ZjFlNmYwMWJjNzYyN2Y1YWVAMDgwMTEyMGEwMDZmMWU2ZjAxYmM3NjI3ZjVhZTIyNzYwOGI0Y2EwNTFhMjAwMDAwMDAwMDAwMDAwMDAwMDUwMGYxYzhmMmZkYzU4YTYzYzZiMjAxZmMyZWQ2Mjk5NjJkM2RmYTMzZmU3Y2ViMzIwMDNhNGMwMDAwMDAwZTRkNDU1ODQ2NDE1MjRkMmQ2MjM5MzM2NTM2MzAwMDAwMDAwMDAwMDE2YWI1MDAwMDAwMDk2ZjFlNmYwMWJjNzYyN2Y1YWUwMDAwMDAwYzRjNGI0ZDQ1NTgyZDY0MzE2MzM0NjE2MjAwMDAwMDAwMDAwMDczM2IwMDAwMDAwOTZlOTAxOGExZjBjYzlhOWFlZkA2MzZmNmQ3MDZmNzU2ZTY0NTI2NTc3NjE3MjY0NzM1MDcyNmY3ODc5QDAwMDAwMDAwMDAwMDAwMDAwNTAwNGY3OWVjNDRiYjEzMzcyYjVhYzlkOTk2ZDc0OTEyMGY0NzY0Mjc2MjdjZWI=","prevTxHash":"6e66745472616e7366657257697468534343616c6c","originalTxHash":"6e66745472616e7366657257697468534343616c6c","callType":"0","timestamp":5040},"73637257697468526566756e64":{"nonce":0,"gasLimit":0,"gasPrice":0,"value":"101676480000000","sender":"65726431717171717171717171717171717067713537737a77756432717579737563726c71326539376e74647973646c377634656a7a33716e336e6a7134","receiver":"657264316566397878336b336d3839617a6634633478633938777063646e78356830636e787936656d34377236646334616c756430757771783234663530","senderShard":0,"receiverShard":0,"data":"QDZmNmJAMDE3ZDE1QDAwMDAwMDBlNGQ0NTU4NDY0MTUyNGQyZDYyMzkzMzY1MzYzMDAwMDAwMDAwMDAwMTdkMTUwMDAwMDAwOTcwNDUxNzNjYzk3NTU0YjY1ZEAwMTc4YWY=","prevTxHash":"66363339636237613032333131393165303465633139646362313335396264393361303366653864633461323861383064303038333563356431633938386638","originalTxHash":"6e66745472616e7366657257697468534343616c6c","callType":"0","timestamp":5040}}}`
)

func TestNFTTransferCrossShardWithSCCall(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("nftTransferWithSCCall")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	scrHash1 := []byte("scrHash2")
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{txHash},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash1},
			},
		},
	}

	scr1 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		GasLimit:       148957500,
		SndAddr:        []byte("erd1ef9xx3k3m89azf4c4xc98wpcdnx5h0cnxy6em47r6dc4alud0uwqx24f50"),
		RcvAddr:        []byte("erd1qqqqqqqqqqqqqpgq78y09lw93f3udvsplshdv2vk957l5vl70n4splrad2"),
		Data:           []byte("ESDTNFTTransfer@4c4b4641524d2d336634663962@016534@6f1e6f01bc7627f5ae@0801120a006f1e6f01bc7627f5ae227608b4ca051a2000000000000000000500f1c8f2fdc58a63c6b201fc2ed629962d3dfa33fe7ceb32003a4c0000000e4d45584641524d2d6239336536300000000000016ab5000000096f1e6f01bc7627f5ae0000000c4c4b4d45582d643163346162000000000000733b000000096e9018a1f0cc9a9aef@636f6d706f756e645265776172647350726f7879@000000000000000005004f79ec44bb13372b5ac9d996d749120f476427627ceb"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): &transaction.Transaction{
				Nonce:    79,
				SndAddr:  []byte("erd1ef9xx3k3m89azf4c4xc98wpcdnx5h0cnxy6em47r6dc4alud0uwqx24f50"),
				RcvAddr:  []byte("erd1ef9xx3k3m89azf4c4xc98wpcdnx5h0cnxy6em47r6dc4alud0uwqx24f50"),
				GasLimit: 150000000,
				GasPrice: 1000000000,
				Data:     []byte("ESDTNFTTransfer@4c4b4641524d2d336634663962@016534@6f1e6f01bc7627f5ae@00000000000000000500f1c8f2fdc58a63c6b201fc2ed629962d3dfa33fe7ceb@636f6d706f756e645265776172647350726f7879@000000000000000005004f79ec44bb13372b5ac9d996d749120f476427627ceb"),
				Value:    big.NewInt(0),
			},
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): scr1,
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(expectedTxNFTTransferSCCallSource), genericResponse.Docs[0].Source)

	// EXECUTE transfer on the destination shard
	bodyDstShard := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash1},
			},
		},
	}
	scrWithRefund := []byte("scrWithRefund")
	refundValueBig, _ := big.NewInt(0).SetString("101676480000000", 10)
	poolDstShard := &indexer.Pool{
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): scr1,
			string(scrWithRefund): &smartContractResult.SmartContractResult{
				SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"),
				RcvAddr:        []byte("erd1ef9xx3k3m89azf4c4xc98wpcdnx5h0cnxy6em47r6dc4alud0uwqx24f50"),
				PrevTxHash:     []byte("f639cb7a0231191e04ec19dcb1359bd93a03fe8dc4a28a80d00835c5d1c988f8"),
				OriginalTxHash: txHash,
				Value:          refundValueBig,
				Data:           []byte("@6f6b@017d15@0000000e4d45584641524d2d6239336536300000000000017d15000000097045173cc97554b65d@0178af"),
			},
		},
	}

	err = esProc.SaveTransactions(bodyDstShard, header, poolDstShard)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(expectedTxNFTTransferSCCallAfterRefund), genericResponse.Docs[0].Source)
}

// TODO check also indexes that are altered
func TestNFTTransferCrossShard(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("nftTransfer")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}
	scrHash1 := []byte("scrHash1")
	scrHash2 := []byte("scrHash2")
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{txHash},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash2},
			},
		},
	}

	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		SndAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
		RcvAddr:        []byte("erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"),
		Data:           []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	refundValueBig, _ := big.NewInt(0).SetString("40365000000000", 10)
	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): &transaction.Transaction{
				Nonce:    79,
				SndAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				RcvAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				GasLimit: 5000000,
				GasPrice: 1000000000,
				Data:     []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@00000000000000000500a7a02771aa07090e607f02b25f4d6d241bff32b990a2"),
				Value:    big.NewInt(0),
			},
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): &smartContractResult.SmartContractResult{
				Nonce:          80,
				Value:          refundValueBig,
				GasPrice:       1000000000,
				SndAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				Data:           []byte("@6f6b"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
			},
			string(scrHash2): scr2,
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(expectedTxNFTTransfer), genericResponse.Docs[0].Source)

	// EXECUTE transfer on the destination shard
	bodyDstShard := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash2},
			},
		},
	}
	scr3WithErrHash := []byte("scrWithError")
	poolDstShard := &indexer.Pool{
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash2): scr2,
			string(scr3WithErrHash): &smartContractResult.SmartContractResult{
				SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				PrevTxHash:     []byte("1546eb9970a6dc1710b6528274e75d5095c1349706f4ff70f52a1f58e1156316"),
				OriginalTxHash: txHash,
				Data:           []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179@75736572206572726f72"),
			},
		},
	}

	err = esProc.SaveTransactions(bodyDstShard, header, poolDstShard)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(expectedTxNFTTransferFailOnDestination), genericResponse.Docs[0].Source)
}

func TestNFTTransferCrossShardImportDBScenarioFirstIndexDestinationAfterSource(t *testing.T) {
	setLogLevelDebug()

	esClient, err := createESClient(esURL)
	require.Nil(t, err)

	accounts := &mock.AccountsStub{}
	feeComputer := &mock.EconomicsHandlerMock{}
	shardCoordinator := &mock.ShardCoordinatorMock{}

	esProc, err := CreateElasticProcessor(esClient, accounts, shardCoordinator, feeComputer)
	require.Nil(t, err)

	txHash := []byte("nftTransferCross")
	scrHash1 := []byte("scrHashCross1")
	scrHash2 := []byte("scrHashCross2")
	header := &dataBlock.Header{
		Round:     50,
		TimeStamp: 5040,
	}

	scr2 := &smartContractResult.SmartContractResult{
		Nonce:          0,
		GasPrice:       1000000000,
		SndAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
		RcvAddr:        []byte("erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"),
		Data:           []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179"),
		PrevTxHash:     txHash,
		OriginalTxHash: txHash,
	}

	// EXECUTE transfer on the destination shard
	bodyDstShard := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash2},
			},
		},
	}
	scr3WithErrHash := []byte("scrWithError")
	poolDstShard := &indexer.Pool{
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash2): scr2,
			string(scr3WithErrHash): &smartContractResult.SmartContractResult{
				SndAddr:        []byte("erd1qqqqqqqqqqqqqpgq57szwud2quysucrlq2e97ntdysdl7v4ejz3qn3njq4"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				PrevTxHash:     []byte("1546eb9970a6dc1710b6528274e75d5095c1349706f4ff70f52a1f58e1156316"),
				OriginalTxHash: txHash,
				Data:           []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@08011202000122e50108011204434f4f4c1a20e0f3ecf555f63f2d101241dfc98b4614aff9284edd50b46a1c6e36b83558744d20c4132a2e516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a324368747470733a2f2f697066732e696f2f697066732f516d5a7961565631786a7866446255575a503178655a7676544d3156686f61346f594752444d706d4a727a52435a3a41746167733a436f6f6c3b6d657461646174613a516d5869417850396e535948515954546143357358717a4d32645856334142516145355241725932777a4e686179@75736572206572726f72"),
			},
		},
	}

	ids := []string{hex.EncodeToString(txHash)}
	genericResponse := &GenericResponse{}
	err = esProc.SaveTransactions(bodyDstShard, header, poolDstShard)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(txWithStatusOnly), genericResponse.Docs[0].Source)

	// execute on source

	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			{
				Type:            dataBlock.TxBlock,
				SenderShardID:   0,
				ReceiverShardID: 0,
				TxHashes:        [][]byte{txHash},
			},
			{
				Type:            dataBlock.SmartContractResultBlock,
				SenderShardID:   0,
				ReceiverShardID: 1,
				TxHashes:        [][]byte{scrHash2},
			},
		},
	}

	refundValueBig, _ := big.NewInt(0).SetString("40365000000000", 10)
	pool := &indexer.Pool{
		Txs: map[string]coreData.TransactionHandler{
			string(txHash): &transaction.Transaction{
				Nonce:    79,
				SndAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				RcvAddr:  []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				GasLimit: 5000000,
				GasPrice: 1000000000,
				Data:     []byte("ESDTNFTTransfer@434f4c45435449452d323663313838@01@01@00000000000000000500a7a02771aa07090e607f02b25f4d6d241bff32b990a2"),
				Value:    big.NewInt(0),
			},
		},
		Scrs: map[string]coreData.TransactionHandler{
			string(scrHash1): &smartContractResult.SmartContractResult{
				Nonce:          80,
				Value:          refundValueBig,
				GasPrice:       1000000000,
				SndAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				RcvAddr:        []byte("erd1ure7ea247clj6yqjg80unz6xzjhlj2zwm4gtg6sudcmtsd2cw3xs74hasv"),
				Data:           []byte("@6f6b"),
				PrevTxHash:     txHash,
				OriginalTxHash: txHash,
			},
			string(scrHash2): scr2,
		},
	}
	err = esProc.SaveTransactions(body, header, pool)
	require.Nil(t, err)

	err = esClient.DoMultiGet(ids, indexerdata.TransactionsIndex, true, genericResponse)
	require.Nil(t, err)

	compareTxs(t, []byte(completeTxWithStatus), genericResponse.Docs[0].Source)
}