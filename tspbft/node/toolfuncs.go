package node

import (
	"log"
	"os"
	"strings"
	"sync"
)

func (n *Node) GetPrimary () string {
	RawTable := os.Getenv("TSPBFT_NODE_TABLE_A")
	Tables  := strings.Split(RawTable,";")
	return Tables[0]
}

func (n *Node) GetSubPrimary () string {
	switch n.group {
	case "B","AB":
		return n.table["B"][0]
	case "C","AC":
		return n.table["C"][0]
	case "D","AD":
		return n.table["D"][0]
	default:
		log.Printf("Error,this is primary node,send request to orderer2")
		return n.table["B"][0]
	}
}

func (n *Node) GetAllSubPrimary () []string {
	RawTable := os.Getenv("TSPBFT_NODE_TABLE_A")
	Tables  := strings.Split(RawTable,";")
    var AllSubPrimary = make([]string,len(Tables) - 1)
	for k,v := range Tables{
		if k == 0 {
			continue
		}
		AllSubPrimary[k-1] = v
	}
	return AllSubPrimary
}

func (n *Node) WhetherPrimary () bool {
	test := os.Getenv("TSPBFT_NODE_GROUP")
	if test == "A" {
		return true
	}
	return false
}

func (n *Node) WhetherSubPrimary () bool {
	test := os.Getenv("TSPBFT_NODE_GROUP")
	if test == "AB" || test ==  "AC" || test == "AD" {
		return true
	}
	return false
}

// the execute op num now in state
type ExecuteOpNum struct {
	num    int
	locker *sync.RWMutex
}

func NewExecuteOpNum() *ExecuteOpNum {
	return &ExecuteOpNum{
		num:    0,
		locker: new(sync.RWMutex),
	}
}

func (n *ExecuteOpNum) Get() int {
	return n.num
}

func (n *ExecuteOpNum) Inc() {
	n.num = n.num + 1
}

func (n *ExecuteOpNum) Dec() {
	n.Lock()
	n.num = n.num - 1
	n.UnLock()
}

func (n *ExecuteOpNum) Lock() {
	n.locker.Lock()
}

func (n *ExecuteOpNum) UnLock() {
	n.locker.Unlock()
}
