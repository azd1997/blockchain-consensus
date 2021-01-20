/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 2:15 PM
* @Description: The file is for
***********************************************************************/

package consensus

import (
	"errors"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
	"github.com/azd1997/blockchain-consensus/modules/peerinfo"
	"github.com/azd1997/blockchain-consensus/requires"
	"strings"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/consensus/pot"
)

// Consensus 共识接口。事实上一个Consensus实例代表一个基于该共识协议的共识节点
type Consensus interface {
	Init() error // Init 初始化(运行)。 New之后需要Init
	Inited() bool
	Close() error // Close 关闭状态机服务，执行一些必要的清理工作

	// 共识节点必需有处理各类消息的能力
	HandleMsg(msg *defines.Message) error
	// 状态机循环，负责状态的切换 go
	StateMachineLoop()
	// 消息处理循环 go
	MsgHandleLoop()

	// MsgChannel 对于Consensus的上层来说，需要调用该函数，
	// 得到消息channel，根据该channel拿消息去发送到网络中
	// TODO: 发送的结果是成功还是失败？状态机需不需要考虑？
	//OutMsgChan() chan *defines.MessageWithError

	// 接收消息的channel，需要将该chan移交给网络模块去写消息
	// Consensus模块在内部循环读该chan，处理Message
	MsgInChan() chan *defines.Message

	LocalTxInChan() chan *defines.Transaction // 本地传入的交易的chan
}

// NewConsensus 新建一个共识状态机
func NewConsensus(typ string,
	id string, duty defines.PeerDuty,
	pit peerinfo.Pit, bc requires.BlockChain,
	net bnet.BNet, msgchan chan []byte) (Consensus, error) {

		typ = strings.ToLower(typ) // 支持pot, Pot等大小写
	switch typ {
	case "pot":
		return pot.New(id, duty, pit, bc, net, msgchan)
	default:
		return nil, errors.New("unknown consensus type")
	}
}
