package node

import (
	"log"
)

func (n *Node) RequestRecvThread() {
	log.Printf("Start receiving the request")
	for {
		msg := <-n.RequestRecv
		// check if it's subprimary
		if !n.WhetherSubPrimary() {
			// if it's lastreply just send it to client directely
			log.Printf("This is %d", n.id)
			if n.lastreply.Equal(msg) {
				// TODO just Reply
			} else {
				n.SendReqtoSubPrimary(msg)
			}
		}
		for _,v := range msg.Requests{
			log.Printf("[Req]This is req:%d from client:%d",v.N,v.C,v.T)
			n.buffer.AppendToRequestQueue(v)
		}
		n.PrePrepareSendNotify <- true
	}
}
