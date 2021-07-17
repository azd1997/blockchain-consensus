/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 20:39
* @Description: 节点信息
***********************************************************************/

package defines

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
)

//// Peer 节点信息
//type Peer struct {
//	ID string
//	Addr string
//	Data []byte
//}

// PeerAttr 节点属性 正常、断连、恶意
type PeerAttr uint8

const (
	PeerAttr_Normal       PeerAttr = 0
	PeerAttr_Disconnected PeerAttr = 1
	PeerAttr_Malicious    PeerAttr = 2
)

// PeerRole 节点角色：病人、医生、医院、研究机构
// TODO：节点角色的话允许外部自定义，并注册处理函数，从而限制各个角色的行为
// 角色信息被编码进id字段，无需显示指定
type PeerRole uint8

const (
	PeerRole_Patient  PeerRole = 1
	PeerRole_Hospital PeerRole = 2
)

// PeerDuty 节点职责: 普通、种子、工人（共识节点）
type PeerDuty uint8

// 预定义的三种节点职责：None/Seed/Peer
const (
	PeerDuty_None PeerDuty = 0
	PeerDuty_Seed PeerDuty = 1

	// PeerDuty_Peer 对等节点，承担共识节点的责任
	PeerDuty_Peer PeerDuty = 2
)

func (duty PeerDuty) String() string {
	switch duty {
	case PeerDuty_None:
		return "None"
	case PeerDuty_Peer:
		return "Peer"
	case PeerDuty_Seed:
		return "Seed"
	default:
		return "Unknown"
	}
}

// PeerInfo 节点信息
type PeerInfo struct {
	Id   string
	Addr string
	Duty PeerDuty
	Attr PeerAttr
	Data []byte	// json编码你的一个map[string]string  {"alias":"Eiger"}
}

// String
func (pi *PeerInfo) String() string {
	b, err := json.Marshal(pi)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// Encode 编码
func (pi *PeerInfo) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(pi)
	return buf.Bytes(), err
}

// Decode 解码
// pi := new(PeerInfo)
func (pi *PeerInfo) Decode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	return dec.Decode(pi)
}
