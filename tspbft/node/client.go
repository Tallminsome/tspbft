package node

import (
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"log"
	"time"
)

func (n *Node) ClientThread() {
	log.Printf("Start buffering the request")
	requestBuffer := make([]*message.BufferReq, 0)
	sortreqBuffer := make([]*message.BufferReq, 0)
	var timer <-chan time.Time
	for {
		select {
		case msg := <-n.BufferReqRecv:
			timer = nil
			requestBuffer = append(requestBuffer,msg)
			sortreqBuffer = n.SortRequest(requestBuffer)
			if msg.O.Type == message.TYPECONFIG {
				//若为系统配置请求,则无需打包,需优先配置
				request := &message.Request{
					Requests: sortreqBuffer,
					T:        message.TimeStamp(time.Now().UnixNano()),
				}
				n.SendReqtoSubPrimary(request)
				log.Printf("[Client] send request(%d) due to config", len(sortreqBuffer))
				requestBuffer = make([]*message.BufferReq, 0)
				sortreqBuffer = make([]*message.BufferReq, 0)
			}else if len(requestBuffer) >= 50 {
				//等到消息发送到100个时,一起排序打包发送
				request := &message.Request{
					Requests: sortreqBuffer,
					T:        message.TimeStamp(time.Now().UnixNano()),
				}
				n.SendReqtoSubPrimary(request)
				log.Printf("[Client] send request(%d) because len >= 100", len(requestBuffer))
				requestBuffer = make([]*message.BufferReq, 0)
				sortreqBuffer = make([]*message.BufferReq, 0)
			}
			timer = time.After(time.Second)
		case <-timer:
			//一段时间内没有收到100个请求,触发超时重发
			timer = nil
			if len(requestBuffer) > 0 {
				request := &message.Request{
					Requests: sortreqBuffer,
					T:        message.TimeStamp(time.Now().UnixNano()),
				}
				n.SendReqtoSubPrimary(request)
				log.Printf("[Client] send request(%d) due to over time", len(requestBuffer))
				requestBuffer = make([]*message.BufferReq, 0)
				sortreqBuffer = make([]*message.BufferReq, 0)
			}
			timer = time.After(time.Second)
		}
	}
}

func (n *Node) SortRequest(reqbuf []*message.BufferReq) []*message.BufferReq {
	//请求大小为1或0时,不用排序
	if len(reqbuf) <= 1 {
		log.Printf("[sort]len req <= 1")
		return reqbuf
	}
	//根据请求的时间戳大小进行排序,按时间顺序从前往后排序
	for k,_ := range reqbuf {
		if k == len(reqbuf) - 1 {
			break
		}
		if reqbuf[k].T > reqbuf[k+1].T {
			min := reqbuf[k+1]
			reqbuf[k+1] = reqbuf[k]
			reqbuf[k] = min
		}
	}
	return reqbuf
}
