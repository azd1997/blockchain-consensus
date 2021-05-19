/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/7/20 9:36 PM
* @Description: Node的定义
***********************************************************************/

package bcc

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
	"github.com/azd1997/blockchain-consensus/modules/bnet/conn_net"
	"github.com/azd1997/blockchain-consensus/modules/bnet/budp"
	"github.com/azd1997/blockchain-consensus/modules/consensus"
	"github.com/azd1997/blockchain-consensus/modules/consensus/pot"
	"github.com/azd1997/blockchain-consensus/modules/ledger"
	"github.com/azd1997/blockchain-consensus/modules/ledger/simplechain"
	"github.com/azd1997/blockchain-consensus/modules/pitable"
	"github.com/azd1997/blockchain-consensus/modules/pitable/memorypit"
	"github.com/azd1997/blockchain-consensus/modules/pitable/simplepit"
	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/test"
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
// 众多模块的集合体
type Node struct {

	// 节点ID，与账户共用一个ID
	id   string
	duty defines.PeerDuty
	addr string

	// msgBus 消息总线
	// net -> pot; txmaker -> pot
	msgBus chan *defines.Message

	// kv 存储
	kv requires.Store
	// bc 区块链（相当于日志持久器）
	bc requires.BlockChain
	// pit 节点信息表
	pit pitable.Pit
	// css 共识状态机
	css consensus.Consensus
	// net 网络模块
	net bnet.BNet
	// tv 交易解释器（业务持有）与验证器（业务自定义，传入Node）
	tv requires.Validator

	// 日志输出目的地
	//LogDest string
}

// DefaultNode 所有模块都采用默认提供的组件
//func DefaultNode(id string, duty defines.PeerDuty, addr string) (*Node, error) {
//	node := &Node{
//		id:   id,
//		duty: duty,
//		addr: addr,
//
//		kv:  nil,
//		bc:  nil,
//		pit: nil,
//		css: nil,
//		net: nil,
//		tv:  nil,
//	}
//
//	kv := test.NewStore()
//	bc, err := simplechain.NewBlockChain(id)
//	if err != nil {
//		return nil, err
//	}
//	pit, err := simplepit.NewSimplePit(id)
//	if err != nil {
//		return nil, err
//	}
//	bus := make(chan []byte, 1000)
//	netm, err := budp.NewUDPNet(id, addr, nil, nil)
//	if err != nil {
//		return nil, err
//	}
//	css, err := pot.New(id, duty, pit, bc, netm, bus)
//}

// NewNode 构建Node
func NewNode(
	id string, duty defines.PeerDuty, // 账户配置
	addr string,
	initedPit pitable.Pit, initedBc requires.BlockChain,
	msgbus chan *defines.Message, initedNet bnet.BNet, initedCss consensus.Consensus,
	seeds map[string]string, //预配置的种子节点
	peers map[string]string, // 预配置的共识节点
) (*Node, error) {

	node := &Node{
		id:   id,
		duty: duty,
		addr: addr,
	}

	// 构建bc
	if initedBc == nil {
		bc, err := ledger.New("simplechain", id)
		if err != nil {
			return nil, err
		}
		err = bc.Init()
		if err != nil {
			return nil, err
		}
		initedBc = bc
	}
	node.bc = initedBc

	// 构建节点表
	if initedPit == nil {
		pit, err := pitable.New("simplepit", id)
		if err != nil {
			return nil, err
		}
		err = pit.Init()
		if err != nil {
			return nil, err
		}
		initedPit = pit
	}
	node.pit = initedPit
	// 预配置节点表
	if err := node.pit.Set(&defines.PeerInfo{
		Id:   node.id,
		Addr: node.addr,
		Attr: 0,
		Duty: node.duty,
		Data: nil,
	}); err != nil {
		return nil, err
	}
	node.pit.AddPeers(peers)
	node.pit.AddSeeds(seeds)

	// 消息总线
	if msgbus == nil {
		msgbus = make(chan *defines.Message, 100)
	}

	// 构建网络模块
	netmod, err := bnet.NewBNet(id, "udp", addr, msgchan)
	if err != nil {
		return nil, err
	}
	err = netmod.Init()
	if err != nil {
		return nil, err
	}
	node.net = netmod

	// 构建共识状态机
	pm, err := pot.New(id, duty, pit, bc, netmod, msgchan)
	if err != nil {
		return nil, err
	}
	err = pm.Init()
	if err != nil {
		return nil, err
	}
	node.css = pm

	return node, nil
}

//// NewNode 构建Node
//func NewNode(
//	id string, duty defines.PeerDuty, // 账户配置
//	consensusType string, // 共识配置
//	ln requires.Listener, dialer requires.Dialer, // 网络配置
//	kv requires.Store, bc requires.BlockChain, // 外部依赖
//	logdest log.LogDest, // 日志配置
//) (*Node, error) {
//
//	node := &Node{
//		id: id,
//		kv: kv,
//		bc: bc,
//	}
//
//	// 构建节点表
//	pit := memorypit.NewPeerInfoTable(kv)
//	err := pit.Init()
//	if err != nil {
//		return nil, err
//	}
//	node.pit = pit
//
//	// 构建共识状态机
//	css, err := NewConsensus(consensusType)
//	if err != nil {
//		return nil, err
//	}
//	node.css = css
//	cssin, cssout := css.InMsgChan(), css.OutMsgChan()
//
//	// 构建网络模块
//	opt := &conn_net.Option{
//		Listener: ln,
//		Dialer:   dialer,
//		MsgIn:    cssout,
//		MsgOut:   cssin,
//		LogDest:  logdest,
//		Pit:      pit,
//	}
//	netmod, err := conn_net.NewNet(opt)
//	if err != nil {
//		return nil, err
//	}
//	node.net = netmod
//
//	return node, nil
//}

//// Init 初始化
//func (s *Node) Init() error {
//	// 准备好PeerInfoTable
//	err := s.pit.Init()
//	if err != nil {
//		return err
//	}
//
//	// 网络模块初始化
//	err = s.net.Init()
//	if err != nil {
//		return err
//	}
//
//	// 共识模块初始化
//	err = s.css.Init()
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

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
