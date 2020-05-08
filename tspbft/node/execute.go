package node

import "log"

func (n *Node) ReadyToExecute(digest string) {
	n.buffer.AppendToExecuteQueue(n.buffer.FetchPrePrepareMsg(digest))
	n.ExecuteNotify <- true
	log.Printf("ReadyToExecute")
}
