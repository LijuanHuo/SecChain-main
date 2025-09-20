package main

import (
	//"blockEmulator/build"
	"blockEmulator/core"
	"blockEmulator/params"
	//"blockEmulator/shard"
	"blockEmulator/utils"
	//"strconv"

	//"github.com/spf13/pflag"

	//"blockEmulator/consensus_shard/pbft_all/pbft_log"
	"blockEmulator/commitment"
	"bytes"
	//"encoding/hex"
	"fmt"
	"log"
	"math/big"
	//"math/rand"
	"time"
    "encoding/csv"
    //"log"


	"github.com/alinush/go-mcl"

	//"blockEmulator/test"
	//"sync"
	"os"
	// "encoding/gob"
	 "io"
	//"encoding/json"
	//"io/ioutil"
	"strings"
    "encoding/json"
	"io/ioutil"
    "encoding/hex"

	//"blockEmulator/mysockets"
	//"crypto/sha256"
	//"net"
)
func main() {
	fmt.Printf("Hello\n")
	utils.Test()
    // txExcuted := make([]*core.Transaction, 0)
    // for _, tx := range block.Body {
    //     txExcuted = append(txExcuted, tx)
    // }
    // Fetch transactions from CSV file
    //txExcuted, err := FetchTxListFromCSV()
    // if err != nil {
    //     log.Panicf("Failed to fetch transactions from CSV: %v", err)
    // }
	dataLines := []string{
        ",,,0xf967aa80d80d6f22df627219c5113a118b57d0ef,0x9008d19f58aabd9ed0d60971565aa8510560ab41,,0,0,500000000000000000000000",
        ",,,0x03bada9ff1cf0d0264664b43977ed08feee32584,0x9008d19f58aabd9ed0d60971565aa8510560ab41,,0,0,18152905298971505661",
        ",,,0x9008d19f58aabd9ed0d60971565aa8510560ab41,0x27239549dd40e1d60f5b80b0c4196923745b1fd2,,0,0,498542955950186730094592",
        ",,,0x05aaa0053fa5c28e8c558d4c648cc129bea45018,0x27239549dd40e1d60f5b80b0c4196923745b1fd2,,0,0,7334760148650696200",
        ",,,0x27239549dd40e1d60f5b80b0c4196923745b1fd2,0x05aaa0053fa5c28e8c558d4c648cc129bea45018,,0,0,166180985316728910031530",
    }
	txExcuted := make([]*core.Transaction, 0)
	for i, line := range dataLines {
        // 分割每一行数据
        fields := strings.Split(line, ",")
        
        // 跳过空行或字段数量不符合预期的行
        if len(fields) < 9 {
            log.Printf("Skipping invalid line %d: %s", i+1, line)
            continue
        }

        // 调用 data2tx 创建交易对象
        tx, ok := data2tx(fields, uint64(i))
        if ok {
            txExcuted = append(txExcuted, tx)
        } else {
            log.Printf("Failed to create transaction from line %d: %s", i+1, line)
        }
    }




    /////////////////////////////////////////////////////////////
    //Aggregate proofs from transactions in the current block
    aggLoPf, aggGloPf := AggregateProofs(txExcuted)
    
    // Serialize aggregated proofs
    serializedAggLoPf := aggLoPf.Local_proof_content.Serialize()
    serializedAggGloPf := aggGloPf.Global_proof_content.Serialize()

    // Convert byte slices to []byte for JSON marshaling
    aggLoPfBytes := make([]byte, len(serializedAggLoPf))
    copy(aggLoPfBytes, serializedAggLoPf)
    aggGloPfBytes := make([]byte, len(serializedAggGloPf))
    copy(aggGloPfBytes, serializedAggGloPf)
	
}

func FetchTxListFromCSV() ([]*core.Transaction, error) {
	txlist := []*core.Transaction{}

	txfile, err := os.Open(params.FileInput)
	if err != nil {
		log.Panic(err)
	}
	defer txfile.Close()
	reader := csv.NewReader(txfile)
	nowDataNum := 0
	for {
		data, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panic(err)
		}
		if nowDataNum == params.TotalDataSize {
			break
		}

		if tx, ok := data2tx(data, uint64(nowDataNum)); ok {
			//fmt.Printf("tx.Sender: %s\n", tx.Sender)
			txlist = append(txlist, tx) 
			nowDataNum++
		}
	}
	return txlist, nil
}

func data2tx(data []string, nonce uint64) (*core.Transaction, bool) {
	if data[6] == "0" && data[7] == "0" && len(data[3]) > 16 && len(data[4]) > 16 && data[3] != data[4] {
		val, ok := new(big.Int).SetString(data[8], 10)
		if !ok {
			log.Panic("new int failed\n")
		}
		tx := core.NewTransaction(data[3][2:], data[4][2:], val, nonce)
		return tx, true
	}
	return &core.Transaction{}, false
}

func AggregateProofs(txExcuted []*core.Transaction) (AggLProof commitment.LocalProof, AggGProof commitment.GlobalProof){
    start := time.Now()
	var err error
	params.ProofConfig, err = LoadParams()
	if err != nil && !os.IsNotExist(err) {
		log.Fatal("Failed to load parameters:", err)
	}
	elapsed := time.Since(start)
	fmt.Printf("LoadParams() took %s to complete.\n", elapsed)


	shardManager := utils.NewShardManager(params.ShardNum)
	for _, tx := range txExcuted {
		shard, index := shardManager.AddAddress(tx.Sender)
		fmt.Printf("Address %s is in shard %d with index %d\n", tx.Sender, shard, index)
	}

	for _, tx := range txExcuted {
		shard, index, exists := shardManager.GetIndex(tx.Sender)
		if exists {
			fmt.Printf("Address %s has shard %d and index %d\n", tx.Sender, shard, index)
		} else {
			fmt.Printf("Address %s does not exist\n", tx.Sender)
		}
	}
	indexMatrix :=  utils.CreateIndexMatrix(shardManager)

	for i, row := range indexMatrix {
		fmt.Printf("Shard %d: %v\n", i, row)
	}

	//globalSize := params.ShardNum
	globalSize := len(indexMatrix)
	localSize :=  len(indexMatrix)
	N :=len(params.ProofConfig.Matrix)

	fmt.Println("globalSize:", globalSize)
    fmt.Println("localSize:", localSize)
    fmt.Println("N:", N)
    if globalSize <= 0 || localSize <= 0 || N <= 0 {
        log.Panic("Invalid sizes for globalSize, localSize, or N")
    }

    // Ensure other configuration parameters are not nil and have valid lengths
    if len(params.ProofConfig.Values) == 0 || len(params.ProofConfig.LProofs) == 0 {
        log.Panic("Some configuration parameters are not initialized properly")
    }
	/*aggregate proofs*/
	//step 1: prepare the aggregated proofs
	in_set := make([][]int, globalSize)
	in_sub_value := make([][]mcl.Fr, globalSize)
	in_sub_proof := make([][]commitment.LocalProof, globalSize)
	out_set := make([]int, globalSize)
	out_sub_value := make([]mcl.G1, globalSize)
	out_sub_proof := make([]commitment.GlobalProof, globalSize)
	for i := 0; i < globalSize; i++ {
		 line_in_set := make([]int, len(indexMatrix[i]))
		 line_in_sub_value := make([]mcl.Fr, len(indexMatrix[i]))
		 line_in_sub_proof := make([]commitment.LocalProof, len(indexMatrix[i]))

		 for j, idx := range indexMatrix[i] {
			 line_in_set[j] = idx
			line_in_sub_value[j] = params.ProofConfig.Values[i*params.ProofConfig.GlobalPublicParams.N+idx]   
			 line_in_sub_proof[j] = params.ProofConfig.LProofs[i*params.ProofConfig.GlobalPublicParams.N+idx]
		 }
		in_set[i] = line_in_set
	    in_sub_value[i] = line_in_sub_value
		in_sub_proof[i] = line_in_sub_proof

		out_set[i] = i
		out_sub_value[i] = params.ProofConfig.LComs[i].Local_commitment
		out_sub_proof[i] = params.ProofConfig.GProof[i]
	}

	//step2: aggregate proofs
	startTime := time.Now()
	params.ProofConfig.AggLProof = commitment.AggregateLocalProof(out_sub_value, in_sub_proof, in_set, in_sub_value, N)
	//params.ProofConfig.AggLProof = commitment.AggregateLocalProof(out_sub_value, in_sub_proof, squareIndexMatrix, frbalanceMatrix, n)
	duration := time.Since(startTime)
	fmt.Println("Aggregate", globalSize*localSize, "local proofs takes", duration.Microseconds(), "us")

	startTime = time.Now()
	//aggGloPf := AggregateGlobalProof(gCom.Global_commitment, out_sub_proof, out_set, out_sub_value, N)
	params.ProofConfig.AggGProof = commitment.AggregateGlobalProof(params.ProofConfig.GCom.Global_commitment, out_sub_proof, out_set, out_sub_value, N)
	duration = time.Since(startTime)
	fmt.Println("Aggregate", globalSize, "global proofs takes", duration.Microseconds(), "us")

	///////////////////////////
	/*batch verification*/
	startTime = time.Now()
	// res := params.ProofConfig.AggLProof.BatchVerifyLocalProof(params.GlobalPublicParams, out_sub_value, squareIndexMatrix, frbalanceMatrix)
	res := params.ProofConfig.AggLProof.BatchVerifyLocalProof(params.ProofConfig.GlobalPublicParams, out_sub_value, in_set, in_sub_value)
	//res := AAGGLocal.BatchVerifyLocalProof(params.ProofConfig.GlobalPublicParams, out_sub_value, in_set, in_sub_value)
	fmt.Println("The verification of aggLoPf is {}", res)
	if !res {
		fmt.Println("The verification of aggLoPf is {}", res)
	}
	duration = time.Since(startTime)
	fmt.Println("Batch verify the aggregated local proof takes", duration.Microseconds(), "us")

	startTime = time.Now()
	res = params.ProofConfig.AggGProof.BatchVerifyGlobalProof(params.ProofConfig.GlobalPublicParams, params.ProofConfig.GCom.Global_commitment, out_set, out_sub_value)
	fmt.Println("The verification of aggGloPf is {}", res)
	if !res {
		fmt.Println("The verification of aggGloPf is {}", res)
	}
	duration = time.Since(startTime)
	fmt.Println("Batch verify the aggregated global proof takes", duration.Microseconds(), "us")


	return params.ProofConfig.AggLProof, params.ProofConfig.AggGProof
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

func bytesMatrixToStringsMatrix(byteSlices [][]byte) [][]string {
	var strSlices [][]string
	for _, byteSlice := range byteSlices {
		var strSlice []string
		for len(byteSlice) > 0 {
			lengthToEncode := 20

			if len(byteSlice) < 20 {
				lengthToEncode = len(byteSlice)
			}

			hexStr := hex.EncodeToString(byteSlice[:lengthToEncode])

			hexStr = "0x" + hexStr
			// fmt.Printf("  hexStr: %s\n", hexStr)

			strSlice = append(strSlice, hexStr)

			byteSlice = byteSlice[lengthToEncode:]
		}

		strSlices = append(strSlices, strSlice)
	}
	return strSlices
}

func LoadParams() (params.PfConfig, error) {
	var config params.PfConfig

	jsonData, err := ioutil.ReadFile("params.json")
	if err != nil {
		return config, err
	}

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
