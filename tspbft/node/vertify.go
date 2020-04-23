package node

import (
	"encoding/json"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/server"
	"log"
)

func (n *Node) CommitRecvAndBroadThread() {
	// if its not sub-Primary
	if !n.WhetherSubPrimary() {
		return
	}
	for {
		select {
		case msg := <-n.CommitRecv:
			if !n.CheckCommitMsg(msg) {
				continue
			}
			n.buffer.BufferCommitMsg(msg)
			log.Printf("[Commit] node(%s) received commit to the msg(%d) from primary", n.group, msg.N)
            n.buffer.CommitState[msg.D] = true
			//broad cast the commit msg to replica
			content, err := json.Marshal(msg)
			if err != nil {
				log.Printf("error to marshal json")
				return
			}

			log.Printf("[Commit] Send to peers")
			n.Broadcast(content, server.CommitEntry)



			//if n.buffer.WhetherToExecute(msg.D, n.comcfg.FaultNum, msg.V, msg.N) {
			//	n.ReadyToExecute(msg.D)
			//}
		}
	}

}
