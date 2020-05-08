package server

import (
	"encoding/json"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"log"
	"net/http"
)

func (ser *HttpServer) HttpBufferReq (w http.ResponseWriter, r *http.Request){
	var msg message.BufferReq
	log.Printf("who send request: %s",r.RemoteAddr)
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.bufferreqRecv <- &msg
}

func (ser *HttpServer) HttpRequest (w http.ResponseWriter, r *http.Request){
	var msg message.Request
	log.Printf("who send request: %s",r.RemoteAddr)
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.requestRecv <- &msg
}

func (ser *HttpServer) HttpPreprepare (w http.ResponseWriter, r *http.Request) {
	var msg message.PrePrepare
	log.Printf("who send pre-prepare: %s",r.RemoteAddr)
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.preprepareRecv <- &msg
}

func (ser *HttpServer) HttpPrepare (w http.ResponseWriter, r *http.Request) {
	var msg message.Prepare
	log.Printf("who send prepare: %s",r.RemoteAddr)
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.prepareRecv <- &msg
}

func (ser *HttpServer) HttpCommit (w http.ResponseWriter, r *http.Request) {
	var msg message.Commit
	log.Printf("who send commit: %s",r.RemoteAddr)
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.commitRecv <- &msg
}

func (ser *HttpServer) HttpVerify (w http.ResponseWriter, r *http.Request) {
	var msg message.Verify
	log.Printf("who send verify: %s",r.RemoteAddr)
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.verifyRecv <- &msg
}

func (ser *HttpServer) HttpVerified (w http.ResponseWriter, r *http.Request) {
	var msg message.Verified
	log.Printf("who send verified: %s",r.RemoteAddr)
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.verifiedRecv <- &msg
}

func (ser *HttpServer) HttpCheckpoint (w http.ResponseWriter, r *http.Request) {
	var msg message.CheckPoint
	log.Printf("who send checkpoint: %s",r.RemoteAddr)
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		log.Printf("[Http Error]: %s", err)
		return
	}
	ser.checkpointRecv <- &msg
}