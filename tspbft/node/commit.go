package node

import (
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/server"
	"log"
)

func (n *Node) CommitRecvAndVertifyThread() {
	//如果不是根节点
	if !n.WhetherPrimary() {
		return
	}
	for {
		select {
		case msg := <-n.CommitRecv:
			if !n.CheckCommitMsg(msg) {
				continue
			}
			//缓存commit消息
			n.buffer.BufferCommitMsg(msg)
			log.Printf("[Commit] node(%d) vote to the msg(%d)", msg.I, msg.N)
            if n.buffer.TrueOfCommitMsg(msg.D,n.comcfg.FaultNum) {
				//commit消息验证成功,生成verify消息
            	_, msg, err := message.NewVerifyMsg(n.id,msg)
				if err != nil {
					continue
				}
				//根节点能否将交易写进区块
				if n.buffer.WhetherPrimaryToExecute(msg.D, string(n.group), msg.N) {
					log.Printf("Primary is Ready To Execute")
					n.ReadyToExecute(msg.D)
				}
				//根节点广播verify消息到下层通道主节点
				n.BroadtoSubPrimary(msg, server.VerifyEntry)
				log.Printf("vertify 2f+1 success")
			}
		}
	}
}

func (n *Node) CheckCommitMsg(msg *message.Commit) bool {
	if !n.sequence.CheckValid(msg.N) {
		return false
	}
	return true
}
