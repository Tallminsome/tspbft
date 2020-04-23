package message

import (
	"encoding/json"
	"log"
	"sync"
)

type Buffer struct {
	RequestQueue  []*Request
	RequestLocker *sync.RWMutex

	PrePrepareBuffer map[string]*PrePrepare
	PrePrepareSet    map[string]bool
	PrePrepareLocker *sync.RWMutex

	PrepareSet       map[string]map[Identify]bool
	PrepareState     map[string]bool
	PrepareLocker    *sync.RWMutex

	CommitSet        map[string]map[Identify]bool
	CommitState      map[string]bool
	CommitLocker     *sync.RWMutex

	ExecuteQueue     []*PrePrepare
	ExecuteLocker    *sync.RWMutex

	CheckPointBuffer map[string]map[Identify]bool
	CheckPointState  map[string]bool
	CheckPointLocker *sync.RWMutex
}

func NewBuffer() *Buffer {
	return &Buffer{
		RequestQueue:      make([]*Request, 0),
		RequestLocker:     new(sync.RWMutex),

		PrePrepareBuffer: make(map[string]*PrePrepare),
		PrePrepareSet:    make(map[string]bool),
		PrePrepareLocker: new(sync.RWMutex),

		PrepareSet:       make(map[string]map[Identify]bool),
		PrepareState:     make(map[string]bool),
		PrepareLocker:    new(sync.RWMutex),

		CommitSet:        make(map[string]map[Identify]bool),
		CommitState:      make(map[string]bool),
		CommitLocker:     new(sync.RWMutex),

		ExecuteQueue:     make([]*PrePrepare, 0),
		ExecuteLocker:    new(sync.RWMutex),

		CheckPointBuffer: make(map[string]map[Identify]bool),
		CheckPointState:  make(map[string]bool),
		CheckPointLocker: new(sync.RWMutex),
	}
}
//add request to the end of the queue
func (b *Buffer) AppendToRequestQueue(req *Request) {
	b.RequestLocker.Lock()
	b.RequestQueue = append(b.RequestQueue, req)
	b.RequestLocker.Unlock()
}
//batch to clean the queue
func (b *Buffer) BatchRequest() (batch []*Request) {
	batch = make([]*Request, 0)
	b.RequestLocker.Lock()
	for _,req := range b.RequestQueue {
		batch = append(batch,req)
	}
	b.RequestQueue = make([]*Request,0)
	b.RequestLocker.Unlock()
	return batch
}

func (b *Buffer) SizeofRequestQueue() (l int) {
	b.RequestLocker.RLock()
	l = len(b.RequestQueue)
	b.RequestLocker.RUnlock()
	return
}

// buffer about pre-prepare
func (b *Buffer) BufferPrePrepareMsg(grp string,msg *PrePrepare) {
	b.PrepareLocker.Lock()
	b.PrePrepareBuffer[msg.D] = msg
	var grps  = ""
	switch grp {
	case "AB":
		grps = "B"
	case "AC":
		grps = "C"
	case "AD":
		grps = "D"
	default:
		grps = grp
	}
	b.PrePrepareSet[GroupSequenceString(grps, msg.N)] = true
	b.PrepareLocker.Unlock()
}

func (b *Buffer) CleanPrePrepareMsg(grp string,digest string) {
	b.PrePrepareLocker.Lock()
	msg := b.PrePrepareBuffer[digest]

	delete(b.PrePrepareSet, GroupSequenceString(grp, msg.N))
	delete(b.PrePrepareBuffer, digest)
	b.PrePrepareLocker.Unlock()
}

func (b *Buffer) WhetherExistPrePrepareMsg(grp string, seq Sequence) bool {
	key := GroupSequenceString(grp, seq)
	b.PrePrepareLocker.RLock()
	if _, ok := b.PrePrepareSet[key]; ok {
		b.PrePrepareLocker.RUnlock()
		log.Printf("Exist PrePrepareMsg")
		return true
	}
	b.PrePrepareLocker.RUnlock()
	log.Printf("Do not Exist PrePrepareMsg")
	return false
}

func (b *Buffer) FetchPrePrepareMsg(digest string) (pre *PrePrepare) {
	pre = nil
	b.PrePrepareLocker.RLock()
	if _, ok := b.PrePrepareBuffer[digest]; !ok {
		log.Printf("Error finding pre-prepare message")
		return
	}
	pre = b.PrePrepareBuffer[digest]
	b.PrePrepareLocker.RUnlock()
	return pre
}

//buffer prepare
func (b *Buffer) BufferPrepareMsg(msg *Prepare) {
	b.PrepareLocker.Lock()
	if _, ok := b.PrepareSet[msg.D]; !ok {
		b.PrepareSet[msg.D] = make(map[Identify]bool) //ta de value shi Identify bool
	}
	b.PrepareSet[msg.D][msg.I] = true
	b.PrepareLocker.Unlock()
}

func (b *Buffer) CleanPrepareMsg(digest string) {
	b.PrepareLocker.Lock()
	delete(b.PrepareSet, digest)
	delete(b.PrepareState, digest)
	b.PrepareLocker.Unlock()
}

func (b *Buffer) TrueOfPrepareMsg(digest string, faultnum uint) bool {
	b.PrepareLocker.Lock()
	num := uint(len(b.PrepareSet[digest]))
	log.Printf("length:%d",num)
	_, ok := b.PrepareState[digest]
	if ok {
		log.Printf("Exist Commit State")
	} else {
		log.Printf("Do not exist Commit State")
	}
	if num < 2*faultnum || ok {
		log.Printf("< 2f")
		b.PrepareLocker.Unlock()
		return false
	}
	b.PrepareState[digest] = true
	log.Printf("set preparestate == true")
	b.PrepareLocker.Unlock()
	return true
}
// commit
func (b *Buffer) BufferCommitMsg(msg *Commit) {
	b.CommitLocker.Lock()
	if _, ok := b.CommitSet[msg.D]; !ok {
		b.CommitSet[msg.D] = make(map[Identify]bool)
	}
	b.CommitSet[msg.D][msg.I] = true
	b.CommitLocker.Unlock()
}

func (b *Buffer) CleanCommitMsg(digest string) {
	b.CommitLocker.Lock()
	delete(b.CommitSet, digest)
	delete(b.CommitState, digest)
	b.CommitLocker.Unlock()
}
// 2F + 1
func (b *Buffer) TrueOfCommitMsg(digest string, fault uint) bool {
	b.CommitLocker.Lock()
	num := uint(len(b.CommitSet[digest]))
	log.Printf("Commit num : %d",num)
	_, ok := b.CommitState[digest]
	if ok {
		log.Printf("Exist Commit State")
	} else {
		log.Printf("Do not exist Commit State")
	}
	if num < 2*fault+1 || ok {
		b.CommitLocker.Unlock()
		log.Printf("There is not enough nodes to commit")
		return false
	}
	b.CommitState[digest] = true
	b.CommitLocker.Unlock()
	return true
}

func (b *Buffer) WhetherToExecute(digest string, grp string, seq Sequence) bool {
	b.PrepareLocker.RLock()
	defer b.PrepareLocker.RUnlock()

	_, isPrepare := b.PrepareState[digest]
	_, isCommit  := b.CommitState[digest]
	if isPrepare {
		log.Printf("Exist Prepare")
	}else {
		log.Printf("Do not Exist Prepare")
	}
	if b.WhetherExistPrePrepareMsg(grp, seq) && isPrepare && isCommit {
		return true
	}
	return false
}

func (b *Buffer) AppendToExecuteQueue(msg *PrePrepare) {
	b.ExecuteLocker.Lock()
	// er fen pai xu ,find the first index greater than msg to insert
	count := len(b.ExecuteQueue)    //count record times to find the index
	fir := 0
	for count > 0 {
		step := count / 2
		index := step + fir
		if !(msg.N < b.ExecuteQueue[index].N) {
			fir = index +1
			count = count - step - 1
		} else {
			count = step
		}
	}

	b.ExecuteQueue = append(b.ExecuteQueue, msg)
	copy(b.ExecuteQueue[fir + 1:], b.ExecuteQueue[fir:])//hou yi teng weizhi
	b.ExecuteQueue[fir] = msg
	b.ExecuteLocker.Unlock()
}

func (b *Buffer) BatchExecute(lastseq Sequence) ([]*PrePrepare, Sequence) {
	b.ExecuteLocker.Lock()
	batchs := make([]*PrePrepare, 0)
	index := lastseq
	loop := 0
	// batch form startSeq sequentially
	for {
		if loop == len(b.ExecuteQueue) {
			b.ExecuteQueue = make([]*PrePrepare, 0)
			b.ExecuteLocker.Unlock()
			return batchs, index
		}
		if b.ExecuteQueue[loop].N != index+1 {
			b.ExecuteQueue = b.ExecuteQueue[loop:]
			b.ExecuteLocker.Unlock()
			return batchs, index
		}
		batchs = append(batchs, b.ExecuteQueue[loop])
		loop = loop + 1
		index = index + 1
	}
}

func (b *Buffer) NewCheckPoint(seq Sequence, id Identify) ([]byte, *CheckPoint) {
	temp := make(map[Sequence]string, 0)
	minSeq := seq
	content := ""

	b.PrePrepareLocker.RLock()
	for k, v := range b.PrePrepareBuffer {
		if v.N <= seq {
			temp[v.N] =k
			if v.N < minSeq {
				minSeq = v.N
			}
		}
	}
	b.PrePrepareLocker.RUnlock()

	for minSeq <= seq {
		d := temp[minSeq]
		content = content + d
		minSeq = minSeq + 1
	}

	msg := &CheckPoint{
		N: seq,
		D: Hash([]byte(content)),
		I: id,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, nil
	}
	return data, msg
}

func (b *Buffer) CleanBuffer(grp string,msg *CheckPoint) {
	temp := make(map[Sequence]string, 0)
	minSeq := msg.N

	b.PrePrepareLocker.RLock()
	for k, v := range b.PrePrepareBuffer {
		if v.N <= msg.N {
			temp[v.N] =k
			if v.N < minSeq {
				minSeq = v.N
			}
		}
	}
	b.PrePrepareLocker.RUnlock()

	for minSeq <= msg.N {
		b.CleanPrePrepareMsg(grp,temp[minSeq])
		b.CleanPrepareMsg(temp[minSeq])
		b.CleanCommitMsg(temp[minSeq])
		minSeq = minSeq + 1
	}
}

func (b *Buffer) BufferCheckPointMsg(msg *CheckPoint, id Identify) {
	b.CheckPointLocker.Lock()
	if _, ok := b.CheckPointBuffer[msg.D]; !ok {
		b.CheckPointBuffer[msg.D] = make(map[Identify]bool)
	}
	b.CheckPointBuffer[msg.D][id] = true
	b.CheckPointLocker.Unlock()
}

func (b *Buffer) TrueOfCheckPointMsg(digest string, fault uint) bool {
	flag := false
	b.CheckPointLocker.RLock()
	num := uint(len(b.CheckPointBuffer[digest]))
	_, ok := b.CheckPointState[digest]
	if num < 2*fault || ok {
		b.CheckPointLocker.RUnlock()
		return flag
	}
	b.CheckPointState[digest] = true
	flag = true
	b.CheckPointLocker.RUnlock()
	return flag
}

func (b *Buffer) Show() {
	log.Printf("[Buffer] node buffer size: pre-prepare(%d) prepare(%d) commit(%d)", len(b.PrePrepareBuffer), len(b.PrepareSet), len(b.CommitSet))

}