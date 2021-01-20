/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 2:23 PM
* @Description: The file is for
***********************************************************************/

package bnet

import (
	"errors"
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/modules/bnet/budp"
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

	Send(raddr string, msg []byte) error // 向某人发送消息
	MsgOut() chan []byte                 // 要求结构体内部有一个消息总线chan，用于将得到的消息传输出来
	RecvLoop()                           // go RecvLoop() 消息塞入msgout
}

// NewBNet
func NewBNet(id string, network string, addr string, msgchan chan []byte) (BNet, error) {

	logger := log.NewLogger("NET", id)

	switch network {
	case "udp":
		return budp.NewUDPNet(id, addr, logger, msgchan)
	case "tcp":
		return nil, nil
	default:
		return nil, errors.New("unknown network protocol")
	}
}
