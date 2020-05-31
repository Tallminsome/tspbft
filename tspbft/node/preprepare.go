package node

import (
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/server"
	"log"
	"time"
)

//send pre_prepare messager by request notify or timer
func (n *Node) PrePrepareSendThread() {
	// if it's not sub-primary
	if !n.WhetherSubPrimary() {
		return
	}
	duration := time.Second
	timer := time.After(duration)
	for {
		select {
		case <-n.PrePrepareSendNotify:
			log.Printf("success recv req")
			n.PrePrepareSendHandle()
		case <-timer:
			timer = nil
			n.PrePrepareSendHandle()
			timer = time.After(duration)
		}
	}
}

func (n *Node) PrePrepareSendHandle() {
	n.executeNum.Lock()
	defer n.executeNum.UnLock()
	//处理的打包消息最多不超过一条
	if n.executeNum.Get() >= n.comcfg.ExecuteMaxNum {
		return
	}
	//把队列中的打包消息放在batch里减少网络拥堵
	if n.buffer.SizeofRequestQueue() < 1 {
		return
	}
	batch := n.buffer.BatchRequest()
	if len(batch) < 1 {
		return
	}
	//序列号以及执行数+1
	seq := n.sequence.Add()
	n.executeNum.Inc()
	////生成pre-prepare消息
	content, msg, digest, err := message.NewPreprepareMsg(seq, batch)
	if err != nil {
		log.Printf("Generate pre-prepare message error : %s",err)
		return
	}
	log.Printf("Generate sequence(%d) for message(%s) with request batch size(%d)",seq,digest[0:9],len(batch))
	//缓存pre-prepare消息
	n.buffer.BufferPrePrepareMsg(string(n.group),msg)
	//将pre-prepare消息发送给根节点缓存以便执行
	n.SendPrepreparetoPrimary(content,server.PrePrepareEntry)
	//把pre-prepare消息在通道内进行广播
	n.Broadcast(content,server.PrePrepareEntry)
}

//接收pre-prepare并且发送prepare消息
func (n *Node) PrePrepareRecvAndSendPrepareThread() {
	//如果节点为主节点,则退出该进程
	if n.WhetherSubPrimary() {
		return
	}
	for {
		select {
		case msg := <-n.PrePrepareRecv:
			if !n.CheckPrePrepareMsg(string(n.group),msg){
				continue
			}
			//缓存pre-prepare消息
			n.buffer.BufferPrePrepareMsg(string(n.group),msg)
			if n.WhetherPrimary() {
				//如果是根节点,判断是否能够执行,以防出现commit消息验证通过而pre-prepare消息还没收到的情况
				if n.buffer.WhetherPrimaryToExecute(msg.D, string(n.group), msg.N) {
					log.Printf("Primary is Ready To Execute")
					n.ReadyToExecute(msg.D)
				}
				continue
			}
			//生成prepare消息
			content, prepare, err := message.NewPrepareMsg(n.id, msg)
			log.Printf("[Pre-Prepare] recv pre-prepare(%d) and send the prepare", msg.N)
			if err != nil {
				log.Printf("Wrong in generating NewPrepareMsg")
				continue
			}
			//缓存prepare消息
			n.buffer.BufferPrepareMsg(prepare)
			n.buffer.PrepareLocker.Lock()
			n.buffer.PrepareState[msg.D] = true
			n.buffer.PrepareLocker.Unlock()
			//向主节点发送prepare消息
			n.SendPreparetoSubPrimary(content,server.PrepareEntry)
			if n.buffer.WhetherToExecute(msg.D, string(n.group), msg.N) {
				log.Printf("ToExecute")
				n.ReadyToExecute(msg.D)
			}
		}
	}
}

func (n *Node) CheckPrePrepareMsg(grp string,msg *message.PrePrepare) bool {
	// check the same group and n exist diffrent digest
	if n.buffer.WhetherExistPrePrepareMsg(grp, msg.N) {
		return false
	}
	//check digest
	d, err := message.Digest(msg.M)
	if err != nil {
		return false
	}
	if d != msg.D {
		return false
	}
	// check sequence number
	if !n.sequence.CheckValid(msg.N) {
		return false
	}
	return true
}
