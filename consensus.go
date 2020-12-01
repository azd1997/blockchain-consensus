/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:19
* @Description: Consensus接口
***********************************************************************/

package bcc

import "github.com/azd1997/blockchain-consensus/defines"

// Consensus 共识接口。事实上一个Consensus实例代表一个基于该共识协议的共识节点
type Consensus interface {

	// Init 初始化(运行)。 New之后需要Init
	Init() error

	// Close 关闭状态机服务，执行一些必要的清理工作
	Close() error

	// 共识节点必需有处理各类消息的能力
	handleMsg(msg *defines.Message) error

	// 状态机循环，负责状态的切换
	stateMachineLoop()

	// 消息处理循环
	msgHandleLoop()

	// MsgChannel 对于Consensus的上层来说，需要调用该函数，
	// 得到消息channel，根据该channel拿消息去发送到网络中
	// TODO: 发送的结果是成功还是失败？状态机需不需要考虑？
	OutMsgChan() chan *defines.MessageWithError

	// 接收消息的channel，需要将该chan移交给网络模块去写消息
	// Consensus模块在内部循环读该chan，处理Message
	InMsgChan() chan *defines.Message
}

// NewConsensus 新建一个共识状态机
func NewConsensus(typ string) (Consensus, error) {
	return nil, nil
}
