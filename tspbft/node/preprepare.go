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
	if n.executeNum.Get() >= n.comcfg.ExecuteMaxNum {
		return
	}

	if n.buffer.SizeofRequestQueue() < 1 {
		return
	}
	//batch request to discard network traffic
	batch := n.buffer.BatchRequest()
	if len(batch) < 1 {
		return
	}

	seq := n.sequence.Add()
	n.executeNum.Inc()
	content, msg, digest, err := message.NewPreprepareMsg(seq, batch)
	if err != nil {
		log.Printf("Generate pre-prepare message error : %s",err)
		return
	}
	log.Printf("Generate sequence(%d) for message(%s) with request batch size(%d)",seq,digest[0:9],len(batch))
	// buffer the pre-prepare msg
	n.buffer.BufferPrePrepareMsg(string(n.group),msg)
	//sent pre-prepare to primary in order to execute
	n.SendPrepreparetoPrimary(content,server.PrePrepareEntry)
	//broadcast
	n.Broadcast(content,server.PrePrepareEntry)
}

//receive pre-prepare and send prepare thread
func (n *Node) PrePrepareRecvAndSendPrepareThread() {
	// if its sub-primary or primary
	if n.WhetherSubPrimary() {
		return
	}
	for {
		select {
		case msg := <-n.PrePrepareRecv:
			if !n.CheckPrePrepareMsg(string(n.group),msg){
				continue
			}
		    //buffer pre-prepare message
			n.buffer.BufferPrePrepareMsg(string(n.group),msg)
			if n.WhetherPrimary() {
				if n.buffer.WhetherPrimaryToExecute(msg.D, string(n.group), msg.N) {
					log.Printf("Primary is Ready To Execute")
					n.ReadyToExecute(msg.D)
				}
				continue
			}
			//generate prepare message
			content, prepare, err := message.NewPrepareMsg(n.id, msg)
			log.Printf("[Pre-Prepare] recv pre-prepare(%d) and send the prepare", msg.N)
			if err != nil {
				log.Printf("Wrong in generating NewPrepareMsg")
				continue
			}
			// buffer the prepare msg, vertify 2f backup
			n.buffer.BufferPrepareMsg(prepare)
			n.buffer.PrepareLocker.Lock()
			n.buffer.PrepareState[msg.D] = true  //fa song wan zhi hou she zhi
			n.buffer.PrepareLocker.Unlock()
			// broadcast prepare message
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
