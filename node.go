/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/7/20 9:36 PM
* @Description: Node的定义
***********************************************************************/

package bcc

import (
	"github.com/azd1997/blockchain-consensus/bnet"
	"github.com/azd1997/blockchain-consensus/requires"
)

/*
	共识集群内的节点概念：Node。

	任何一个节点与对端节点连接，都需要将自己作为客户端、对方作为服务器进行连接，也就是需要产生若干个连接Conn.
	对于节点自身而言，需要记录自己对外的连接和其他节点对自己的连接。

	要注意的是：任何情况下，消息的发送(包括消息的回复)只会通过 “自己”->对端节点 间的单向连接发送。
	连接本身不一定是单向的，但我们只使用单向，因此所有的Conn都抽象为单向连接。

	也就是说，每个节点(Node)需要维护自己连向其他节点的连接，以及其他节点连向自己的连接

	对于外部调用来说，应该像这样调用：
	type Node struct {
		isWorker bool	// 标记是否是共识节点
		server *bcc.Server
	}

	if isWorker {

	}
*/

// Node 节点服务器
type Node struct {

	// 节点ID，与账户共用一个ID
	id string

	// 共识状态机
	css Consensus

	// 网络模块
	net *bnet.Net

	//
}

// NewNode 构建Node
func NewNode(id string, consensusType string,
	ln requires.Listener, dialer requires.Dialer) *Node {
	node := &Node{id:id}

	css := NewConsensus(consensusType)
	node.css = css
	cssin, cssout := css.InMsgChan(), css.OutMsgChan()

	netmod := bnet.NewNet(ln, dialer, cssin, cssout)


	return node
}

// Ok 检查Node是否非空，以及内部一些成员是否准备好
func (s *Node) Ok() bool {
	return s != nil && s.css != nil && s.net != nil
}

// IsWorker 判断该Node是否是共识节点
// 这将通过id的第一个字节判断
func (s *Node) IsWorker() bool {
	// TODO
	return true
}

// ID 获取Node的唯一标识
func (s *Node) ID() string {
	return s.id
}

//