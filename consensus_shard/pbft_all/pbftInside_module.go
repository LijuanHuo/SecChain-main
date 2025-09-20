// addtional module for new consensus
package pbft_all

import (
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	"blockEmulator/supervisor/committee"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	//"blockEmulator/utils"
	//"math/big"
	"blockEmulator/commitment"
	"github.com/alinush/go-mcl"
	"bytes"

	//////
	"os"
	"io/ioutil"
    "encoding/hex"
)

// simple implementation of pbftHandleModule interface ...
// only for block request and use transaction relay
type RawRelayPbftExtraHandleMod struct {
	pbftNode *PbftConsensusNode
	// pointer to pbft data
	committee *committee.RelayCommitteeModule 
}

// propose request with different types
func (rphm *RawRelayPbftExtraHandleMod) HandleinPropose1() (bool, *message.Request) {
	// new blocks
	block := rphm.pbftNode.CurChain.GenerateBlock()
	/////////////////////////////////
	if block == nil {
		rphm.pbftNode.pl.Plog.Printf("Failed to decode block from message content.\n")
		//return false
	}
	rphm.pbftNode.pl.Plog.Printf("Block Body: %+v\n", block.Body)//输出  Block Body: []
	////////////////////////////

	////////////////
	txExcuted := make([]*core.Transaction, 0)
	for _, tx := range block.Body {
		txExcuted = append(txExcuted, tx)
	}

	var aggLoPf commitment.LocalProof 
    var aggGloPf  commitment.GlobalProof

	if len(txExcuted) == 0 {
        rphm.pbftNode.pl.Plog.Printf("No transactions to process.\n")
    }else{
		aggLoPf, aggGloPf = AggregateProofs(txExcuted)  
	}
	
	// Serialize aggregated proofs
	serializedAggLoPf := aggLoPf.Local_proof_content.Serialize()
	serializedAggGloPf := aggGloPf.Global_proof_content.Serialize()

	// Convert byte slices to []byte for JSON marshaling
	aggLoPfBytes := make([]byte, len(serializedAggLoPf))
	copy(aggLoPfBytes, serializedAggLoPf)
	aggGloPfBytes := make([]byte, len(serializedAggGloPf))
	copy(aggGloPfBytes, serializedAggGloPf)
	////////////////

	r := &message.Request{
		RequestType: message.BlockRequest,
		ReqTime:     time.Now(),
	}
	r.Msg.Content = block.Encode()
	////////////
	r.Msg.AggLoPf = aggLoPfBytes
	r.Msg.AggGloPf = aggGloPfBytes
	/////////////

	return true, r
}
// propose request with different types
func (rphm *RawRelayPbftExtraHandleMod) HandleinPropose() (bool, *message.Request) {
	// new blocks
	block := rphm.pbftNode.CurChain.GenerateBlock()
	r := &message.Request{
		RequestType: message.BlockRequest,
		ReqTime:     time.Now(),
	}
	r.Msg.Content = block.Encode()

	return true, r
}

// the DIY operation in preprepare
func (rphm *RawRelayPbftExtraHandleMod) HandleinPrePrepare(ppmsg *message.PrePrepare) bool {
	if rphm.pbftNode.CurChain.IsValidBlock(core.DecodeB(ppmsg.RequestMsg.Msg.Content)) != nil {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : not a valid block\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		return false
	}
	rphm.pbftNode.pl.Plog.Printf("S%dN%d : the pre-prepare message is correct, putting it into the RequestPool. \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
	rphm.pbftNode.requestPool[string(ppmsg.Digest)] = ppmsg.RequestMsg
	// merge to be a prepare message
	return true
}

// the operation in prepare, and in pbft + tx relaying, this function does not need to do any.
func (rphm *RawRelayPbftExtraHandleMod) HandleinPrepare(pmsg *message.Prepare) bool {
	fmt.Println("No operations are performed in Extra handle mod")
	return true
}

//Once cross-shard transactions have been processed into intra-shard transactions using the `dealTxByPrepaidAccount` function, there is no longer any need to send relay transactions.
func (rphm *RawRelayPbftExtraHandleMod) HandleinCommit(cmsg *message.Commit) bool {
	r := rphm.pbftNode.requestPool[string(cmsg.Digest)]
	// requestType ...
	block := core.DecodeB(r.Msg.Content)
	// /////////////////////////////////
	// if block == nil {
	// 	rphm.pbftNode.pl.Plog.Printf("Failed to decode block from message content.\n")
	// 	return false
	// }
	// rphm.pbftNode.pl.Plog.Printf("Block Body: %+v\n", block.Body)//output  Block Body: []
	// ////////////////////////////
	rphm.pbftNode.pl.Plog.Printf("S%dN%d : adding the block %d...now height = %d \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, block.Header.Number, rphm.pbftNode.CurChain.CurrentBlock.Header.Number)
	rphm.pbftNode.CurChain.AddBlock(block)
	rphm.pbftNode.pl.Plog.Printf("S%dN%d : added the block %d... \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, block.Header.Number)
	rphm.pbftNode.CurChain.PrintBlockChain()

	// now try to relay txs to other shards (for main nodes)
	if rphm.pbftNode.NodeID == rphm.pbftNode.view {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : main node is trying to process transactions at height = %d \n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID, block.Header.Number)

		txExcuted := make([]*core.Transaction, 0)
		for _, tx := range block.Body {
			txExcuted = append(txExcuted, tx)
		}

		// Send executed transactions to the listener
		bim := message.BlockInfoMsg{
			BlockBodyLength: len(block.Body),
			ExcutedTxs:      txExcuted,
			Epoch:           0,
			SenderShardID:   rphm.pbftNode.ShardID,
			ProposeTime:     r.ReqTime,
			CommitTime:      time.Now(),
		}
		bByte, err := json.Marshal(bim)
		if err != nil {
			log.Panic()
		}
		msg_send := message.MergeMessage(message.CBlockInfo, bByte)
		go networks.TcpDial(msg_send, rphm.pbftNode.ip_nodeTable[params.DeciderShard][0])
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : sended excuted txs\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
		rphm.pbftNode.CurChain.Txpool.GetLocked()
		rphm.pbftNode.writeCSVline([]string{strconv.Itoa(len(rphm.pbftNode.CurChain.Txpool.TxQueue)), strconv.Itoa(len(txExcuted))})
		rphm.pbftNode.CurChain.Txpool.GetUnlocked()
	}
	return true
}

func (rphm *RawRelayPbftExtraHandleMod) HandleReqestforOldSeq(*message.RequestOldMessage) bool {
	fmt.Println("No operations are performed in Extra handle mod")
	return true
}

// the operation for sequential requests
func (rphm *RawRelayPbftExtraHandleMod) HandleforSequentialRequest(som *message.SendOldMessage) bool {
	if int(som.SeqEndHeight-som.SeqStartHeight+1) != len(som.OldRequest) {
		rphm.pbftNode.pl.Plog.Printf("S%dN%d : the SendOldMessage message is not enough\n", rphm.pbftNode.ShardID, rphm.pbftNode.NodeID)
	} else { // add the block into the node pbft blockchain
		for height := som.SeqStartHeight; height <= som.SeqEndHeight; height++ {
			r := som.OldRequest[height-som.SeqStartHeight]
			if r.RequestType == message.BlockRequest {
				b := core.DecodeB(r.Msg.Content)
				rphm.pbftNode.CurChain.AddBlock(b)
			}
		}
		rphm.pbftNode.sequenceID = som.SeqEndHeight + 1
		rphm.pbftNode.CurChain.PrintBlockChain()
	}
	return true
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

	rows := 3// 
    cols := len(txExcuted) 
	indexMatrix := make([][]int, rows)
	for i := range indexMatrix {
		indexMatrix[i] = make([]int, cols)
	   
		for j, tx := range txExcuted {
			//if j < cols { 
				indexMatrix[i][j] = int(tx.Index) 
			//}
		}
	}

 

f, err := os.OpenFile("agg.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
if err != nil {
	log.Fatalf("error opening log file: %v", err)
}

log.SetOutput(f)
log.Printf("len(txExcuted) %d\n", len(txExcuted))

	//globalSize := params.ShardNum
	globalSize := len(indexMatrix)
	localSize :=  len(indexMatrix)
	N :=len(params.ProofConfig.Matrix)

	fmt.Println("globalSize:", globalSize)
    fmt.Println("localSize:", localSize)
    fmt.Println("N:", N)
    // fmt.Printf("params.ProofConfig.Matrix: %+v\n", params.ProofConfig.Matrix)
    // fmt.Printf("params.ProofConfig.Values: %+v\n", params.ProofConfig.Values)
    // fmt.Printf("params.ProofConfig.LProofs: %+v\n", params.ProofConfig.LProofs)
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

	///////////////////////////////

	return params.ProofConfig.AggLProof, params.ProofConfig.AggGProof
}

// Convert [][]byte to [][]string
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