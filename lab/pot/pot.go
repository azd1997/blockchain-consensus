/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/14/20 10:21 AM
* @Description: 测试pot共识运行
***********************************************************************/

package pot

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
	"github.com/azd1997/blockchain-consensus/modules/consensus"
	"github.com/azd1997/blockchain-consensus/modules/ledger"
	"github.com/azd1997/blockchain-consensus/modules/peerinfo"
	"github.com/azd1997/blockchain-consensus/requires"
)

// 配置内容
var configString = `
[account]
id = "id1111"
duty = "seed"
addr = "127.0.0.1:8099"

[consensus]
type = "pot"

[pot]
tick_ms = 500

[pow]

[raft]

[store]
engine = "badger"
database = "./data"

[bnet]
protocol = "btcp"
addr = "127.0.0.1:8099"

# seeds信息对， id-addr
[seeds]
"seed1" = "seed1-addr"
"seed2" = "seed2-addr"

# peers信息对， id-addr.  用于手动补充可信任的节点信息
[peers]
"peer1" = "peer1-addr"
"peer2" = "peer2-addr"
`

// 定义使用pot的node结构体
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
	pit peerinfo.Pit

	// 共识状态机
	pot consensus.Consensus

	// 网络模块
	net bnet.BNet

	// 日志输出目的地
	LogDest string
}

// NewNode 构建Node
func NewNode(
	id string, duty defines.PeerDuty, // 账户配置
	addr string,
	seeds map[string]string, //预配置的种子节点
	peers map[string]string, // 预配置的共识节点
) (*Node, error) {

	node := &Node{
		id:   id,
		duty: duty,
		addr: addr,
	}

	// 构建bc
	bc, err := ledger.NewLedger("simplechain", id)
	if err != nil {
		return nil, err
	}
	err = bc.Init()
	if err != nil {
		return nil, err
	}
	node.bc = bc

	// 构建节点表
	pit, err := peerinfo.NewPit("simplepit", id)
	if err != nil {
		return nil, err
	}
	err = pit.Init()
	if err != nil {
		return nil, err
	}
	node.pit = pit
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

	msgchan := make(chan []byte, 100)

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
	pm, err := consensus.NewConsensus("pot", id, duty, pit, bc, netmod, msgchan)
	if err != nil {
		return nil, err
	}
	err = pm.Init()
	if err != nil {
		return nil, err
	}
	node.pot = pm

	return node, nil
}
