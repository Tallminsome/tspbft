package message

import (
	"encoding/json"
	cb "github.com/hyperledger/fabric/protos/common"
	"strconv"
)

type TimeStamp uint64 // 时间戳格式
type Identify uint64  // 客户端标识格式
type Sequence int64   // 序号

type Primary    bool    // shangcengtongdaozhujiedian
type SubPrimary bool // xiacengtongdaozhujiedian
type Group      string

const TYPENORMAL = "normal"
const TYPECONFIG = "config"

type Operation struct {
	Envelope  *cb.Envelope
	ChannelID string
	ConfigSeq uint64
	Type      string

}
// Request
type BufferReq struct {
	O Operation `json:"operation"`
	T TimeStamp `json:"timestamp"`
	C Identify  `json:"clientID"`
	N int       `json:"reqnum"`
}
// Message (Buffered requests)
type Request struct {
	Requests []*BufferReq `json:"requests"`
	T           TimeStamp `json:"timestamp"`
}
//type Request struct {
//	O Operation `json:"operation"`
//	T TimeStamp `json:"timestamp"`
//	C Identify  `json:"clientID"`
//}

type Result struct {
	
} 

type Reply struct {
	T TimeStamp   `json:"timestamp"`
	C Identify    `json:"clientID"`
	I Identify    `json:"replicaID"`
	R Result      `json:"result"`
}

type PrePrepare struct {
	N Sequence     `json:"sequence"`
	D string       `json:"digest"`
	M []*Request `json:"message"`
}

type Prepare struct {
	N Sequence    `json:"sequence"`
	D string      `json:"digest"`
	I Identify    `json:"replicaID"`
}

type Commit struct {
	N Sequence `json:"sequence"`
	D string   `json:"digest"`
	I Identify `json:"replicaID"`
}

type Verify struct {
	N Sequence `json:"sequence"`
	D string   `json:"digest"`
	I Identify `json:"replicaID"`
}

type Verified struct {
	N Sequence `json:"sequence"`
	D string   `json:"digest"`
	I Identify `json:"replicaID"`
}

type CheckPoint struct {
	N Sequence `json:"sequence"`
	D string	`json:"digest"`
	I Identify `json:"replicaID"`
}

func NewBufferMsg(op Operation, t TimeStamp, id Identify, reqnum int) ([]byte, *BufferReq) {
	msg := &BufferReq{
		O: op,
		T: t,
		C: id,
		N: reqnum,
	}
	content, err := json.Marshal(msg)
	if err != nil {
		return nil, nil
	}
	return content, msg
}

func NewPreprepareMsg(seq Sequence, batch []*Request) ([]byte, *PrePrepare, string, error) {
	d, err := Digest(batch)
	if err != nil {
		return []byte{}, nil, "", nil
	}
	prePrepare := &PrePrepare{
		N: seq,
		D: d,
		M: batch,
	}
	ret, err := json.Marshal(prePrepare)
	if err != nil {
		return []byte{}, nil, "", nil
	}
	return ret, prePrepare, d, nil
}

// return byte, prepare, error
func NewPrepareMsg(id Identify, msg *PrePrepare) ([]byte, *Prepare, error) {
	prepare := &Prepare{
		N: msg.N,
		D: msg.D,
		I: id,
	}
	content, err := json.Marshal(prepare)
	if err != nil {
		return []byte{}, nil, err
	}
	return content, prepare, nil
}

// return byte, commit, error
func NewCommitMsg(id Identify, msg *Prepare) ([]byte, *Commit, error) {
	commit := &Commit{
		N: msg.N,
		D: msg.D,
		I: id,
	}
	content, err := json.Marshal(commit)
	if err != nil {
		return []byte{}, nil, err
	}
	return content, commit, nil
}

func NewVerifyMsg(id Identify, msg *Commit) ([]byte, *Verify, error) {
	verify := &Verify{
		N: msg.N,
		D: msg.D,
		I: id,
	}
	content, err := json.Marshal(verify)
	if err != nil {
		return []byte{}, nil, err
	}
	return content, verify, nil
}

func NewVerifiedMsg(id Identify, msg *Verify) ([]byte, *Verified, error) {
	verified := &Verified{
		N: msg.N,
		D: msg.D,
		I: id,
	}
	content, err := json.Marshal(verified)
	if err != nil {
		return []byte{}, nil, err
	}
	return content, verified, nil
}

func GroupSequenceString(grp string, seq Sequence) string {
	// TODO need better method
	seqStr := strconv.Itoa(int(seq))
	grpStr := grp
	seqLen := 4 - len(seqStr)
	grpLen := 28 - len(grpStr)
	// high 4  for viewStr
	for i := 0; i < seqLen; i++ {
		grpStr = "0" + grpStr
	}
	// low  28 for seqStr
	for i := 0; i < grpLen; i++ {
		seqStr = "0" + seqStr
	}
	return grpStr + seqStr
}
