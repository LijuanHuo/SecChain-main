package params

import (
	"blockEmulator/commitment"
	"math/big"

	//"blockEmulator/core"
	"github.com/alinush/go-mcl"
	//"blockEmulator/utils"
)

var (
	Block_Interval      = 5000  // generate new block interval
	MaxBlockSize_global = 5000  // the block contains the maximum number of transactions
	InjectSpeed         = 5000  // the transaction inject speed
	TotalDataSize       = 100000 // the total number of txs
	BatchSize           = 5000  // supervisor read a batch of txs then send them, it should be larger than inject speed
	// TotalDataSize       = 1600000 // The total number of txs to be injected
	// BatchSize           = 16000  // The supervisor read a batch of txs then send them. The size of a batch is 'BatchSize'
	//BrokerNum           = 10
	NodesInShard        = 4
	ShardNum            = 8
	DataWrite_path      = "./result/"          // measurement data result output path
	LogWrite_path       = "./log"              // log output path
	SupervisorAddr      = "127.0.0.1:18800"    //supervisor ip address
	FileInput           = `./TestTx_100000.csv` //the raw BlockTransaction data path
	//FileInput           = `./TestTx_1M.csv` //the raw BlockTransaction data path
)

type PfConfig struct {
	GlobalPublicParams commitment.PublicParams
	Matrix             [][]*big.Int //balance
	Values             []mcl.Fr
	AcAddress          [][]string //type util.Address = string
	LComs              []commitment.LocalCommitment
	GCom               commitment.GlobalCommitment
	LProofs            []commitment.LocalProof
	GProof             []commitment.GlobalProof
	AggLProof          commitment.LocalProof
	AggGProof          commitment.GlobalProof
}

var ProofConfig PfConfig

type SerializePfConfig struct {
	GlobalPublicParams commitment.PublicParams
	Matrix             [][]*big.Int
	Values             [][]byte 
	AcAddress          [][]byte
	LComs              [][]byte 
	GCom               []byte  
	LProofs            [][]byte
	GProof             [][]byte
	AggLProof          []byte
	AggGProof          []byte
}
