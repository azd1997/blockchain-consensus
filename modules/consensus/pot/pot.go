/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:49
* @Description: Pot结构体就是实现PoT共识的类，其只负责逻辑处理与状态转换
***********************************************************************/

package pot

import (
	"errors"
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
	"github.com/azd1997/blockchain-consensus/modules/peerinfo"
	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/utils/math"
	"sync"
)

const (
	// TickMs 逻辑时钟每个滴答就是500ms
	TickMs = 500

	// DefaultMsgChanLen 默认channel长度
	DefaultMsgChanLen = 100

	// Module_Css 模块名
	Module_Css = "CSS"
)

// Option 选项
//type Option struct {
//	Id   string
//	Duty defines.PeerDuty
//
//	Pit  peerinfo.Pit
//	BC   requires.BlockChain
//	Net bnet.BNet
//}

// Pot pot节点
type Pot struct {
	id   string           // 账户、节点、客户端共用一个ID
	duty defines.PeerDuty // 普通结点/种子节点/工人节点

	//latest bool // 本节点是否追上系统最新进度

	// epoch等效于网络中最新区块索引
	// 节点启动时必须Ready后判断当前处于哪一epoch
	epoch int64

	stage StageType // 阶段
	state StateType	// pot state

	//processes map[string]*defines.Process
	//processesLock *sync.RWMutex
	//processes *processTable

	// 用于p.loopBeforeReady
	nWait          int
	nWaitChan      chan int
	nWaitBlockChan chan *defines.Block

	// 对外提供的消息通道
	// 本机节点生成新交易时，也是构造成交易消息从msgin传入
	msgin chan []byte // 用来传输*Message
	//msgout chan *defines.MessageWithError

	// 本地生成的交易从该chan传入
	localTxIn chan *defines.Transaction

	//// 交易传出通道
	//// 从网络中接收到的交易，需要传入该通道，交给bc存储
	//txout chan *defines.Transaction

	clock *Clock // 滴答器，每次滴答时刻都需要根据当前的状态变量确定状态该如何变更

	// potStartBeforeReady 用于启动时
	potStartBeforeReady chan Moment

	// b1Time	记录b1构造时间 ns
	b1Time int64

	// proofs表可能会因为某些节点出现恶意行为而将其删除
	//proofs          map[string]*Proof // 收集的其他共识节点的证明进度
	//proofsLock      *sync.RWMutex
	//winner          string
	proofs          *proofTable
	maybeNewBlock   *defines.Block
	waitingNewBlock *defines.Block // 等待的新区块，没等到就一直是nil

	udbt *undecidedBlockTable

	// 区块缓存的事交给Blockchain去做，这里不管
	//blocksCache map[string]*defines.Block // 同步到本机节点的区块，但尚未排好序的。也就是序列化没有接着本地最高区块后边的
	//blocksLock *sync.RWMutex

	////////////////////////// 本地依赖 /////////////////////////

	// 都是已经加载好的结构

	// 交易池
	// 1. 添加交易。交易的检验由交易池负责
	// 2. 生成新区块
	//txPool requires.TransactionPool

	// 区块链
	// 1. 添加区块，区块的校验由区块链负责
	// 2. 查询区块
	// 3. 添加交易。交易的检验由交易池负责
	// 4. 生成新区块
	bc requires.BlockChain

	// 节点信息表，传入的需要是已初始化好的
	pit peerinfo.Pit

	net bnet.BNet

	////////////////////////// 本地依赖 /////////////////////////

	*log.Logger
	done   chan struct{}
	inited bool

	///////////////////////// seeds peers ////////////////////
}

// New 新建Pot
func New(id string, duty defines.PeerDuty,
	pit peerinfo.Pit, bc requires.BlockChain,
	net bnet.BNet, msgchan chan []byte) (*Pot, error) {
	logger := log.NewLogger(Module_Css, id)
	if logger == nil {
		return nil, errors.New("nil logger, please init logger first")
	}

	var latestBlockHash []byte
	latestIndex := bc.GetMaxIndex()
	if latestIndex > 0 {
		latestBlock, err := bc.GetBlocksByRange(-1, 1)
		if err != nil {
			return nil, err
		}
		if len(latestBlock) > 0 {
			latestBlockHash = latestBlock[0].SelfHash
		}
	}
	proofs := newProofTable(latestIndex, latestBlockHash)

	p := &Pot{
		id:    id,
		duty:  duty,
		clock: NewClock(false),
		//processes:           newProcessTable(),
		msgin: msgchan,
		//msgout:              make(chan *defines.MessageWithError, DefaultMsgChanLen),
		localTxIn:           make(chan *defines.Transaction, DefaultMsgChanLen),
		potStartBeforeReady: make(chan Moment), //阻塞式
		proofs:              proofs,
		udbt:                newUndecidedBlockTable(),
		bc:                  bc,
		net:                 net,
		done:                make(chan struct{}),
		Logger:              logger,
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

func (p *Pot) Init() error {

	p.Infof("Init start")

	// 启动消息处理循环
	go p.MsgHandleLoop()
	// 启动状态切换循环(没有clock触发)
	go p.StateMachineLoop()

	// 区块链的最新状态
	bc := p.bc.GetMaxIndex()
	var err error
	// 根据节点duty和其本地区块链状态决定以何种逻辑启动
	if p.duty == defines.PeerDuty_Seed { // seed
		if bc == 0 { // 初次启动
			err = p.initForSeedFirstStart()
		} else { // 重启动
			err = p.initForSeedReStart()
		}
	} else { // peer
		if bc == 0 { // 初次启动
			err = p.initForPeerFirstStart()
		} else { // 重启动
			err = p.initForPeerReStart()
		}
	}

	if err != nil {
		return err
	}
	p.inited = true
	p.Infof("Init succ")
	return nil
}

func (p *Pot) Inited() bool {
	panic("implement me")
}

// Close 关闭
func (p *Pot) Close() error {
	close(p.done)
	return nil
}

// MsgOutChan 对外提供消息通道，用于数据向外传输
//func (p *Pot) MsgOutChan() chan *defines.MessageWithError {
//	return p.msgout
//}

// MsgInChan 对外提供消息通道，用于数据向内传输
func (p *Pot) MsgInChan() chan *defines.Message {
	return nil
}

func (p *Pot) LocalTxInChan() chan *defines.Transaction {
	return p.localTxIn
}

// 这个handlemsg的错误其实没啥用
func (p *Pot) HandleMsg(msg *defines.Message) error {
	// 检查消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	// 0-99 数据类
	case defines.MessageType_Blocks:
		return p.handleMsgBlocks(msg)
	case defines.MessageType_NewBlock:
		return p.handleMsgNewBlock(msg)
	case defines.MessageType_Txs:
		return p.handleMsgTxs(msg)
	case defines.MessageType_Peers:
		return p.handleMsgPeers(msg)
	case defines.MessageType_Proof:
		return p.handleMsgProof(msg)

	// 1-199 请求类
	case defines.MessageType_ReqBlockByIndex:
		return p.handleMsgReqBlockByIndex(msg)
	case defines.MessageType_ReqBlockByHash:
		return p.handleMsgReqBlockByHash(msg)
	case defines.MessageType_ReqPeers:
		return p.handleMsgReqPeers(msg)

	// 200-255 控制类

	default:
		return fmt.Errorf("unknown msg type: %s", msg.Type)
	}
}

func (p *Pot) StateMachineLoop() {

	once := new(sync.Once)

	for {
		select {
		case <-p.done:
			p.Infof("stateMachineLoop: return ...")
			return
		case moment := <-p.clock.Tick:
			p.Infof("[t%d] stateMachineLoop: clock tick: %s",
				math.RoundTickNo(moment.Time.UnixNano(), p.b1Time, TickMs), moment.String())

			// 根据当前状态来处理此滴答消息
			stage := p.getStage()
			state := p.getState()
			bcReady := p.isSelfReady()
			duty := p.duty
			mt := moment.Type

			// 根据这些条件，先切换stage，再切换state

			switch stage {

			case StageType_PreInited_RequestNeighbors: // nothing
			case StageType_PreInited_RequestFirstBlock:
				// 通过该chan向启动逻辑传递时刻信号
				if moment.Type == MomentType_PotStart {
					once.Do(func() {
						p.potStartBeforeReady <- moment
					})
				}
			case StageType_PreInited_RequestLatestBlock: // nothing
				// RLB阶段也应该重置proofs，见证pot竞争，接收最新区块
				if moment.Type == MomentType_PotStart {
					if bcReady {
						p.setStage(StageType_InPot)
						p.Infof("switch state from %s to %s",
							StageType_PreInited_RequestLatestBlock, StageType_InPot)
					}
					p.decide(moment)
					p.startPot(moment)
				} else if moment.Type == MomentType_PotOver {
					p.endPot(moment)
				}

			//case StageType_NotReady:
			//	// 能够处理邻居消息,区块消息,最新区块消息
			//
			//	p.Debugf("current is ready? %v", p.isSelfReady())
			//
			//	if moment.Type == MomentType_PotStart {
			//		p.setStage(StageType_InPot)
			//		p.Infof("switch state from %s to %s", StateType_NotReady, StageType_InPot)
			//		p.decide(moment)
			//		p.startPot(moment)
			//	} else if moment.Type == MomentType_PotOver {
			//		p.endPot(moment)
			//	}
			case StageType_InPot:
				if moment.Type == MomentType_PotOver { // 正常情况应该是PotOver时刻到来
					p.setStage(StageType_PostPot)
					p.Infof("switch state from %s to %s", StageType_InPot, StageType_PostPot)
					// 汇总已收集的证明消息，决出胜者，判断自己是否出块，接下去等待胜者区块和seed广播的胜者证明
					p.endPot(moment)
				} else if moment.Type == MomentType_PotStart { // 不可能出现的错误
					p.Errorf("stateMachineLoop: Moment(%s) comes at StateInPot", moment)
				}

			case StageType_PostPot:
				if moment.Type == MomentType_PotStart { // 正常情况应该是PotOver时刻到来
					p.setStage(StageType_InPot)
					p.Infof("switch state from %s to %s", StageType_PostPot, StageType_InPot)
					// 决定出新区块
					p.decide(moment)
					// 汇总已收集的证明消息，决出胜者，判断自己是否出块，接下去等待胜者区块和seed广播的胜者证明
					p.startPot(moment)
				} else if moment.Type == MomentType_PotOver { // 不可能出现的错误
					p.Errorf("stateMachineLoop: Moment(%s) comes at StatePostPot", moment)
				}
			default:
				p.Fatalf("stateMachineLoop: Moment(%s) comes at UnknownState(%s)", moment, state.String())

				//case StateType_NotReady:
				//	// 没准备好，啥也不干，等区块链同步
				//
				//	// 如果追上进度了则切换状态为ReadyCompete
				//	if p.latest() {
				//		p.setStage(StateType_ReadyCompete)
				//	} else {
				//		// 否则请求快照数据
				//		p.broadcastRequestBlocks(true)
				//	}
				//
				//case StateType_ReadyCompete:
				//	// 当前是ReadyCompete，则状态切换为Competing
				//
				//	// 状态切换
				//	p.setStage(StateType_Competing)
				//	// 发起竞争（广播证明消息）
				//	p.broadcastSelfProof()
				//
				//case StateType_Competing:
				//	// 当前是Competing，则状态切换为CompetingEnd，并判断竞赛结果，将状态迅速切换为Winner或Loser
				//
				//	// 状态切换
				//	p.setStage(StateType_CompeteOver)
				//	// 判断竞赛结果，状态切换
				//	if p.winner == p.id { // 自己胜出
				//		p.setStage(StateType_CompeteWinner)
				//		// 广播新区块
				//		p.broadcastNewBlock(p.maybeNewBlock)
				//	} else { // 别人胜出
				//		p.setStage(StateType_CompeteLoser)
				//		// 等待新区块(“逻辑上”的等待，代码中并不需要wait)
				//	}
				//
				//case StateType_CompeteOver:
				//	// 正常来说，tick时不会是恰好CompeteOver而又没确定是Winner/Loser
				//	// 所以暂时无视
				//case StateType_CompeteWinner:
				//	// Winner来说的话，立马广播新区块，广播结束后即切换为Ready
				//	// 所以不太可能tick时状态为Winner
				//	// 暂时无视
				//case StateType_CompeteLoser:
				//	// Loser等待新区块，接收到tick说明还没得到新区块
				//	// 状态切换为Ready
				//
				//	p.setStage(StateType_ReadyCompete)
				//	// 发起竞争（广播证明消息）
				//	p.broadcastSelfProof()
			}

			p.Info(p.bc.Display())
		}
	}
}

func (p *Pot) MsgHandleLoop() {
	var err error
	for {
		select {
		case <-p.done:
			p.Infof("msgHandleLoop: return ...")
			return
		case msgbytes := <-p.msgin:
			msg := new(defines.Message)
			if err := msg.Decode(msgbytes); err != nil {
				p.Errorf("msgHandleLoop: msg decode fail: err=%s, msgbytes=%v", err, msgbytes)
				continue
			}
			err = p.HandleMsg(msg)
			if err != nil {
				p.Errorf("msgHandleLoop: handle msg(%s) fail: msg=%s,err=%s", msg.Desc, msg, err)
			}
		case tx := <-p.localTxIn:
			// 存到本地
			p.bc.TxInChan() <- tx
			// 广播
			err = p.broadcastTx(tx)
			if err != nil {
				p.Errorf("msgHandleLoop: broadcast localtx(%s) fail: tx=%v, err=%s", tx.ShortName(), tx, err)
			}
		}
	}
}
