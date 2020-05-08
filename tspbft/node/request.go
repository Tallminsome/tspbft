package node

import (
	"log"
)
var num = 0
func (n *Node) RequestRecvThread() {
	for {
		msg := <-n.RequestRecv
		num = num + 1
		// check if it's subprimary
		if !n.WhetherSubPrimary() {
			// if it's lastreply just send it to client directely
			log.Printf("This is %d", n.id)
			if n.lastreply.Equal(msg) {
				// TODO just Reply
			}
		}
		log.Printf("[Req]This is req:%d,time stamp:%s",num,msg.T)
		n.buffer.AppendToRequestQueue(msg)
		n.PrePrepareSendNotify <- true
	}
}
