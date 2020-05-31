package node

import (
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/server"
	"log"
)

func (n *Node) PrepareRecvAndSendCommitThread() {
	//判断是否为下层通道主节点，不是则退出该进程
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
			//缓存prepare消息
			n.buffer.BufferPrepareMsg(msg)
			//验证是否收到门限t个prepare消息
			if n.buffer.TrueOfPrepareMsg(msg.D, n.comcfg.FaultNum) {
				log.Printf("[Prepare] prepare msg(%d) vote success and to send commit", msg.N)
				content, msg, err := message.NewCommitMsg(n.id, msg)
				if err != nil {
					continue
				}
				//缓存commit消息
				n.buffer.BufferCommitMsg(msg)
				//发送commit消息到根节点
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