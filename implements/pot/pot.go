/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:49
* @Description: Pot结构体就是实现PoT共识的类，其只负责逻辑处理与状态转换
***********************************************************************/

package pot

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/peerinfo"
	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/utils/log"
)

const (
	// 逻辑时钟每个滴答就是500ms
	TickMs = 500

	// 默认channel长度
	DefaultMsgChanLen = 100

	// 模块名
	Module_Css = "CSS"
)

// Pot pot节点
type Pot struct {
	id string // 账户、节点、客户端共用一个ID

	latest bool // 本节点是否追上系统最新进度

	state StateType // 状态状态

	processes map[string]*defines.Process
	processesLock *sync.RWMutex

	// 对外提供的消息通道
	msgin chan *defines.Message
	msgout chan *defines.Message

	ticker time.Ticker // 滴答器，每次滴答时刻都需要根据当前的状态变量确定状态该如何变更

	// 节点信息表，传入的需要是已初始化好的
	pit *peerinfo.PeerInfoTable

	// proofs表可能会因为某些节点出现恶意行为而将其删除
	proofs          map[string]*Proof // 收集的其他共识节点的证明进度
	proofsLock *sync.RWMutex
	winner          string
	maybeNewBlock   *defines.Block
	waitingNewBlock *defines.Block // 等待的新区块，没等到就一直是nil

	blocksCache map[string]*defines.Block // 同步到本机节点的区块，但尚未排好序的。也就是序列化没有接着本地最高区块后边的
	blocksLock *sync.RWMutex

	////////////////////////// 本地依赖 /////////////////////////

	txPool requires.TransactionPool
	bc requires.BlockChain

	////////////////////////// 本地依赖 /////////////////////////

	*log.Logger
	done chan struct{}
}

// New 新建Pot
func New(id string, logdest log.LogDest, pit *peerinfo.PeerInfoTable) (*Pot, error) {
	p := &Pot{
		id:id,
		state:StateType_NotReady,
		msgin:make(chan *defines.Message, DefaultMsgChanLen),
		msgout: make(chan *defines.Message, DefaultMsgChanLen),
		done:make(chan struct{}),
		Logger:log.NewLogger(logdest, Module_Css, id),
	}

	if pit == nil {
		p.pit = peerinfo.Global()
	} else {
		if pit.Inited() {
			p.pit = pit
		} else {
			return nil, errors.New("PeerInfoTable should be inited")
		}
	}

	return p, nil
}

//////////////////////////// 实现接口 ///////////////////////////

// Init 初始化
func (p *Pot) Init() error {

	return nil
}

// Close 关闭
func (p *Pot) Close() error {
	close(p.done)
	return nil
}

// OutMsgChan 对外提供消息通道，用于数据向外传输
func (p *Pot) OutMsgChan() chan *defines.Message {
	return p.msgout
}

// InMsgChan 对外提供消息通道，用于数据向内传输
func (p *Pot) InMsgChan() chan *defines.Message {
	return p.msgin
}

// 状态切换循环
func (p *Pot) stateMachineLoop() {
	for {
		select {
		case <-p.done:
			p.Logf("stateMachineLoop: return ...\n")
			return
		case <-p.ticker.C:
			// 根据当前状态来处理此滴答消息

			state := StateType(atomic.LoadUint32((*uint32)(&p.state))) // state的读写使用atomic
			switch state {
			case StateType_NotReady:
				// 没准备好，啥也不干，等区块链同步

				// 如果追上进度了则切换状态为ReadyCompete
				if p.latest {
					atomic.StoreUint32((*uint32)(&p.state), StateType_ReadyCompete)
				} else {
					// 否则请求快照数据
					p.requestBlocks()
				}

			case StateType_ReadyCompete:
				// 当前是ReadyCompete，则状态切换为Competing

				// 状态切换
				atomic.StoreUint32((*uint32)(&p.state), StateType_Competing)
				// 发起竞争（广播证明消息）
				p.broadcastProof()

			case StateType_Competing:
				// 当前是Competing，则状态切换为CompetingEnd，并判断竞赛结果，将状态迅速切换为Winner或Loser

				// 状态切换
				atomic.StoreUint32((*uint32)(&p.state), StateType_CompeteOver)
				// 判断竞赛结果，状态切换
				if p.winner == p.id { // 自己胜出
					atomic.StoreUint32((*uint32)(&p.state), StateType_CompeteWinner)
					// 广播新区块
					p.broadcastNewBlock(p.maybeNewBlock)
				} else { // 别人胜出
					atomic.StoreUint32((*uint32)(&p.state), StateType_CompeteLoser)
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

				atomic.StoreUint32((*uint32)(&p.state), StateType_ReadyCompete)
				// 发起竞争（广播证明消息）
				p.broadcastProof()
			}
		}
	}
}

// 消息处理循环
func (p *Pot) msgHandleLoop() {
	var err error
	for {
		select {
		case <-p.done:
			return
		case msg := <- p.msgin:
			err = p.handleMsg(msg)
			if err != nil {
				p.Errorf("msgHandleLoop: handle msg(%v) fail: %s\n", msg, err)
			}
		}
	}
}

// 处理外界消息输入和内部消息
func (p *Pot) handleMsg(msg *defines.Message) error {
	// 检查消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	// 根据当前状态不同，执行不同的消息处理
	state := atomic.LoadUint32((*uint32)(&p.state))
	switch state {
	case StateType_NotReady:
		return p.handleMsgWhenNotReady(msg)
	case StateType_ReadyCompete:
		return p.handleMsgWhenReadyCompete(msg)
	case StateType_Competing:
		return p.handleMsgWhenCompeting(msg)
	case StateType_CompeteOver:
		return p.handleMsgWhenCompeteOver(msg)
	case StateType_CompeteWinner:
		return p.handleMsgWhenCompeteWinner(msg)
	case StateType_CompeteLoser:
		return p.handleMsgWhenCompeteLoser(msg)
	}
	return nil
}

/////////////////////////// 处理消息 /////////////////////////

// 目前的消息只有类：区块消息、证明消息、
// 快照消息（用于NotReady的节点快速获得区块链部分，目前直接用区块消息代表）。

func (p *Pot) handleMsgWhenNotReady(msg *defines.Message) error {
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_Data:
		// 该阶段只能处理block消息
		// 根据EntryType来处理
		for _, ent := range msg.Entries {
			switch ent.Type {
			case defines.EntryType_Blocks:
				return p.handleEntryBlocks(ent)
			default: // 其他类型则忽略
			}
		}
	case defines.MessageType_Req:
		// 该阶段不能处理Req消息
	default:
		// 报错打日志
	}
	return nil
}

// ReadyCompete状态下可以收集交易、新区块，以及区块和邻居请求
func (p *Pot) handleMsgWhenReadyCompete(msg *defines.Message) error {
	// 检查自身是否追上最新进度
	if err := p.checkLatestProcess(); err != nil {
		return err
	}
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_Data:
		// ReadyCompete状态下不允许接收区块和证明
		for i:=0; i<len(msg.Entries); i++ {
			ent := msg.Entries[i]
			switch ent.Type {
			case defines.EntryType_Blocks:
				return p.handleEntryBlocks(ent)
			case defines.EntryType_Proof:
			case defines.EntryType_NewBlock:
				return p.handleEntryNewBlock(ent)
			case defines.EntryType_Transaction:
				return p.handleEntryTransaction(ent)
			default:
				return errors.New("unknown entry type")
			}
		}
	case defines.MessageType_Req:
		// 检查请求内容，查看
		if len(msg.Entries) > 0 || len(msg.Reqs) == 0 {
			return errors.New("not a req msg")
		}
		for i:=0; i<len(msg.Reqs); i++ {
			req := msg.Reqs[i]
			switch req.Type {
			case defines.RequestType_Blocks:
				return p.handleRequestBlocks(req)
			case defines.RequestType_Neighbors:
				return p.handleRequestNeighbors(req)
			default: // 其他类型则忽略
				return errors.New("unknown req type")
			}
		}
	default:
		return errors.New("unknown msg type")
	}

	return nil
}

// Competing状态下msg处理
func (p *Pot) handleMsgWhenCompeting(msg *defines.Message) error {
	// 检查自身是否追上最新进度
	if err := p.checkLatestProcess(); err != nil {
		return err
	}
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_Data:
		// ReadyCompete状态下不允许接收区块，允许接收证明
		if len(msg.Entries) != 1 {
			return errors.New("recv proof only, but len(msg.Entries) != 1")
		}

		ent := msg.Entries[0]
		switch ent.Type {
		case defines.EntryType_Proof:
			return p.handleEntryProof(ent, msg.From)
		default: // 其他类型则忽略
		}

	case defines.MessageType_Req:
		// 检查请求内容，查看
		if len(msg.Entries) > 0 || len(msg.Reqs) == 0 {
			return errors.New("not a req msg")
		}
		// 根据RequestType来处理
		for _, req := range msg.Reqs {
			switch req.Type {
			case defines.RequestType_Blocks:
				return p.handleRequestBlocks(req)
			case defines.RequestType_Neighbors:
				return p.handleRequestNeighbors(req)
			default: // 其他类型则忽略
			}
		}

	default:
		return errors.New("unknown msg type")
	}
	return nil
}

// CompeteOver 状态
// 此状态极为短暂，该期间只能照常处理交易、邻居消息，其他消息不处理
func (p *Pot) handleMsgWhenCompeteOver(msg *defines.Message) error {
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_None:
		// 啥也不干
	case defines.MessageType_Data:
		for i:=0; i<len(msg.Entries); i++ {
			ent := msg.Entries[i]
			switch ent.Type {
			case defines.EntryType_Blocks:
			case defines.EntryType_Proof:
			case defines.EntryType_NewBlock:
			case defines.EntryType_Transaction:
			default:
				return errors.New("unknown entry type")
			}
		}
	case defines.MessageType_Req:
		for i:=0; i<len(msg.Reqs); i++ {
			req := msg.Reqs[i]
			switch req.Type {
			case defines.RequestType_Blocks:
			case defines.RequestType_Neighbors:
			default:
				return errors.New("unknown entry type")
			}
		}
	default:
		return errors.New("unknown msg type")
	}
	return nil
}

func (p *Pot) handleMsgWhenCompeteWinner(msg *defines.Message) error {
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_None:
		// 啥也不干
	case defines.MessageType_Data:
		for i:=0; i<len(msg.Entries); i++ {
			ent := msg.Entries[i]
			switch ent.Type {
			case defines.EntryType_Blocks:
			case defines.EntryType_Proof:
			case defines.EntryType_NewBlock:
			case defines.EntryType_Transaction:
			default:
				return errors.New("unknown entry type")
			}
		}
	case defines.MessageType_Req:
		for i:=0; i<len(msg.Reqs); i++ {
			req := msg.Reqs[i]
			switch req.Type {
			case defines.RequestType_Blocks:
			case defines.RequestType_Neighbors:
			default:
				return errors.New("unknown entry type")
			}
		}
	default:
		return errors.New("unknown msg type")
	}
	return nil
}

// CompeteLoser 状态
//
func (p *Pot) handleMsgWhenCompeteLoser(msg *defines.Message) error {
	// CompeteLoser状态下可能短暂落后其他节点
	//

	// 处理新区块、交易两种数据，处理Neighbors
	switch msg.Type {
	case defines.MessageType_None:
	case defines.MessageType_Data:
		// 处理所有Entry
		for i:=0; i<len(msg.Entries); i++ {
			ent := msg.Entries[i]
			switch ent.Type {
			case defines.EntryType_NewBlock:

			case defines.EntryType_Blocks:

			case defines.EntryType_Proof:
			case defines.EntryType_Transaction:

			default:
				return errors.New("unknown entry type")
			}
		}
	case defines.MessageType_Req:
		for i:=0; i<len(msg.Reqs); i++ {
			req := msg.Reqs[i]
			switch req.Type {
			case defines.RequestType_Blocks:
			case defines.RequestType_Neighbors:
			default:
				return errors.New("unknown entry type")
			}
		}
	default:
		return errors.New("unknown msg type")
	}

	return nil
}

////////////////////////// 主动操作 //////////////////////////

// send指向网络中（或者说外部依赖的网络模块）发送消息。 注意本地消息不要通过该方法使用
// 这个发送消息可以是单播也可以是多播，具体看
func (p *Pot) send(msg *defines.Message) {
	p.msgout <- msg
}

//// broadcast
//func (n *Node) broadcast(msg *defines.Message) {
//
//}

// 广播证明
func (p *Pot) broadcastProof() error {
	// 首先需要获取证明
	nb := p.txPool.GenBlock()
	p.maybeNewBlock = nb
	proof := &Proof{
		TxsNum:    uint64(len(nb.Txs)),
		TxsMerkle: nb.Merkle,
	}

	proofBytes, err := proof.Encode()
	if err != nil {
		return err
	}
	entry := &defines.Entry{
		Base:      nb.PrevHash,
		BaseIndex: nb.Index - 1,
		Type:      defines.EntryType_Proof,
		Data:      proofBytes,
	}

	// 广播
	p.pit.RangePeers(func(peer *defines.PeerInfo) error {
		if peer.Id != p.id {
			msg := &defines.Message{
				Version: defines.CodeVersion,
				Type:    defines.MessageType_Data,
				From:    p.id,
				To:      peer.Id,
				Entries: []*defines.Entry{entry},
			}
			err := msg.Sign()
			if err != nil {
				return err
			}
			p.send(msg)
		}

		return nil
	})

	return nil
}

// 广播新区块
func (p *Pot) broadcastNewBlock(nb *defines.Block) error {

	blockBytes, err := nb.Encode()
	if err != nil {
		//p.Errorf("broadcast new block(%v) fail: %s\n", nb, err)
		return err
	}

	entry := &defines.Entry{
		Base:      nb.PrevHash,
		BaseIndex: nb.Index - 1,
		Type:      defines.EntryType_Blocks,
		Data:      blockBytes,
	}

	// 广播
	p.pit.RangePeers(func(peer *defines.PeerInfo) error {
		if peer.Id != p.id {
			msg := &defines.Message{
				Version: defines.CodeVersion,
				Type:    defines.MessageType_Data,
				From:    p.id,
				To:      peer.Id,
				Entries: []*defines.Entry{entry},
			}
			err := msg.Sign()
			if err != nil {
				return err
			}
			p.send(msg)
		}
		return nil
	})
	return nil
}

// 暂时是直接向一个节点发送区块请求，但最好后面改成随机选举三个去请求
func (p *Pot) requestBlocks() {
	// TODO
}

// 向节点回应其请求的区块
func (p *Pot) responseBlocks(to string, blocks ...*defines.Block) error {
	l := len(blocks)
	if l == 0 {
		return nil
	}

	blockBytes := make([][]byte, l)
	for i:=0; i<l; i++ {
		b, err := blocks[i].Encode()
		if err != nil {
			//p.Errorf("response blocks fail: %s\n", err)
			return err
		}
		blockBytes[i] = b
	}

	entries := make([]*defines.Entry, l)
	for i := 0; i < l; i++ {
		entries[i] = &defines.Entry{
			BaseIndex: blocks[i].Index - 1,
			Base:      blocks[i].PrevHash,
			Type:      defines.EntryType_Blocks,
			Data:      blockBytes[i],
		}
	}

	msg := &defines.Message{
		Version: defines.CodeVersion,
		Type:    defines.MessageType_Data,
		From:    p.id,
		To:      to,
		Entries: entries,
	}
	err := msg.Sign()
	if err != nil {
		// p.Errorf("response blocks fail: %s\n", err)
		return err
	}

	p.send(msg)
	return nil
}

///////////////////////////////////////////////////////////

// 获取某个id的进度
func (p *Pot) getProcess(id string) defines.Process {
	var process defines.Process
	p.processesLock.RLock()
	process = *(p.processes[id])
	p.processesLock.RUnlock()
	return process
}

// 设置某个id的进度
func (p *Pot) setProcess(id string, process defines.Process) {
	p.processesLock.Lock()
	p.processes[id] = &process
	p.processesLock.Unlock()
}

// 获取某个id的proof
func (p *Pot) getProof(id string) Proof {
	var proof Proof
	p.proofsLock.RLock()
	proof = *(p.proofs[id])
	p.proofsLock.RUnlock()
	return proof
}

// 设置某个id的proof
func (p *Pot) setProof(id string, proof Proof) {
	p.proofsLock.Lock()
	p.proofs[id] = &proof
	p.proofsLock.Unlock()
}

// 获取某个区块缓存
func (p *Pot) getBlock(hash []byte) *defines.Block {
	key := fmt.Sprintf("%x", hash)
	var block *defines.Block
	p.blocksLock.RLock()
	block = p.blocksCache[key]
	p.blocksLock.RUnlock()
	return block
}

// 添加某个区块缓存
func (p *Pot) addBlock(block *defines.Block) {
	key := fmt.Sprintf("%x", block.SelfHash)
	p.blocksLock.Lock()
	p.blocksCache[key] = block
	p.blocksLock.Unlock()
}

///////////////////////////////////////////////////////

// 检查是否追上最新进度
func (p *Pot) checkLatestProcess() error {
	return nil
}

/////////////////////////////////////////////////////////

// 这里的处理都指的是在当前状态允许处理的情况

// 处理EntryBlocks
//
func (p *Pot) handleEntryBlocks(ent *defines.Entry) error {
	// 检查ent.BaseIndex和Base
	process := p.getProcess(p.id)
	if ent.BaseIndex < process.Index {
		return nil
	}
	if ent.BaseIndex == process.Index && !bytes.Equal(ent.Base, process.Hash) {
		return errors.New("mismatched ent.BaseIndex or ent.Base")
	}
	// 解码
	block := new(defines.Block)
	err := block.Decode(ent.Data)
	if err != nil {
		return err
	}
	// 检查区块本身格式/签名的正确性
	err = block.Verify()
	if err != nil {
		return err
	}
	// 检查block与ent中携带的Base/BaseIndex信息是否一致
	if !bytes.Equal(block.PrevHash, ent.Base) || block.Index != ent.BaseIndex + 1 {
		return errors.New("mismatched block.PreHash or block.Index")
	}
	// 如果其序号 = 本地的index+1，那么检查其有效性
	if block.Index == ent.BaseIndex + 1 {
		// 将Block传递到检查器进行检查

		// 检查通过后追加到本地分支

	} else if block.Index == ent.BaseIndex + 1 {
		// 如果序号是不连续的，暂时先保留在blocksCache
		p.addBlock(block)
	}
	return nil
}

func (p *Pot) handleEntryProof(ent *defines.Entry, from string) error {
	// 解码
	var proof Proof
	if err := proof.Decode(ent.Data); err != nil {
		return err
	}
	// 检查ent的Base信息是否合理
	process := p.getProcess(p.id)
	if ent.BaseIndex != process.Index || !bytes.Equal(ent.Base, process.Hash) {
		return errors.New("mismatched ent.Base or ent.BaseIndex")
	}
	// 如果证明信息有效，用以更新本地winner
	p.setProof(from,  proof)
	if proof.GreaterThan(p.proofs[p.winner]) {
		p.winner = from
	}
	return nil
}

// 将请求的区块返回回去
func (p *Pot) handleRequestBlocks(req *defines.Request) error {

	return nil
}

// 将自身的邻居表整理回发
func (p *Pot) handleRequestNeighbors(req *defines.Request) error {
	return nil
}