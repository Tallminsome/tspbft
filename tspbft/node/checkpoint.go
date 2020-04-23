package node

import "log"

func (n *Node) CheckPointRecvThread() {
	for {
		select {
		case msg := <-n.CheckPointRecv:
			n.buffer.BufferCheckPointMsg(msg, msg.I)
			if n.buffer.TrueOfCheckPointMsg(msg.D, n.comcfg.FaultNum) {
				n.buffer.Show()
				log.Printf("CheckPoint is successfully produced, then clean the buffer")
				n.buffer.CleanBuffer(string(n.group),msg) //clean seq<checkpoint
				n.sequence.CheckPoint()
				log.Printf("[CheckPoint] reset the water mark (%d) - (%d)", n.sequence.waterL, n.sequence.waterH)
				n.buffer.Show()
			}
		}
	}
}
