package utils

import (
	"blockEmulator/params"
	"log"
	"strconv"
	"fmt"
	"sync"
)

// the default method
func Addr2Shard(addr Address) int {
	last16_addr := addr[len(addr)-8:]
	num, err := strconv.ParseUint(last16_addr, 16, 64)
	if err != nil {
		log.Panic(err)
	}
	return int(num) % params.ShardNum
}
//////////////////////////////////////////////////////
// ShardManager 管理所有分片及其内部地址索引
type ShardManager struct {
	mu           sync.RWMutex
	shardToIndex map[int]map[Address]int // 每个分片到地址索引的映射
	nextIDPerShard map[int]int          // 每个分片下一个可用的编号
}

// NewShardManager 创建一个新的 ShardManager 实例
func NewShardManager(shardNum int) *ShardManager {
	sm := &ShardManager{
		shardToIndex:   make(map[int]map[Address]int),
		nextIDPerShard: make(map[int]int),
	}
	for i := 0; i < shardNum; i++ {
		sm.shardToIndex[i] = make(map[Address]int)
		sm.nextIDPerShard[i] = 0
	}
	return sm
}

// AddAddress 添加地址并返回其在分片内的索引
func (sm *ShardManager) AddAddress(addr Address) (shard int, index int) {
	shard = Addr2Shard(addr)

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 如果地址已经存在，则直接返回已有的索引
	if idx, exists := sm.shardToIndex[shard][addr]; exists {
		return shard, idx
	}

	// 否则，为新地址分配新的索引
	sm.shardToIndex[shard][addr] = sm.nextIDPerShard[shard]
	sm.nextIDPerShard[shard]++

	return shard, sm.shardToIndex[shard][addr]
}

// GetIndex 快速查找地址在分片内的索引
func (sm *ShardManager) GetIndex(addr Address) (shard int, index int, exists bool) {
	shard = Addr2Shard(addr)

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	index, exists = sm.shardToIndex[shard][addr]
	return shard, index, exists
}

func CreateIndexMatrix(sm *ShardManager) [][]int {
	// 第一次遍历，确定每个分片的最大索引数
	maxIndices := make([]int, params.ShardNum)
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for shard := range sm.shardToIndex {
		maxIndices[shard] = len(sm.shardToIndex[shard])
	}

	// 初始化索引矩阵
	indexMatrix := make([][]int, params.ShardNum)
	for i := range indexMatrix {
		indexMatrix[i] = make([]int, maxIndices[i])
		for j := range indexMatrix[i] {
			indexMatrix[i][j] = -1 // 使用 -1 表示未填充的位置
		}
	}

	// 第二次遍历，填充索引矩阵
	for shard, addresses := range sm.shardToIndex {
		for _, index := range addresses {
			indexMatrix[shard][index] = int(index) // 将索引值填入矩阵
		}
	}

	return indexMatrix
}

func Test() {
	params.ShardNum = 5 // 设置分片数量

	// 创建一个 ShardManager 实例
	shardManager := NewShardManager(params.ShardNum)

	// 假设这是几个地址
	addresses := []Address{
		"0xf967aa80d80d6f22df627219c5113a118b57d0ef",
		"0x9008d19f58aabd9ed0d60971565aa8510560ab41",
		"0x03bada9ff1cf0d0264664b43977ed08feee32584",
		"0x27239549dd40e1d60f5b80b0c4196923745b1fd2",
	}

	// 为每个地址添加并打印其分片和索引
	for _, addr := range addresses {
		shard, index := shardManager.AddAddress(addr)
		fmt.Printf("Address %s is in shard %d with index %d\n", addr, shard, index)
	}

	// 测试快速查找功能
	testAddr := "0x9008d19f58aabd9ed0d60971565aa8510560ab41"
	shard, index, exists := shardManager.GetIndex(testAddr)
	if exists {
		fmt.Printf("Address %s has shard %d and index %d\n", testAddr, shard, index)
	} else {
		fmt.Printf("Address %s does not exist\n", testAddr)
	}
}