package main

import (
	"blockEmulator/build"
	"blockEmulator/core"
	"blockEmulator/params"
	//"blockEmulator/shard"
	"blockEmulator/utils"
	//"strconv"

	"github.com/spf13/pflag"

	//"blockEmulator/consensus_shard/pbft_all/pbft_log"
	"blockEmulator/commitment"
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	//"math/rand"
	//"time"

	"github.com/alinush/go-mcl"

	//"blockEmulator/test"
	//"sync"
	"os"
	// "encoding/gob"
	// "io"
	"encoding/json"
	"io/ioutil"
	//"strings"

	//"blockEmulator/mysockets"
	//"crypto/sha256"
	//"net"
)

var (
	shardNum   int
	nodeNum    int
	shardID    int
	nodeID     int
	modID      int
	isClient   bool
	isGen      bool
	relay1Txs  []*core.Transaction
	ipAddr     string = ""
	Idx2Addr   []string
	Addr2Idx   map[string]int
	Addr2Trade map[string][]int
	userB      []*big.Int
)

func convertBigIntMatrixToMclFrMatrix(Matrix [][]*big.Int) []mcl.Fr {
	var result []mcl.Fr

	for _, row := range Matrix {
		for _, val := range row {
			// 创建一个新的mcl.Fr实例并从*big.Int设置值
			var fr mcl.Fr
			fr = *mcl.BigIntToFr(val)
			// if err != nil {
			// 	return nil, err
			// }
			result = append(result, fr)
		}
	}

	return result
}

type SerializablePublicParams struct {
	M                        int
	N                        int
	Pp_generators_alpha      [][]byte
	Pp_generators_alpha_beta [][][]byte
	Vp_generators_alpha      [][]byte
	Vp_generators_beta       [][]byte
	Vp_gt_elt                []byte
	G1                       []byte
	G2                       []byte
}

// 将 [][]byte 转换为 [][]string
func bytesMatrixToStringsMatrix(byteSlices [][]byte) [][]string {
	var strSlices [][]string
	for _, byteSlice := range byteSlices {
		var strSlice []string
		for len(byteSlice) > 0 {
			lengthToEncode := 20

			if len(byteSlice) < 20 {
				lengthToEncode = len(byteSlice)
			}

			// 编码前lengthToEncode字节为十六进制字符串
			hexStr := hex.EncodeToString(byteSlice[:lengthToEncode])

			// 添加前缀 "0x"
			hexStr = "0x" + hexStr
			// fmt.Printf("  hexStr: %s\n", hexStr)

			// 添加到 strSlice
			strSlice = append(strSlice, hexStr)

			// 移除已经处理的部分
			byteSlice = byteSlice[lengthToEncode:]
		}

		strSlices = append(strSlices, strSlice)
	}
	return strSlices
}

// 从文件中加载参数
func LoadParams() (params.PfConfig, error) {
	var config params.PfConfig

	// 读取文件中的 JSON 字节
	jsonData, err := ioutil.ReadFile("params.json")
	if err != nil {
		return config, err
	}

	// 反序列化 JSON 字节为参数结构体
	var serializableParams params.SerializePfConfig
	err = json.Unmarshal(jsonData, &serializableParams)
	if err != nil {
		return config, err
	}

	config.GlobalPublicParams = serializableParams.GlobalPublicParams
	config.Matrix = serializableParams.Matrix

	for _, value := range serializableParams.Values {
		var fr mcl.Fr
		err := fr.Deserialize(value)
		if err != nil {
			return config, err
		}
		config.Values = append(config.Values, fr)
	}

	config.AcAddress = bytesMatrixToStringsMatrix(serializableParams.AcAddress)

	for _, lcom := range serializableParams.LComs {
		var lc commitment.LocalCommitment
		err := lc.Local_commitment.Deserialize(lcom)
		if err != nil {
			return config, err
		}
		config.LComs = append(config.LComs, lc)
	}

	var gc commitment.GlobalCommitment
	err = gc.Global_commitment.Deserialize(serializableParams.GCom)
	if err != nil {
		return config, err
	}
	config.GCom = gc

	config.LProofs = make([]commitment.LocalProof, len(serializableParams.LProofs))
	config.GProof = make([]commitment.GlobalProof, len(serializableParams.GProof))

	for i, lproof := range serializableParams.LProofs {
		var lp commitment.LocalProof
		err := lp.Local_proof_content.Deserialize(lproof)
		if err != nil {
			return config, err
		}
		config.LProofs[i] = lp
	}

	for i, gproof := range serializableParams.GProof {
		var gp commitment.GlobalProof
		err := gp.Global_proof_content.Deserialize(gproof)
		if err != nil {
			return config, err
		}
		config.GProof[i] = gp
	}

	var aggLP commitment.LocalProof
	err = aggLP.Local_proof_content.Deserialize(serializableParams.AggLProof)
	if err != nil {
		return config, err
	}
	config.AggLProof = aggLP

	var aggGP commitment.GlobalProof
	err = aggGP.Global_proof_content.Deserialize(serializableParams.AggGProof)
	if err != nil {
		return config, err
	}
	config.AggGProof = aggGP

	return config, nil
}

// func mygo() string {
// 	var err error
// 	params.ProofConfig, err = LoadParams()
// 	if err != nil && !os.IsNotExist(err) {
// 		log.Fatal("Failed to load parameters:", err)
// 	}

// 	params.ProofConfig.Matrix, params.ProofConfig.AcAddress = shard.FetchAccountsFromShards()
// 	// maxAggSize := len(params.ProofConfig.Matrix)
// 	/*generate global commitment*/
// 	// N := maxAggSize
// 	params.ProofConfig.Values = convertBigIntMatrixToMclFrMatrix(params.ProofConfig.Matrix)

// 	// var relay1Txs []*core.Transaction
// 	relay1Txs, err = shard.FetchTxListFromCSV()
// 	if err != nil {
// 		log.Fatalf("Failed to fetch transactions: %v", err)
// 	}

// 	//init3List(relay1Txs)

// 	// fmt.Printf("11111relay1Txs \n", relay1Txs[0].Sender)
// 	indexMatrix, _ := FindTransactionAddressesAndBalances(relay1Txs[:10], core.ShardAccountStates)

// 	globalSize := params.ShardNum
// 	localSize := len(indexMatrix)
// 	//N :=len(indexMatrix)

// 	fmt.Println(globalSize)
// 	fmt.Println(localSize)

// 	 return fmt.Sprintf("%dus", globalSize)
// }


func main() {
	fmt.Printf("Hello\n")
	//commitment.TestAlg()

	//start := time.Now()
	//mygo()
	var err error
	params.ProofConfig, err = LoadParams()
	if err != nil && !os.IsNotExist(err) {
		log.Fatal("Failed to load parameters:", err)
	}
	//elapsed := time.Since(start)
	//fmt.Printf("mygo() took %s to completeHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHHH.\n", elapsed)

	//----------------------------------------------------------
	pflag.IntVarP(&shardNum, "shardNum", "S", 2, "indicate that how many shards are deployed")
	pflag.IntVarP(&nodeNum, "nodeNum", "N", 4, "indicate how many nodes of each shard are deployed")
	pflag.IntVarP(&shardID, "shardID", "s", 0, "id of the shard to which this node belongs, for example, 0")
	pflag.IntVarP(&nodeID, "nodeID", "n", 0, "id of this node, for example, 0")
	pflag.IntVarP(&modID, "modID", "m", 3, "choice Committee Method,for example, 0, [CLPA_Broker,CLPA,Broker,Relay] ")
	pflag.BoolVarP(&isClient, "client", "c", false, "whether this node is a client")
	pflag.BoolVarP(&isGen, "gen", "g", false, "generation bat")
	pflag.Parse()

	if isGen {
		build.GenerateBatFile(nodeNum, shardNum, modID)
		build.GenerateShellFile(nodeNum, shardNum, modID)
		return
	}

	if isClient {
		build.BuildSupervisor(uint64(nodeNum), uint64(shardNum), uint64(modID))
	} else {
		build.BuildNewPbftNode(uint64(nodeID), uint64(nodeNum), uint64(shardID), uint64(shardNum), uint64(modID))
	}
	// // //----------------------------------------------------------


}

func FindTransactionAddressesAndBalances(relay1Txs []*core.Transaction, shardAccountStates [][]*core.AccountState) ([][]int, [][]*big.Int) {
	// 初始化下标矩阵和余额矩阵
	indexMatrix := make([][]int, params.ShardNum)
	balanceMatrix := make([][]*big.Int, params.ShardNum)

	// 遍历所有分片
	for i := 0; i < params.ShardNum; i++ {
		// 为当前分片初始化下标和余额
		indexMatrix[i] = make([]int, 0)
		balanceMatrix[i] = make([]*big.Int, 0)

		// 遍历所有交易，查找每个交易的发送方在当前分片中的位置
		for _, tx := range relay1Txs {
			// 打印调试信息
			// fmt.Printf("Transaction Sender: %s\n", tx.Sender)
			// 查找Sender地址的索引和余额，同时获取分片ID
			shardID, senderIdx, senderBalance := FindAccountIndexAndBalance(tx.Sender, shardAccountStates)

			// 打印调试信息
			// fmt.Printf("Transaction Sender: %s, Shard ID: %d, Sender Index: %d, Sender Balance: %s\n", tx.Sender, shardID, senderIdx, senderBalance)

			// 只有当发送方位于当前分片时，才记录位置和余额
			if shardID == i && senderIdx != -1 {
				indexMatrix[i] = append(indexMatrix[i], senderIdx)
				balanceMatrix[i] = append(balanceMatrix[i], senderBalance)
			}
		}
	}

	return indexMatrix, balanceMatrix
}

func FindAccountIndexAndBalance(address utils.Address, shardAccountStates [][]*core.AccountState) (int, int, *big.Int) {
	for shardId, accounts := range shardAccountStates {
		//fmt.Printf("Checking Shard ID: %d\n", shardId)
		for idx, acc := range accounts {
			//    fmt.Printf("Checking acc.AcAddres: %s\n", acc.AcAddress)
			// 	fmt.Printf("Checking address: %s\n", address)
			// 只有当地址长度大于等于2时才去除 "0x" 前缀
			// if len(acc.AcAddress) >= 2 && strings.HasPrefix(acc.AcAddress, "0x") {
			//     acc.AcAddress = acc.AcAddress[2:]
			// }
			if address == acc.AcAddress {
				// fmt.Printf("Found match at Shard ID: %d, Index: %d\n", shardId, idx)
				return shardId, idx, acc.Balance
			}
		}
	}
	// fmt.Println("No match found")
	return -1, -1, (*big.Int)(nil) // 如果没找到，返回-1和nil
}

func sliceToString(slice []int) string {
	var b bytes.Buffer
	for i, v := range slice {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%d", v))
	}
	return b.String()
}

// 辅助函数，用于将big.Int切片转换为字符串
func sliceToBigIntString(slice []*big.Int) string {
	var b bytes.Buffer
	for i, v := range slice {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(v.String())
	}
	return b.String()
}
