package node

import (
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"log"
	"time"
)

func (n *Node) ClientThread() {
	log.Printf("Start buffering the request")
	requestBuffer := make([]*message.BufferReq, 0)
	var timer <-chan time.Time
	for {
		select {
		case msg := <-n.BufferReqRecv:
			log.Printf("[BufReq]This is req:%d from client:%d,time:%s",msg.N,msg.C,msg.T)
			timer = nil
			requestBuffer = append(requestBuffer,msg)
			// xian xie config
			if msg.O.Type == message.TYPECONFIG {
				request := &message.Request{
					Requests: requestBuffer,
					T:        message.TimeStamp(time.Now().UnixNano()),
				}
				n.SendReqtoSubPrimary(request)
				log.Printf("[Client] send request(%d) due to config", len(requestBuffer))
				requestBuffer = make([]*message.BufferReq, 0)
			}else if len(requestBuffer) >= 100 {
				log.Printf("Enter the line:%d",n.id)
				request := &message.Request{
					Requests: requestBuffer,
					T:        message.TimeStamp(time.Now().UnixNano()),
				}
				n.SendReqtoSubPrimary(request)
				log.Printf("[Client] send request(%d) because reatch the line", len(requestBuffer))
				requestBuffer = make([]*message.BufferReq, 0)
			}
			log.Printf("len in over time:%d",len(requestBuffer))
			timer = time.After(time.Second)
		case <-timer:
			log.Printf("Enter the timer:%d",n.id)
			timer = nil
			if len(requestBuffer) > 0 {
				request := &message.Request{
					Requests: requestBuffer,
					T:        message.TimeStamp(time.Now().UnixNano()),
				}
				n.SendReqtoSubPrimary(request)
				log.Printf("[Client] send request(%d) due to over time", len(requestBuffer))
				requestBuffer = make([]*message.BufferReq, 0)
			}
			timer = time.After(time.Second)
		}
	}
}
