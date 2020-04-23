package node

import (
	"fmt"
	"log"
)

func (n *Node) RequestRecvThread() {
	log.Printf("Start receiving the request")
	for {
		msg := <- n.RequestRecv
		fmt.Printf("Enter the request")
		// check if it's subprimary
		if !n.WhetherSubPrimary(){
			// if it's lastreply just send it to client directely
			fmt.Printf("This is %d",n.id)
			if n.lastreply.Equal(msg) {
				// TODO just Reply
			}else {
				// TODO just send it to sub-primary
				n.SendReqtoSubPrimary(msg)
				fmt.Println("[Request] SendtoSubPrimary")
			}
		}
		n.buffer.AppendToRequestQueue(msg)
		n.PrePrepareSendNotify <- true
	}
}
