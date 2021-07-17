/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 2:23 PM
* @Description: The file is for
***********************************************************************/

package bnet

import (
	"errors"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/modules/bnet/btcp"
	"github.com/azd1997/blockchain-consensus/modules/bnet/btcp_dual"
	"github.com/azd1997/blockchain-consensus/modules/bnet/budp"
)


type NetType uint8

const (
	NetType_bUDP NetType = iota
	NetType_bTCP
	NetType_bTCP_Dual
)

// 考虑到具体的BNet实现必然是统一网络协议的，所以，没必要在发送时使用net.Addr，直接使用“ip:port”的string

type BNet interface {
	ID() string
	Addr() string
	Network() string

	Init() error
	Inited() bool
	Ok() bool // Ok 检查Net所依赖的对象是否初始化好
	Close() error
	Closed() bool

	Send(id string, raddr string, msg *defines.Message) error // 向某人发送消息
	// Broadcast 这个无差别广播是基于已经建立的连接进行广播，会对所传入的msg，替换to并重新签名
	Broadcast(msg *defines.Message) error // 无差别广播
	// IDAddrs 获取当前net模块内所有可用的节点。[2]string{ID,Addr}
	IDAddrs() [][2]string

	// SetMsgOutChan [不建议调用]
	SetMsgOutChan(bus chan *defines.Message) // 给结构体设置一个消息总线chan，用于将得到的消息传输出来
	// RecvLoop [不可调用]
	RecvLoop() // go RecvLoop() 消息塞入msgout

	DisplayAllConns(brief bool) string // 展示所有连接信息
}

// NewBNet
func NewBNet(id string, network NetType, addr string,
		msgchan chan *defines.Message) (BNet, error) {

	logger := log.NewLogger("NET", id)

	switch network {
	case NetType_bUDP:
		return budp.New(id, addr, logger, msgchan)
	case NetType_bTCP:
		return btcp.New(id, addr, logger, msgchan)
	case NetType_bTCP_Dual:
		return btcp_dual.New(id, addr, logger, msgchan)
	default:
		return nil, errors.New("unknown network protocol")
	}
}
