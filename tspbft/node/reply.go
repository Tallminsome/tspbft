package node

import (
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	cb "github.com/hyperledger/fabric/protos/common"
	"log"
)

var test_reqeust_num uint64 = 0
func (n *Node) ExecuteAndReplyThread() {
	for {
		select {
		case <-n.ExecuteNotify:
			// execute batch
			batchs, lastSeq := n.buffer.BatchExecute(n.sequence.GetLastSequence())
			if len(batchs) == 0 {
				log.Printf("[Reply] lost sequence now(%d)", n.sequence.GetLastSequence())
				continue
			}
			n.sequence.SetLastSequence(lastSeq)
			//check point
			//if n.sequence.ReadyToCheckPoint() {
			//	log.Printf("[CheckPoint]Ready to Checkpoint")
			//	CheckSeq := n.sequence.GetCheckpoint()
			//	content, checkpoint := n.buffer.NewCheckPoint(CheckSeq, n.id)
			//
			//	n.buffer.BufferCheckPointMsg(checkpoint, n.id)
			//	log.Printf("[Reply] ready to create check point to sequence(%d) msg(%s)", CheckSeq, checkpoint.D[0:9])
			//	if n.WhetherPrimary(){
			//		continue
			//	}
			//	n.BroadcastCheckpointMsg(content, server.CheckPointEntry)
			//}
			// map the digest to request
			requestBatchs := make([]*message.BufferReq, 0)
			for i:=0; i<len(batchs); i++ {
				for j:=0; j<len(batchs[i].M); j++ {
					for k:=0; k<len(batchs[i].M[j].Requests); k++ {
						requestBatchs = append(requestBatchs, batchs[i].M[j].Requests[k])
					}
				}
			}
			test_reqeust_num = test_reqeust_num + uint64(len(requestBatchs))
			log.Printf("[Reply] set last sequence(%d) already execute request(%d)", lastSeq, test_reqeust_num)
			//pending state
			pending := make(map[string]bool)
			for _, r := range requestBatchs {
				log.Printf("[Execute]This req is:%d from client:%d,timestamp is:%s",r.N,r.C,r.T)
				Msg        := r.O.Envelope
				Channel    := r.O.ChannelID
				ConfigSeq  := r.O.ConfigSeq
				switch r.O.Type {
				case message.TYPECONFIG:
					var err error
					seq := n.Supports[Channel].Sequence()
					if ConfigSeq < seq {
						if Msg, _, err = n.Supports[Channel].ProcessConfigMsg(r.O.Envelope); err != nil {
							log.Println(err)
						}
					}
					batch := n.Supports[Channel].BlockCutter().Cut()
					if batch != nil {
						block := n.Supports[Channel].CreateNextBlock(batch)
						n.Supports[Channel].WriteBlock(block, nil)
					}
					pending[Channel] = false
					block := n.Supports[Channel].CreateNextBlock([]*cb.Envelope{Msg})
					n.Supports[Channel].WriteConfigBlock(block, nil)
				case message.TYPENORMAL:
					seq := n.Supports[Channel].Sequence()
					if ConfigSeq < seq {
						if _, err := n.Supports[Channel].ProcessNormalMsg(Msg); err != nil {
							log.Println(err)
						}
					}
					batchs, p := n.Supports[Channel].BlockCutter().Ordered(Msg)
					for _, batch := range batchs {
						block := n.Supports[Channel].CreateNextBlock(batch)
						n.Supports[Channel].WriteBlock(block, nil)
					}
					pending[Channel] = p
				}
			}
			for k,v := range pending  {
				if v {
					batch := n.Supports[k].BlockCutter().Cut()
					if batch != nil {
						block := n.Supports[k].CreateNextBlock(batch)
						n.Supports[k].WriteBlock(block, nil)
					}
				}
			}
		}
	}

}