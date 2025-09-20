package committee

import (
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/params"
	//"blockEmulator/shard"
	"blockEmulator/supervisor/signal"
	"blockEmulator/supervisor/supervisor_log"
	"blockEmulator/utils"
	"encoding/csv"
	"encoding/json"
	//"fmt"
	"io"
	"log"
	"math/big"
	"os"
	"time"
)

type RelayCommitteeModule struct {
	csvPath      string
	dataTotalNum int
	nowDataNum   int
	batchDataNum int
	IpNodeTable  map[uint64]map[uint64]string
	sl           *supervisor_log.SupervisorLog
	Ss           *signal.StopSignal // to control the stop message sending
	AccountState *core.AccountState     
	ProcessedTxs []*core.Transaction 
	ShardManager *utils.ShardManager 

}

func NewRelayCommitteeModule(Ip_nodeTable map[uint64]map[uint64]string, Ss *signal.StopSignal, slog *supervisor_log.SupervisorLog, csvFilePath string, dataNum, batchNum int) *RelayCommitteeModule {
	return &RelayCommitteeModule{
		csvPath:      csvFilePath,
		dataTotalNum: dataNum,
		batchDataNum: batchNum,
		nowDataNum:   0,
		IpNodeTable:  Ip_nodeTable,
		Ss:           Ss,
		sl:           slog,
		AccountState: &core.AccountState{
			Prepaid: make(map[int][]*core.PrepaidAccount),
		},
		ShardManager: utils.NewShardManager(params.ShardNum),//NewAdd
	}
}

// transfrom, data to transaction
// check whether it is a legal txs meesage. if so, read txs and put it into the txlist
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

func (rthm *RelayCommitteeModule) HandleOtherMessage([]byte) {}

func (rthm *RelayCommitteeModule) txSending(txlist []*core.Transaction) {
	// the txs will be sent
	sendToShard := make(map[uint64][]*core.Transaction)

	for idx := 0; idx <= len(txlist); idx++ {
		if idx > 0 && (idx%params.InjectSpeed == 0 || idx == len(txlist)) {
			// send to shard
			for sid := uint64(0); sid < uint64(params.ShardNum); sid++ {
				it := message.InjectTxs{
					Txs:       sendToShard[sid],
					ToShardID: sid,
				}
				itByte, err := json.Marshal(it)
				if err != nil {
					log.Panic(err)
				}
				send_msg := message.MergeMessage(message.CInject, itByte)
				go networks.TcpDial(send_msg, rthm.IpNodeTable[sid][0])

			}
			sendToShard = make(map[uint64][]*core.Transaction)
			time.Sleep(time.Second)
		}
		if idx == len(txlist) {
			break
		}
		tx := txlist[idx]
		/////////////
		_, index := rthm.ShardManager.AddAddress(tx.Sender)//NewSdd
		tx.Index = int64(index)
		////////////
		sendersid := uint64(utils.Addr2Shard(tx.Sender))
		sendToShard[sendersid] = append(sendToShard[sendersid], tx)
	}
}

// read transactions, the Number of the transactions is - batchDataNum
func (rthm *RelayCommitteeModule) MsgSendingControl() {
	txfile, err := os.Open(rthm.csvPath)
	if err != nil {
		log.Panic(err)
	}
	defer txfile.Close()
	reader := csv.NewReader(txfile)
	txlist := make([]*core.Transaction, 0) // save the txs in this epoch (round)

	for {
		data, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Panic(err)
		}
		if tx, ok := data2tx(data, uint64(rthm.nowDataNum)); ok {
			txlist = append(txlist, tx)
			rthm.nowDataNum++
		}

		// re-shard condition, enough edges
		if len(txlist) == int(rthm.batchDataNum) || rthm.nowDataNum == rthm.dataTotalNum {
			//New Add
			itx := rthm.DealTxByPrepaidAccount(txlist)
			rthm.txSending(itx)
			// Store the processed transactions in the shared data structure
			rthm.ProcessedTxs = append(rthm.ProcessedTxs, itx...)// New Add
			// reset the variants about tx sending
			txlist = make([]*core.Transaction, 0)
			rthm.Ss.StopGap_Reset()
		}

		if rthm.nowDataNum == rthm.dataTotalNum {
			break
		}
	}
}

// no operation here
func (rthm *RelayCommitteeModule) HandleBlockInfo(b *message.BlockInfoMsg) {
	rthm.sl.Slog.Printf("received from shard %d in epoch %d.\n", b.SenderShardID, b.Epoch)
}


func (rthm *RelayCommitteeModule) DealTxByPrepaidAccount(txs []*core.Transaction) (itxs []*core.Transaction) {
	itxs = make([]*core.Transaction, 0)

	for _, tx := range txs {
		sendersid := uint64(utils.Addr2Shard(tx.Sender))
		receiversid := uint64(utils.Addr2Shard(tx.Recipient))

		if sendersid != receiversid {
			
			prepaidAccount := &core.PrepaidAccount{
				AcAddress:   tx.Sender,
				Nonce:       0, 
				Balance:     big.NewInt(10000), //Assume the initial balance is 1000
				ShardID:     int(receiversid),
			}

			// Add the prepaid account to the mapping of the destination shard.
			rthm.AccountState.Prepaid[prepaidAccount.ShardID] = append(rthm.AccountState.Prepaid[prepaidAccount.ShardID], prepaidAccount)

			// Create a new transaction from the prepaid account to the recipient.
			newTx := core.NewTransaction(
				prepaidAccount.AcAddress,      // The sender is a prepaid account address.
				tx.Recipient,                  // The recipient remains unchanged.
				tx.Value,                      
				tx.Nonce,                      
			)

			// Add the new transaction to the list.
			itxs = append(itxs, newTx)
		} else {
			// If the sender and receiver are in the same shard, directly add the transaction to the list.
			itxs = append(itxs, tx)
		}
	}

	return itxs
}
