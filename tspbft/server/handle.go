package server

import (
	"encoding/json"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"log"
	"net/http"
)

func (ser *HttpServer) HttpRequest (w http.ResponseWriter, r *http.Request){
	var msg message.Request
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.requestRecv <- &msg
}

func (ser *HttpServer) HttpPreprepare (w http.ResponseWriter, r *http.Request) {
	var msg message.PrePrepare
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.preprepareRecv <- &msg
}

func (ser *HttpServer) HttpPrepare (w http.ResponseWriter, r *http.Request) {
	var msg message.Prepare
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.prepareRecv <- &msg
}

func (ser *HttpServer) HttpCommit (w http.ResponseWriter, r *http.Request) {
	var msg message.Commit
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.commitRecv <- &msg
}

func (ser *HttpServer) HttpCheckpoint (w http.ResponseWriter, r *http.Request) {
	var msg message.CheckPoint
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.checkpointRecv <- &msg
}