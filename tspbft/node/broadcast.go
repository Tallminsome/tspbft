package node

import (
	"bytes"
	"encoding/json"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/server"
	"log"
	"net/http"
)

func (n *Node) SendComtoPrimary (content []byte, handle string) { // send de bushi msg ershi toupiaojieguo
	go SendPost(content, n.GetPrimary() + handle)
}

func (n *Node) BroadtoSubPrimary (msg *message.Commit, handle string) { // send de bushi msg ershi toupiaojieguo
	content, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error to marshal json")
		return
	}
	for k,v := range n.table {
		if k != "A" {
			continue
		}
		for i := 1; i < len(v); i++ {
			go SendPost(content, v[i] + handle)
		}
	}
}

func (n *Node) SendReqtoSubPrimary (msg *message.Request) {
	content, err := json.Marshal(msg)
	if err != nil {
		log.Printf("error to marshal json")
		return
	}
	for i:=0; i < 3; i++{
		go SendPost(content, n.GetAllSubPrimary()[i] + server.RequestEntry)
	}
}

func (n *Node) SendPreparetoSubPrimary (content []byte, handle string) {
	go SendPost(content, n.GetSubPrimary() + handle)
}

func (n *Node) Broadcast(content []byte, handle string) {
	for k, v := range n.table {
		//do not send to primary
		if k == "A"  {
			continue
		}
		//do not send to my self
		for i := 1; i < len(v); i++ {
			go SendPost(content, v[i] + handle)  //v[0] shi zi ji
		}
	}
}

func SendPost(content []byte, url string) {
	buff := bytes.NewBuffer(content)
	if _, err := http.Post(url, "application/json", buff); err != nil {
		log.Printf("[Send] send to %s error: %s", url, err)
	}
}