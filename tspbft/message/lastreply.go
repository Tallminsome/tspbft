package message

import "sync"

// node last reply to discard requests whose timestamp is lower than lastreply
type LastReply struct {
	reply  *Reply
	locker *sync.RWMutex
}

func NewLastReply() *LastReply {
	return &LastReply{
		reply:  nil,
		locker: new(sync.RWMutex),
	}
}

// only read
func (r *LastReply) Equal(msg *Request) bool {
	r.locker.RLock()
	flag := true
	if r.reply == nil || r.reply.T != msg.T {
		flag = false
	}
	r.locker.RUnlock()
	return flag
}

func (r *LastReply) Set(msg *Reply) {
	r.locker.Lock()
	r.reply = msg
	r.locker.Unlock()
}
