package server

import (
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/cmf"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"log"
	"net/http"
	"strconv"
)

const (
	RequestEntry     = "/request"
	PrePrepareEntry  = "/preprepare"
	PrepareEntry     = "/prepare"
	CommitEntry      = "/commit"
	VerifyEntry      = "/verify"
	VerifiedEntry    = "/verified"
	CheckPointEntry  = "/checkpoint"
)

//http jiantongqingqiu
type HttpServer struct {
	port    int
	server  *http.Server

	requestRecv    chan *message.Request
	preprepareRecv chan *message.PrePrepare
	prepareRecv    chan *message.Prepare
	commitRecv     chan *message.Commit
	verifyRecv     chan *message.Verify
	verifiedRecv     chan *message.Verified
	checkpointRecv chan *message.CheckPoint
}

func NewServer(conf *cmf.ShareConfig) *HttpServer {
	httpserver := &HttpServer{
		port:           conf.Port,
		server:         nil,
	}
	return httpserver
}

//config server: register the handle channel
func (ser *HttpServer) RegisterChan (r chan *message.Request, pre chan *message.PrePrepare, p chan *message.Prepare,
	c chan *message.Commit, v chan *message.Verify, vd chan *message.Verified, ck chan *message.CheckPoint)  {
	log.Printf("Registering the chan for listen func")
	ser.requestRecv     = r
	ser.preprepareRecv  = pre
	ser.prepareRecv     = p
	ser.commitRecv      = c
	ser.verifyRecv      = v
	ser.verifiedRecv    = vd
	ser.checkpointRecv  = ck
}

func (ser *HttpServer) Run() {
	//register se
	log.Printf("[Node] start the listen server")
	ser.RegisterServer()
}

func (ser *HttpServer) RegisterServer() {
	log.Printf("[Server] set listen port:%d\n", ser.port)

	httpRegister := map[string]func(http.ResponseWriter, *http.Request){
		RequestEntry:    ser.HttpRequest,
		PrePrepareEntry: ser.HttpPreprepare,
		PrepareEntry:    ser.HttpPrepare,
		CommitEntry:     ser.HttpCommit,
		VerifyEntry:     ser.HttpVerify,
		VerifiedEntry:   ser.HttpVerified,
		CheckPointEntry: ser.HttpCheckpoint,
	}

	mux := http.NewServeMux()
	for k, v := range httpRegister {
		log.Printf("[Server] register the func for %s", k)
		mux.HandleFunc(k, v)
	}

	ser.server = &http.Server{
		Addr:    ":" + strconv.Itoa(ser.port),
		Handler: mux,
	}

	if err := ser.server.ListenAndServe(); err != nil {
		log.Printf("[Server Error] %s", err)
		return
	}


}

