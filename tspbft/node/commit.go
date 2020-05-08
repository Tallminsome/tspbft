package node

import (
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/server"
	"log"
)

func (n *Node) CommitRecvAndVertifyThread() {
	// if its not Primary
	if !n.WhetherPrimary() {
		return
	}
	for {
		select {
		case msg := <-n.CommitRecv:
			if !n.CheckCommitMsg(msg) {
				continue
			}
			n.buffer.BufferCommitMsg(msg)
			log.Printf("[Commit] node(%d) vote to the msg(%d)", msg.I, msg.N)
            if n.buffer.TrueOfCommitMsg(msg.D,n.comcfg.FaultNum) {
				_, msg, err := message.NewVerifyMsg(n.id,msg)
				if err != nil {
					continue
				}
				if n.buffer.WhetherPrimaryToExecute(msg.D, string(n.group), msg.N) {
					log.Printf("Primary is Ready To Execute")
					n.ReadyToExecute(msg.D)
				}
				n.BroadtoSubPrimary(msg, server.VerifyEntry)
				log.Printf("vertify 2f+1 success")
			}
		}
	}
}

func (n *Node) CheckCommitMsg(msg *message.Commit) bool {
	//if n.view != msg.V {
	//	return false
	//}
	if !n.sequence.CheckValid(msg.N) {
		return false
	}
	return true
}
