package node

import (
	"github.com/hyperledger/fabric/orderer/consensus"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/cmf"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/server"
	"log"
)

var TSNode *Node = nil
type Node struct {
	comcfg    *cmf.ShareConfig
	server    *server.HttpServer

	id        message.Identify
	pri       message.Primary
	sub       message.SubPrimary
	group     message.Group
	table     map[message.Group][]string
	faultnum  uint

	lastreply   *message.LastReply
	sequence    *Sequence
	executeNum  *ExecuteOpNum

	buffer      *message.Buffer

	BufferReqRecv    chan *message.BufferReq
	RequestRecv      chan *message.Request
	PrePrepareRecv   chan *message.PrePrepare
	PrepareRecv      chan *message.Prepare
	CommitRecv       chan *message.Commit
	VerifyRecv       chan *message.Verify
	VerifiedRecv     chan *message.Verified
	CheckPointRecv   chan *message.CheckPoint

	PrePrepareSendNotify  chan bool
	ExecuteNotify         chan bool

	Supports          map[string]consensus.ConsenterSupport


}

func NewNode(conf *cmf.ShareConfig, support consensus.ConsenterSupport) *Node {
	node := &Node {
		comcfg:                conf,
		server:                server.NewServer(conf),

		id:                    conf.Id,
		pri:                   conf.Primary,
		sub:                   conf.SubPrimary,
		group:                 conf.Group,
		table:                 conf.Table,
		faultnum:              conf.FaultNum,

		lastreply:             message.NewLastReply(),
		sequence:              NewSequence(conf),
		executeNum:            NewExecuteOpNum(),

		buffer:                message.NewBuffer(),

		BufferReqRecv:         make(chan *message.BufferReq,1000),
		RequestRecv:           make(chan *message.Request),
		PrePrepareRecv:        make(chan *message.PrePrepare),
		PrepareRecv:           make(chan *message.Prepare),
		CommitRecv:            make(chan *message.Commit),
		VerifyRecv:            make(chan *message.Verify),
		VerifiedRecv:          make(chan *message.Verified),
		CheckPointRecv:        make(chan *message.CheckPoint),

		PrePrepareSendNotify:  make(chan bool), //hou mian mei 100 jiushi yao zuse le shoudaole cai jinxiayige
		ExecuteNotify:         make(chan bool, 400),
		Supports:              make(map[string]consensus.ConsenterSupport),
	}
	log.Printf("[Node] the node id:%d, group:%s, fault number:%d\n", node.id, node.group, node.faultnum)
	node.RegisterChain(support)
	return node
}

func (n *Node) RegisterChain (support consensus.ConsenterSupport)  {
	if _, ok := n.Supports[support.ChainID()]; ok {
		return
	}
	log.Printf("[Node] Register the chain(%s)", support.ChainID())
	n.Supports[support.ChainID()] = support
}

func (n *Node) Run() {
	//first register chan for server
	n.server.RegisterChan(n.BufferReqRecv,n.RequestRecv, n.PrePrepareRecv, n.PrepareRecv, n.CommitRecv, n.VerifyRecv, n.VerifiedRecv, n.CheckPointRecv)
	go n.server.Run()
	if n.WhetherPrimary() {
		go n.ClientThread()
	}
	go n.RequestRecvThread()
	go n.PrePrepareSendThread()
	go n.PrePrepareRecvAndSendPrepareThread()
	go n.PrepareRecvAndSendCommitThread()
	go n.CommitRecvAndVertifyThread()
	go n.VerifyRecvAndBroadThread()
	go n.ReplicaRecvVerifyThread()
	go n.ExecuteAndReplyThread()
	go n.CheckPointRecvThread()
}

func (n *Node) GetId() message.Identify {
	return n.id
}