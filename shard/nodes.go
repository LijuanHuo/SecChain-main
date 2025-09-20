// definition of node and shard

package shard

import (
	//"crypto/rand"
	//"blockEmulator/core"
	//"blockEmulator/utils"
	//"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	//"math/big"
	//"math/rand"
	"os"
	"strings"
	//"time"
	//"sync"
	//"github.com/ethereum/go-ethereum/common"// For Ethereum address handling
)

type Node struct {
	NodeID  uint64
	ShardID uint64
	IPaddr  string
}

func (n *Node) PrintNode() {
	v := []interface{}{
		n.NodeID,
		n.ShardID,
		n.IPaddr,
	}
	fmt.Printf("%v\n", v)
}

//---------------

type Shard struct {
	ShardID   uint64
	Nodes     []*Node
	Leader    *Node
	NodeCount uint64
	//shardLock sync.Mutex // For concurrent access safety
}

func AssignNodesFromCSV(filePath string) ([]Node, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	uniqueAddresses := make(map[string]bool)
	var nodes []Node
	nodeID := uint64(1)

	reader := csv.NewReader(file)
	for {
		record, err := reader.Read()
		if err != nil {
			if err == csv.ErrFieldCount {
				continue // Skip empty lines or lines with less than expected fields
			} else if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

		// Extracting sender and receiver addresses, trimming and converting to Ethereum address format
		senderAddr := strings.TrimSpace(record[3])
		receiverAddr := strings.TrimSpace(record[4])

		// Ensuring uniqueness and adding to nodes list
		if !uniqueAddresses[senderAddr] {
			uniqueAddresses[senderAddr] = true
			nodes = append(nodes, Node{NodeID: nodeID, IPaddr: senderAddr})
			nodeID++
		}
		if !uniqueAddresses[receiverAddr] {
			uniqueAddresses[receiverAddr] = true
			nodes = append(nodes, Node{NodeID: nodeID, IPaddr: receiverAddr})
			nodeID++
		}
	}

	return nodes, nil
}

func PrintNodeInfo() {
	nodes, err := AssignNodesFromCSV("./TestTx_100.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, node := range nodes {
		fmt.Printf("NodeID: %d, IPaddr: %s\n", node.NodeID, node.IPaddr)
	}
}

func Sharding(shardCount int, nodes []Node) ([]Shard, error) {
	if shardCount <= 0 {
		return nil, errors.New("shard count must be greater than 0")
	}

	// Initialize the shard
	shards := make([]Shard, shardCount)
	for i := range shards {
		shards[i] = Shard{
			ShardID:   uint64(i),
			Nodes:     []*Node{},
			Leader:    nil,
			NodeCount: 0,
		}
	}

	// Allocate nodes to their respective shards
	for _, node := range nodes {
		shardID := int(node.NodeID % uint64(shardCount))
		shards[shardID].Nodes = append(shards[shardID].Nodes, &node)
		shards[shardID].NodeCount++
	}

	return shards, nil
}

func ExecuteSharding() {

	nodes, err := AssignNodesFromCSV("./TestTx_100.csv")
	if err != nil {
		fmt.Println("Error assigning nodes from CSV:", err)
		return
	}

	shardCount := 4 
	shards, err := Sharding(shardCount, nodes)
	if err != nil {
		fmt.Println("Error during sharding:", err)
		return
	}

	for _, shard := range shards {
		fmt.Printf("Shard ID: %d, Node Count: %d\n", shard.ShardID, shard.NodeCount)
		
	}
}


