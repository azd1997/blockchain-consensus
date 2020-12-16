/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/7/20 9:36 PM
* @Description: Node的定义
***********************************************************************/

package bcc

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
	"github.com/azd1997/blockchain-consensus/modules/peerinfo"
	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/utils/log"
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

// Option Node所需的选项
type Option struct {
}

// PrepareOption 准备Option
//func PrepareOption(configReader io.Reader) (*Option, error) {
//	// 解析配置
//	tc, err := config.ParseConfig(configReader)
//	if err != nil {
//		return nil, err
//	}
//
//	、、
//}

// Node 节点服务器
type Node struct {

	// 节点ID，与账户共用一个ID
	id   string
	duty defines.PeerDuty
	addr string

	// kv 存储
	kv requires.Store
	// bc 区块链（相当于日志持久器）
	bc requires.BlockChain

	// 节点信息表
	pit *peerinfo.PeerInfoTable

	// 共识状态机
	css Consensus

	// 网络模块
	net *bnet.Net

	// 日志输出目的地
	LogDest log.LogDest
}

// NewNode 构建Node
func NewNode(
	id string, duty defines.PeerDuty, // 账户配置
	consensusType string, // 共识配置
	ln requires.Listener, dialer requires.Dialer, // 网络配置
	kv requires.Store, bc requires.BlockChain, // 外部依赖
	logdest log.LogDest, // 日志配置
) (*Node, error) {

	node := &Node{
		id: id,
		kv: kv,
		bc: bc,
	}

	// 构建节点表
	pit := peerinfo.NewPeerInfoTable(kv)
	err := pit.Init()
	if err != nil {
		return nil, err
	}
	node.pit = pit

	// 构建共识状态机
	css, err := NewConsensus(consensusType)
	if err != nil {
		return nil, err
	}
	node.css = css
	cssin, cssout := css.InMsgChan(), css.OutMsgChan()

	// 构建网络模块
	opt := &bnet.Option{
		Listener: ln,
		Dialer:   dialer,
		MsgIn:    cssout,
		MsgOut:   cssin,
		LogDest:  logdest,
		Pit:      pit,
	}
	netmod, err := bnet.NewNet(opt)
	if err != nil {
		return nil, err
	}
	node.net = netmod

	return node, nil
}

// Init 初始化
func (s *Node) Init() error {
	// 准备好PeerInfoTable
	err := s.pit.Init()
	if err != nil {
		return err
	}

	// 网络模块初始化
	err = s.net.Init()
	if err != nil {
		return err
	}

	// 共识模块初始化
	err = s.css.Init()
	if err != nil {
		return err
	}

	return nil
}

// Ok 检查Node是否非空，以及内部一些成员是否准备好
func (s *Node) Ok() bool {
	return s != nil && s.css != nil && s.net != nil && s.bc != nil
}

// IsConsensusNode 判断该Node是否是共识节点
func (s *Node) IsConsensusNode() bool {
	return s.duty == defines.PeerDuty_Peer
}

// ID 获取Node的唯一标识
func (s *Node) ID() string {
	return s.id
}

//
