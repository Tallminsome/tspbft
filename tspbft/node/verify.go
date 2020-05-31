package node

import (
	"encoding/json"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/server"
	"log"
)

func (n *Node) VerifyRecvAndBroadThread() {
	//判断是否为下层通道主节点，不是则退出该进程
	if !n.WhetherSubPrimary() {
		return
	}
	for {
		select {
		case msg := <-n.VerifyRecv:
			if !n.CheckVerifyMsg(msg) {
				continue
			}
			n.buffer.BufferVerifyMsg(msg)
			log.Printf("[Verify] node(%s) received verify to the msg(%d) from primary", n.group, msg.N)
            n.buffer.VerifyLocker.Lock()
			n.buffer.VerifyState[msg.D] = true
			n.buffer.VerifyLocker.Unlock()
			if n.buffer.WhetherToExecute(msg.D, string(n.group), msg.N) {
			    log.Printf("Sub-Primary is Ready To Execute")
				n.executeNum.Dec()
			    n.ReadyToExecute(msg.D)
		    }
			//生成verified消息并向通道内的副本节点广播verified消息
			_, verifiedmsg, err := message.NewVerifiedMsg(n.id,msg)
			content, err := json.Marshal(verifiedmsg)
			if err != nil {
				log.Printf("error to marshal json")
				return
			}
			n.Broadcast(content, server.VerifiedEntry)
		}
	}
}

func (n *Node) ReplicaRecvVerifyThread() {
	//如果节点不是副本节点,则退出进程
	if n.WhetherPrimary() || n.WhetherSubPrimary() {
		return
	}
	for {
		select {
		case msg := <-n.VerifiedRecv:
			n.buffer.VerifyLocker.Lock()
			n.buffer.VerifyState[msg.D] = true
			n.buffer.VerifyLocker.Unlock()
			if n.buffer.WhetherToExecute(msg.D, string(n.group), msg.N) {
				log.Printf("ToExecute")
				n.ReadyToExecute(msg.D)
			}
		}
	}
}

func (n *Node) CheckVerifyMsg(msg *message.Verify) bool {
	if !n.sequence.CheckValid(msg.N) {
		return false
	}
	return true
}