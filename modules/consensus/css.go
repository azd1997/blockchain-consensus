/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 2:15 PM
* @Description: The file is for
***********************************************************************/

package consensus

import (
	"errors"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
	"github.com/azd1997/blockchain-consensus/modules/consensus/pot"
	"github.com/azd1997/blockchain-consensus/modules/pitable"
	"github.com/azd1997/blockchain-consensus/requires"
)

type ConsensusType uint8

const (
	ConsensuType_PoT ConsensusType = iota
	ConsensuType_PoW
	ConsensuType_PBFT
	ConsensuType_Raft
)

// Consensus 共识接口。事实上一个Consensus实例代表一个基于该共识协议的共识节点
type Consensus interface {
	Init() error  // Init 初始化(运行)。 New之后需要Init
	Inited() bool // Inited 判断是否初始化好
	Ok() bool     // Ok 检查Net所依赖的对象是否初始化好
	Close() error // Close 关闭状态机服务，执行一些必要的清理工作
	Closed() bool

	/*不可调用，只是为了约束Consensus该有的实现*/

	// HandleMsg [不可调用] 共识节点必需有处理各类消息的能力
	HandleMsg(msg *defines.Message) error
	// StateMachineLoop [不可调用] 状态机循环，负责状态的切换 go
	StateMachineLoop()
	// MsgHandleLoop [不可调用] 消息处理循环 go
	MsgHandleLoop()
	// SetMsgInChan [不建议调用] 设置接收消息的channel，需要将该chan移交给网络模块去写消息
	// Consensus模块在内部循环读该chan，处理Message
	SetMsgInChan(bus chan *defines.Message)
}

// New 新建一个共识状态机
func New(typ ConsensusType,
	id string, duty defines.PeerDuty,
	pit pitable.Pit, bc requires.BlockChain,
	net bnet.BNet, msgchan chan *defines.Message) (Consensus, error) {

	switch typ {
	case ConsensuType_PoT:
		return pot.New(id, duty, pit, bc, net, msgchan)
	default:
		return nil, errors.New("unknown consensus type")
	}
}
