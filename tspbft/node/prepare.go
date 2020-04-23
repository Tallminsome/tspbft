package node

import (
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/server"
	"log"
)

func (n *Node) PrepareRecvAndSendCommitThread() {
	// if its not sub-primary
	if !n.WhetherSubPrimary() {
		return
	}
	for {
		select {
		case msg := <-n.PrepareRecv:
			if !n.CheckPrepareMsg(msg) {
				log.Printf("[wrong checkprepare]")
				continue
			}
			log.Printf("[OK checkprepare]")
			//buffer
			n.buffer.BufferPrepareMsg(msg)
			log.Printf("OK Bufferpreparemsg")
			//vertify 2f
			if n.buffer.TrueOfPrepareMsg(msg.D, n.comcfg.FaultNum) {
				log.Printf("[Prepare] prepare msg(%d) vote success and to send commit", msg.N)
				content, msg, err := message.NewCommitMsg(n.id, msg)
				if err != nil {
					continue
				}
				// buffer commit msg
				n.buffer.BufferCommitMsg(msg)
				//send vote result to PrimaryNode
				n.SendComtoPrimary(content, server.CommitEntry)
			}
			//if n.buffer.WhetherToExecute(msg.D, n.comcfg.FaultNum, string(n.group), msg.N) {
			//	n.ReadyToExecute(msg.D)
			//}
		}
	}
}

func (n *Node) CheckPrepareMsg(msg *message.Prepare) bool {
	//if n.view != msg.V {
	//	return false
	//}
	//check digest
	//pre := <- n.PrePrepareRecv
	//d, err := message.Digest(pre.M)
	//if err != nil {
	//	return false
	//}
	//if d!= msg.D {
	//	return false
	//}
	if !n.sequence.CheckValid(msg.N) {
		return false
	}
	return true
}