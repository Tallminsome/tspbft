package node

import (
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/server"
	"log"
	"time"
)

//send pre_prepare messager by request notify or timer
func (n *Node) PrePrepareSendThread() {
	log.Printf("Start sending preprepare message")
	duration := time.Second
	timer := time.After(duration)
	for {
		select {
		case <-n.PrePrepareSendNotify:
			n.PrePrepareSendHandle()
		case <-timer:
			timer = nil
			n.PrePrepareSendHandle()
			timer = time.After(duration)
		}
	}
}

func (n *Node) PrePrepareSendHandle() {
	//buffer is empty or execute op num max
	n.executenum.locker.Lock()
	defer n.executenum.locker.Unlock()
	if n.executenum.Get() >= n.comcfg.ExecuteMaxNum {
		return
	}
	if n.buffer.SizeofRequestQueue() < 1 {
		return
	}
	// if it's not sub-primary
	if !n.WhetherSubPrimary() {
		return
	}
	//batch request to discard network traffic
	batch := n.buffer.BatchRequest()
	if len(batch) < 1 {
		return
	}

	seq := n.sequence.Add()
	n.executenum.Inc()
	content, msg, digest, err := message.NewPreprepareMsg(seq, batch)
	if err != nil {
		log.Printf("Generate pre-prepare message error : %s",err)
		return
	}
	log.Printf("Generate sequence(%d) for message(%s) with request batch size(%d)",seq,digest[0:9],len(batch))
	// buffer the pre-prepare msg
	n.buffer.BufferPrePrepareMsg(string(n.group),msg)
	//broadcast
	n.Broadcast(content,server.PrePrepareEntry)
}

//receive pre-prepare and send prepare thread
func (n *Node) PrePrepareRecvAndSendPrepareThread() {
	// if its sub-primary or primary
	if n.WhetherSubPrimary() || n.WhetherPrimary() {
		return
	}
	for {
		select {
		case msg := <-n.PrePrepareRecv:
			if !n.CheckPrePrepareMsg(string(n.group),msg){
				continue
				log.Printf("CheckPreprepareMsg Error")
			}
		    //buffer pre-prepare message
			n.buffer.BufferPrePrepareMsg(string(n.group),msg)
			//generate prepare message
			content, prepare, err := message.NewPrepareMsg(n.id, msg)
			log.Printf("[Pre-Prepare] recv pre-prepare(%d) and send the prepare", msg.N)
			if err != nil {
				continue
				log.Printf("Wrong in generating NewPrepareMsg")
			}
			// buffer the prepare msg, vertify 2f backup
			n.buffer.BufferPrepareMsg(prepare)
			log.Printf("OK with BufferPrepareMsg")
			n.buffer.PrepareState[msg.D] = true  //fa song wan zhi hou she zhi
			// broadcast prepare message
			n.SendPreparetoSubPrimary(content,server.PrepareEntry)
			// when commit and prepare vote success but not recv pre-prepare
			if n.buffer.WhetherToExecute(msg.D, string(n.group), msg.N) {
				log.Printf("ToExecute")
				n.ReadyToExecute(msg.D)
			}
		}
	}
}

func (n *Node) CheckPrePrepareMsg(grp string,msg *message.PrePrepare) bool {
	//missing check signature
	var grps = ""
	switch grp {
	case "B":
		grps = "AB"
	case "C":
		grps = "AC"
	case "D":
		grps = "AD"
	}
	if n.buffer.WhetherExistPrePrepareMsg(grps, msg.N) {
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
