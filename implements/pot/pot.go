/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:49
* @Description: Node结构体就是实现PoT共识的类，其具体职责包含了输入输出处理与状态转换
***********************************************************************/

package pot

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
	"sync/atomic"
	"time"
)


const (
	// 逻辑时钟每个滴答就是500ms
	TickMs = 500
)

// Node pot节点
type Node struct {

	id string	// 账户、节点、客户端共用一个ID

	latest bool	// 本节点是否追上系统最新进度

	peers map[string]*defines.Peer	// 共识节点表

	state StateType		// 状态状态

	msgChan chan *defines.Message	// 对外提供的消息通道

	ticker time.Ticker	// 滴答器，每次滴答时刻都需要根据当前的状态变量确定状态该如何变更

	proofs map[string]*Proof	// 收集的其他共识节点的证明进度
	winner string
	maybeNewBlock *defines.Block
	waitingNewBlock *defines.Block	// 等待的新区块，没等到就一直是nil

	blocksCache map[string]*defines.Block	// 同步到本机节点的区块，但尚未排好序的。也就是序列化没有接着本地最高区块后边的

	////////////////////////// 本地依赖 /////////////////////////

	txPool requires.TransactionPool

	////////////////////////// 本地依赖 /////////////////////////
}

func New() *Node {
	return &Node{}
}

//////////////////////////// 实现接口 ///////////////////////////

// 状态切换循环
func (n *Node) stateMachineLoop() {
	for {
		select {
		case <-n.ticker.C:
			// 根据当前状态来处理此滴答消息

			state := StateType(atomic.LoadUint32((*uint32)(&n.state)))		// state的读写使用atomic
			switch state {
			case StateType_NotReady:
				// 没准备好，啥也不干，等区块链同步


				// 如果追上进度了则切换状态为ReadyCompete
				if n.latest {
					atomic.StoreUint32((*uint32)(&n.state), StateType_ReadyCompete)
				} else {
					// 否则请求快照数据
					n.requestBlocks()
				}

			case StateType_ReadyCompete:
				// 当前是ReadyCompete，则状态切换为Competing

				// 状态切换
				atomic.StoreUint32((*uint32)(&n.state), StateType_Competing)
				// 发起竞争（广播证明消息）
				n.broadcastProof()

			case StateType_Competing:
				// 当前是Competing，则状态切换为CompetingEnd，并判断竞赛结果，将状态迅速切换为Winner或Loser

				// 状态切换
				atomic.StoreUint32((*uint32)(&n.state), StateType_CompeteOver)
				// 判断竞赛结果，状态切换
				if n.winner == n.id {	// 自己胜出
					atomic.StoreUint32((*uint32)(&n.state), StateType_CompeteWinner)
					// 广播新区块
					n.broadcastBlock(n.maybeNewBlock)
				} else {	// 别人胜出
					atomic.StoreUint32((*uint32)(&n.state), StateType_CompeteLoser)
					// 等待新区块(“逻辑上”的等待，代码中并不需要wait)
				}

			case StateType_CompeteOver:
				// 正常来说，tick时不会是恰好CompeteOver而又没确定是Winner/Loser
				// 所以暂时无视
			case StateType_CompeteWinner:
				// Winner来说的话，立马广播新区块，广播结束后即切换为Ready
				// 所以不太可能tick时状态为Winner
				// 暂时无视
			case StateType_CompeteLoser:
				// Loser等待新区块，接收到tick说明还没得到新区块
				// 状态切换为Ready

				atomic.StoreUint32((*uint32)(&n.state), StateType_ReadyCompete)
				// 发起竞争（广播证明消息）
				n.broadcastProof()
			}
		}
	}
}

// 处理外界消息输入和内部消息
func (n *Node) HandleMsg(msg *defines.Message) {
	// 检查消息签名，防止被篡改
	// TODO

	state := atomic.LoadUint32((*uint32)(&n.state))
	switch state {
	case StateType_NotReady:
		n.handleMsgWhenNotReady(msg)
	case StateType_ReadyCompete:
		n.handleMsgWhenReadyCompete(msg)
	case StateType_Competing:
		n.handleMsgWhenCompeting(msg)
	case StateType_CompeteOver:
		n.handleMsgWhenCompeteOver(msg)
	case StateType_CompeteWinner:
		n.handleMsgWhenCompeteWinner(msg)
	case StateType_CompeteLoser:
		n.handleMsgWhenCompeteLoser(msg)
	}
}

// 向外界提供消息通道
func (n *Node) MsgChannel() <-chan *defines.Message {
	return n.msgChan
}


/////////////////////////// 处理消息 /////////////////////////

// 目前的消息只有类：区块消息、证明消息、
// 快照消息（用于NotReady的节点快速获得区块链部分，目前直接用区块消息代表）。

func (n *Node) handleMsgWhenNotReady(msg *defines.Message) {
	switch msg.Type {
	case defines.MessageType_Data:
		// 该阶段只能处理block消息

		// 根据EntryType来处理
		for _, ent := range msg.Entries {
			switch ent.EntryType {
			case defines.EntryType_Block:
				// 检查区块本身格式的正确性

				// 如果其序号 = 本地的index+1，那么检查其有效性



				// 如果序号是不连续的，暂时先保留在blocksCache


			default:	// 其他类型则忽略

			}
		}


	case defines.MessageType_Req:
		// 该阶段不能处理Req消息
	default:
		// 报错打日志
	}
}

func (n *Node) handleMsgWhenReadyCompete(msg *defines.Message) {
	switch msg.Type {
	case defines.MessageType_Data:
		// ReadyCompete状态下不允许接收区块和证明

	case defines.MessageType_Req:

		// 检查请求内容，查看
		// 根据EntryType来处理
		for _, req := range msg.Reqs {
			switch req.Type {
			case defines.RequestType_Blocks:
				// 检查请求本身格式的正确性

				// 如果检查请求的区块范围有效，回复区块


			case defines.RequestType_Neighbors:


			default:	// 其他类型则忽略

			}
		}

	default:
		// 报错打日志
	}
}

func (n *Node) handleMsgWhenCompeting(msg *defines.Message) {
	switch msg.Type {
	case defines.MessageType_Data:
		// ReadyCompete状态下不允许接收区块，允许接收证明

		if len(msg.Entries) != 1 {
			// 报错 TODO
			return
		}

		switch msg.Entries[0].Type {
		case defines.EntryType_Proof:
			// 解码
			var proof Proof
			err := proof.Decode()
			if err != nil {
				// 报错

				return
			}

			// 如果证明信息有效，用以更新本地winner
			n.proofs[msg.From] = &proof
			if proof.GreaterThan(n.proofs[n.winner]) {
				n.winner = msg.From
			}


	default:	// 其他类型则忽略

	}

	case defines.MessageType_Req:

		// 检查请求内容，查看
		// 根据EntryType来处理
		for _, ent := range msg.Entries {
			switch ent.Type {
			case defines.EntryType_Block:
				// 检查请求本身格式的正确性

				// 如果检查请求的区块范围有效，回复区块


			default:	// 其他类型则忽略

			}
		}

	default:
		// 报错打日志
	}
}

func (n *Node) handleMsgWhenCompeteOver(msg *defines.Message) {

}

func (n *Node) handleMsgWhenCompeteWinner(msg *defines.Message) {

}

func (n *Node) handleMsgWhenCompeteLoser(msg *defines.Message) {

}

////////////////////////// 主动操作 //////////////////////////

// send指向网络中（或者说外部依赖的网络模块）发送消息。 注意本地消息不要通过该方法使用
// 这个发送消息可以是单播也可以是多播，具体看
func (n *Node) send(msg *defines.Message) {
	n.msgChan <- msg
}

//// broadcast
//func (n *Node) broadcast(msg *defines.Message) {
//
//}

// 广播证明
func (n *Node) broadcastProof() {
	// 首先需要获取证明
	nb := n.txPool.GenBlock()
	n.maybeNewBlock = nb
	proof := &Proof{
		//Base:nb.PrevHash,
		//BaseIndex:nb.Index-1,
		TxsNum:uint64(len(nb.Txs)),
		TxsMerkle:nb.Merkle,
	}

	proofBytes := proof.Encode()

	entry := &defines.Entry{
		Base:nb.PrevHash,
		BaseIndex:uint64(nb.Index-1),
		Type:defines.EntryType_Proof,
		Data:proofBytes,
	}

	// 广播
	for to := range n.peers {
		msg := &defines.Message{
			Version:defines.CodeVersion,
			Type:defines.MessageType_Data,
			From:n.id,
			To:to,
			Sig:nil,	// TODO
			Entries:[]*defines.Entry{entry},
		}
		n.send(msg)
	}

}

// 广播区块
func (n *Node) broadcastBlock(b *defines.Block) {

	blockBytes := b.Encode()

	entry := &defines.Entry{
		Base:b.PrevHash,
		BaseIndex:uint64(b.Index-1),
		Type:defines.EntryType_Block,
		Data:blockBytes,
	}

	// 广播
	for to := range n.peers {
		msg := &defines.Message{
			Version:defines.CodeVersion,
			Type:defines.MessageType_Data,
			From:n.id,
			To:to,
			Sig:nil,	// TODO
			Entries:[]*defines.Entry{entry},
		}
		n.send(msg)
	}
}

// 暂时是直接向一个节点发送区块请求，但最好后面改成随机选举三个去请求
func (n *Node) requestBlocks() {
	// TODO
}

// 向节点回应其请求的区块
func (n *Node) responseBlocks(to string, blocks ...*defines.Block) {
	l := len(blocks)
	if l == 0 {
		return
	}

	entries := make([]*defines.Entry, l)
	for i:=0; i<l; i++ {
		entries[i] = &defines.Entry{
			BaseIndex: blocks[i].Index-1,
			Base: blocks[i].PrevHash,
			Type:defines.EntryType_Block,
			Data:blocks[i].Encode(),
		}
	}

	msg := &defines.Message{
		Version:defines.CodeVersion,
		Type:defines.MessageType_Data,
		From:n.id,
		To:to,
		Sig:nil,	// TODO
		Entries:entries,
	}

	n.send(msg)
}