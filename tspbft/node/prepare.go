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
			//buffer
			n.buffer.BufferPrepareMsg(msg)
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
		}
	}
}

func (n *Node) CheckPrepareMsg(msg *message.Prepare) bool {
	if !n.sequence.CheckValid(msg.N) {
		return false
	}
	return true
}