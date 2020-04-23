package node

import (
	"sync"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/message"
	"github.com/hyperledger/fabric/orderer/consensus/tspbft/cmf"
)


type Sequence struct {
	lastSequence  message.Sequence
	checkSequence message.Sequence
	stepSequence  message.Sequence
	sequence      message.Sequence
	waterL	      message.Sequence
	waterH        message.Sequence
	checkLocker   bool
	locker 		  *sync.RWMutex
}

func NewSequence(conf *cmf.ShareConfig) *Sequence {
	seq := &Sequence{
		lastSequence:  0,   //start from 0
		checkSequence: conf.CheckPointNum,   //lastsquence-1
		stepSequence:  conf.CheckPointNum,   //meiyixiabu
		sequence:      0,   //start from 0
		waterL:        message.Sequence(conf.WaterLow),
		waterH:        message.Sequence(conf.WaterHigh),
		checkLocker:   false,
		locker:        new(sync.RWMutex),
	}
	return seq
}

func (seq *Sequence) Add() message.Sequence {
	seq.locker.Lock()  //WRITE ONLY
	seq.sequence = seq.sequence + 1
	seq.locker.Unlock()
	return seq.sequence
}

func (seq *Sequence) CheckValid(mseq message.Sequence) bool {
	seq.locker.RLock()  //read only
	defer seq.locker.RUnlock()
	if mseq < seq.lastSequence {
		return false
	}
	if mseq < seq.waterL || mseq > seq.waterH {
		return false
	}
	return true
}

func (seq *Sequence) SetLastSequence(sequence message.Sequence) {
	seq.locker.Lock()
	seq.lastSequence = sequence
	seq.locker.Unlock()
}

func (seq *Sequence) GetLastSequence() (sequence message.Sequence) {
	seq.locker.RLock()
	sequence = seq.lastSequence
	seq.locker.RUnlock()
	return sequence
}
//get check sequence
func (seq *Sequence) GetCheckpoint() (sequence message.Sequence) {
	seq.locker.RLock()
	sequence = seq.checkSequence
	seq.locker.RUnlock()
	return sequence
}
//to reset the waterlevel after fresh
func (seq *Sequence) CheckPoint() {
	seq.locker.Lock()
	seq.waterL = seq.checkSequence + 1
	seq.checkSequence = seq.checkSequence + seq.stepSequence   //shi yin wei ba zhiqian d qing le
	seq.waterH = seq.checkSequence + seq.stepSequence * 2
	seq.checkLocker = false   //guo le ready to checkpoint
	seq.locker.Unlock()
	return
}

func (seq *Sequence) ReadyToCheckPoint() (flag bool){
	flag = false
	seq.locker.RLock()
	if seq.lastSequence >= seq.checkSequence && !seq.checkLocker {
		seq.checkLocker = true
		flag = true
	}
	seq.locker.RUnlock()
	return
}