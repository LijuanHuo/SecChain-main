package test

import (
	// "blockEmulator/build"
	// "github.com/spf13/pflag"
	"blockEmulator/params"
	"blockEmulator/shard"
	"blockEmulator/core"
	"blockEmulator/utils"
	//"blockEmulator/consensus_shard/pbft_all/pbft_log"
	"fmt"
	"log"
	"bytes"
	//"encoding/hex"
	"time"
	"math/big"
	"blockEmulator/commitment"
	"github.com/alinush/go-mcl"
	//"blockEmulator/test"
	//"sync"
	"os"
	// "encoding/gob"
	// "io"
	"encoding/json"
	"io/ioutil"
)

func test (){
	var relay1Txs []*core.Transaction
	relay1Txs, err := shard.FetchTxListFromCSV()
	if err != nil {
		log.Fatalf("Failed to fetch transactions: %v", err)
	}
	indexMatrix, balanceMatrix := FindTransactionAddressesAndBalances(relay1Txs[:10], core.ShardAccountStates)

fmt.Println("balanceMatrix = [")
for i, balances := range balanceMatrix {
	fmt.Printf("  [%s],  // shard %d index \n", sliceToBigIntString(balances), i)
}
fmt.Println("]")

	globalSize := params.ShardNum
	localSize :=  len(indexMatrix)
	N :=len(params.ProofConfig.Matrix)

	fmt.Println(globalSize)
	fmt.Println(localSize)
	/*aggregate proofs*/
	//step 1: prepare the aggregated proofs
	in_set := make([][]int, globalSize)
	in_sub_value := make([][]mcl.Fr, globalSize)
	in_sub_proof := make([][]commitment.LocalProof, globalSize)
	out_set := make([]int, globalSize)
	out_sub_value := make([]mcl.G1, globalSize)
	out_sub_proof := make([]commitment.GlobalProof, globalSize)
	for i := 0; i < globalSize; i++ {
		 // Note: The array size is no longer fixed at localSize, but is dynamically created based on the lengths of set[i] and frbalanceMatrix[i]
		 line_in_set := make([]int, len(indexMatrix[i]))
		 line_in_sub_value := make([]mcl.Fr, len(indexMatrix[i]))
		 line_in_sub_proof := make([]commitment.LocalProof, len(indexMatrix[i]))
 
		 // Traverse the length of set[i] instead of a fixed localSize.
		 for j, idx := range indexMatrix[i] {
			 // Directly use the values in set[i] and value_sub_vector[i]
			 line_in_set[j] = idx
			// line_in_sub_value[j] = frbalanceMatrix[i][j] 
			line_in_sub_value[j] = params.ProofConfig.Values[i*N+idx]   
			line_in_sub_proof[j] = params.ProofConfig.LProofs[i*N+idx]
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
	params.ProofConfig.AggLProof = commitment. AggregateLocalProof(out_sub_value, in_sub_proof, in_set, in_sub_value, N)
	//params.ProofConfig.AggLProof = commitment.AggregateLocalProof(out_sub_value, in_sub_proof, squareIndexMatrix, frbalanceMatrix, n)
	duration := time.Since(startTime)
	fmt.Println("Aggregate", globalSize*localSize, "local proofs takes", duration.Microseconds(), "us")

	startTime = time.Now()
	//aggGloPf := AggregateGlobalProof(gCom.Global_commitment, out_sub_proof, out_set, out_sub_value, N)
	params.ProofConfig.AggGProof = commitment.AggregateGlobalProof(params.ProofConfig.GCom.Global_commitment, out_sub_proof, out_set, out_sub_value, N)
	duration = time.Since(startTime)
	fmt.Println("Aggregate", globalSize, "global proofs takes", duration.Microseconds(), "us")

	/*batch verification*/
	startTime = time.Now()
	// res := params.ProofConfig.AggLProof.BatchVerifyLocalProof(params.ProofConfig.GlobalPublicParams, out_sub_value, squareIndexMatrix, frbalanceMatrix)
	res := params.ProofConfig.AggLProof.BatchVerifyLocalProof(params.ProofConfig.GlobalPublicParams, out_sub_value, in_set, in_sub_value)
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
}



func FindTransactionAddressesAndBalances(relay1Txs []*core.Transaction, shardAccountStates [][]*core.AccountState) ([][]int, [][]*big.Int) {
    // Initialize the index matrix and balance matrix
    indexMatrix := make([][]int, params.ShardNum)
    balanceMatrix := make([][]*big.Int, params.ShardNum)

    // Traverse all shards
    for i := 0; i < params.ShardNum; i++ {
        // Initialize the index and balance for the current shard
        indexMatrix[i] = make([]int, 0)
        balanceMatrix[i] = make([]*big.Int, 0)

        // Iterate through all transactions to find the location of each transaction's sender within the current shard.
        for _, tx := range relay1Txs {
            // Look up the index and balance of the sender address while obtaining the shard ID.
            shardID, senderIdx, senderBalance := FindAccountIndexAndBalance(tx.Sender, shardAccountStates)

            // Only when the sender is located in the current shard will the position and balance be recorded.
            if shardID == i && senderIdx != -1 {
                indexMatrix[i] = append(indexMatrix[i], senderIdx)
                balanceMatrix[i] = append(balanceMatrix[i], senderBalance)
            }
        }
    }

    return indexMatrix, balanceMatrix
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
// Auxiliary function for converting a big.Int slice to a string
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



func FindAccountIndexAndBalance(address utils.Address, shardAccountStates [][]*core.AccountState) (int, int, *big.Int) {
	for shardId, accounts := range shardAccountStates {
		for idx, acc := range accounts {
			if address == acc.AcAddress {
				return shardId, idx, acc.Balance
			}
		}
	}
	return -1, -1, (*big.Int)(nil) // If not found, return -1 and nil.
}
func ExtractG1FromLocalCommitments(lcs []commitment.LocalCommitment) []mcl.G1 {
    g1s := make([]mcl.G1, len(lcs))
    for i, lc := range lcs {
        g1s[i] = lc.Local_commitment // Let Global_commitment be a field of LocalCommitment.
    }
    return g1s
}

func ExtractG1FromLocalProofs(lps []commitment.LocalProof) []mcl.G1 {
    g1s := make([]mcl.G1, len(lps))
    for i, lp := range lps {
        g1s[i] = lp.Local_proof_content // Assume Local_proof_content is a field of LocalProof.
    }
    return g1s
}