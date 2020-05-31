package node

import (
	"log"
)
var num = 0
func (n *Node) RequestRecvThread() {
	for {
		msg := <-n.RequestRecv
		num = num + 1
		log.Printf("[Req]This is req:%d,time stamp:%s",num,msg.T)
		//添加进请求队列中
		n.buffer.AppendToRequestQueue(msg)
		//进入pre-prepare消息生成及发送线程
		n.PrePrepareSendNotify <- true
	}
}
