package tspbft

import (
	"fmt"
	"github.com/hyperledger/fabric/orderer/consensus"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/cmf"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/node"
	cb "github.com/hyperledger/fabric/protos/common"
	"log"
	"time"
)
var reqnum = 0
type Chain struct {
	outchain chan struct{}
	support consensus.ConsenterSupport
	tspbftNode *node.Node
}

func NewChain(support consensus.ConsenterSupport) *Chain {
	logger.Info("Creating New Chain-----ChainID:", support.ChainID())
	//if there is no node
	if node.TSNode == nil {
		node.TSNode = node.NewNode(cmf.LoadConfig(),support)
		node.TSNode.Run()
	}else {
		node.TSNode.RegisterChain(support)
	}

	ch := &Chain{
		outchain:   make(chan struct{}),
		support:    support,
		tspbftNode: node.TSNode,
	}
	return ch

}
//deal with the interface
//start
func (chain *Chain) Start (){
	logger.Info("start")
}
//close the chain
func (chain *Chain) Halt (){
	logger.Info("hault")
	select {
	case <- chain.outchain:
	default:
		close(chain.outchain)
	}
}
//error send
func (chain *Chain) Errored() <-chan struct{} {
	return chain.outchain
}
//before orderer config
func (chain *Chain) WaitReady() error {
	logger.Info("Wait Ready")
	return nil
}
// receive order
func (chain *Chain) Order (env *cb.Envelope, configseq uint64) error {
	logger.Info("Noraml Type")
	reqnum = reqnum + 1
	select {
	case <- chain.outchain:
		logger.Info("error exit chain")
		fmt.Errorf("Exiting...")
	default:
	}
	O := message.Operation{
		Envelope:   env,
		ChannelID:  chain.support.ChainID(),
		ConfigSeq:  configseq,
		Type:       message.TYPENORMAL,
	}
	log.Printf("[Req]Gnerate normal")
	_, msg := message.NewBufferMsg(O, message.TimeStamp(time.Now().UnixNano()), chain.tspbftNode.GetId(),reqnum)
	log.Printf("[Chain]Normal Type,[Req]id:%d, client:%d",reqnum,msg.C)

	chain.tspbftNode.SendRequesttoPrimary(msg)
	return nil
}

func (chain *Chain) Configure (config *cb.Envelope, configseq uint64) error {
	logger.Info("Config Type")
	reqnum = reqnum + 1
	select {
	case <- chain.outchain:
		logger.Info("error exit config")
		return fmt.Errorf("Exiting...")
	default:

	}
	O := message.Operation{
		Envelope:   config,
		ChannelID:  chain.support.ChainID(),
		ConfigSeq:  configseq,
		Type:       message.TYPECONFIG,
	}
	log.Printf("[Req]Gnerate config")
	_, msg := message.NewBufferMsg(O, message.TimeStamp(time.Now().UnixNano()), chain.tspbftNode.GetId(),reqnum)
	log.Printf("[Chain]Config Type,[Req]id:%d, client:%d",reqnum,msg.C)
	chain.tspbftNode.SendRequesttoPrimary(msg)
	return nil
}