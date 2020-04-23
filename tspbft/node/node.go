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
	executenum  *ExecuteOpNum

	buffer      *message.Buffer

	RequestRecv      chan *message.Request
	PrePrepareRecv   chan *message.PrePrepare
	PrepareRecv      chan *message.Prepare
	CommitRecv       chan *message.Commit
	CheckPointRecv   chan *message.CheckPoint

	PrePrepareSendNotify  chan bool
	ExecuteNotify         chan bool

	Supports          map[string]consensus.ConsenterSupport


}

func NewNode(conf *cmf.ShareConfig, support consensus.ConsenterSupport) *Node {
	node := &Node{
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
		executenum:            NewExecuteOpNum(),

		buffer:                message.NewBuffer(),

		RequestRecv:           make(chan *message.Request),
		PrePrepareRecv:        make(chan *message.PrePrepare),
		PrepareRecv:           make(chan *message.Prepare),
		CommitRecv:            make(chan *message.Commit),
		CheckPointRecv:        make(chan *message.CheckPoint),

		PrePrepareSendNotify:  make(chan bool), //hou mian mei 100 jiushi yao zuse le shoudaole cai jinxiayige
		ExecuteNotify:         make(chan bool, 100),
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
	n.server.RegisterChan(n.RequestRecv, n.PrePrepareRecv, n.PrepareRecv, n.CommitRecv, n.CheckPointRecv)
	go n.server.Run()
	go n.RequestRecvThread()
	go n.PrePrepareSendThread()
	go n.PrePrepareRecvAndSendPrepareThread()
	go n.PrepareRecvAndSendCommitThread()
	go n.CommitRecvAndVertifyThread()
	go n.CommitRecvAndBroadThread()
	go n.ExecuteAndReplyThread()
	go n.CheckPointRecvThread()
}