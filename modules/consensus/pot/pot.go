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
	"github.com/azd1997/blockchain-consensus/measure/report"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
	"github.com/azd1997/blockchain-consensus/modules/ledger"
	"github.com/azd1997/blockchain-consensus/modules/pitable"
	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/utils/math"
	"strings"
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

// Pot pot节点
type Pot struct {

	id   string           // 账户、节点、客户端共用一个ID
	duty defines.PeerDuty // 普通结点/种子节点/工人节点

	// 创世节点信息
	// 对于非创世节点，genesisData无意义
	// 对于创世节点，genesisId, genesisAddr 置空，genesisData必填
	//genesisId, genesisAddr string
	genesisData string	// 如果genesis非空，则本机节点创世

	// epoch等效于网络中最新区块索引
	// 节点启动时必须Ready后判断当前处于哪一epoch
	epoch int64

	stage StageType // 阶段
	state StateType // pot state

	// 用于Init阶段的一些变量
	nWait                 int
	nWaitChan             chan int
	nWaitBlockChan        chan *defines.Block
	onceWaitFirstPotStart sync.Once
	potStartBeforeReady   chan Moment

	// b1Time	记录b1构造时间 ns
	b1Time int64

	// 对外提供的消息通道
	// 本机节点生成新交易时，也是构造成交易消息从msgin传入
	msgin chan *defines.Message // 用来传输*Message

	clock *Clock // 滴答器，每次滴答时刻都需要根据当前的状态变量确定状态该如何变更

	// 竞争相关
	proofs          *proofTable
	maybeNewBlock   *defines.Block
	waitingNewBlock *defines.Block // 等待的新区块，没等到就一直是nil
	udbt            *undecidedBlockTable

	////////////////////////// 本地依赖 /////////////////////////

	// 区块链
	// 1. 添加区块，区块的校验由区块链负责
	// 2. 查询区块
	// 3. 添加交易。
	// 4. 生成新区块
	bc ledger.Ledger

	// 节点信息表，传入的需要是已初始化好的
	pit pitable.Pit

	net bnet.BNet

	// 报告器
	monitorId, monitorHost string
	report.Reporter

	////////////////////////// 本地依赖 /////////////////////////

	*log.Logger
	done   chan struct{}
	inited bool

	///////////////////////// seeds peers ////////////////////

}

////////////////////////  ///////////////////////
// 正常运行都需要的成员不使用SetXXX注入，全放进New()中，
// 这样可以约束调用者，避免出错。

// New 新建Pot
func New(id string, duty defines.PeerDuty,
	pit pitable.Pit, bc requires.BlockChain,
	net bnet.BNet, msgchan chan *defines.Message,
	monitorId, monitorHost string,
	//genesisId, genesisAddr string,
	genesisData ...string) (*Pot, error) {

	logger := log.NewLogger(Module_Css, id)
	if logger == nil {
		return nil, errors.New("nil logger, please init logger first")
	}

	// 构建reporter
	reporter, err := report.NewGeneralReporter(monitorId, monitorHost, net)
	if err != nil {
		return nil, fmt.Errorf("NewGeneralReporter fail: %s", err)
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

	genesisStr := ""
	if len(genesisData) > 0 {
		genesisStr = strings.Join(genesisData, " ")
	}
	if strings.TrimSpace(genesisStr) == "" {
		genesisStr = ""
	}

	p := &Pot{
		id:                  id,
		duty:                duty,
		//genesisId: genesisId,
		//genesisAddr: genesisAddr,
		genesisData: genesisStr,
		clock:               NewClock(false),
		msgin:               msgchan,
		potStartBeforeReady: make(chan Moment), //阻塞式
		proofs:              proofs,
		udbt:                newUndecidedBlockTable(),
		bc:                  bc,
		net:                 net,
		done:                make(chan struct{}),
		Logger:              logger,
		// 报告器
		monitorId: monitorId,
		monitorHost: monitorHost,
		Reporter:reporter,
	}

	if pit == nil || !pit.Inited() {
		return nil, errors.New("PeerInfoTable is nil or un-inited")
	} else {
		p.pit = pit
	}

	return p, nil
}

//////////////////////////// 实现接口 ///////////////////////////

// Init 启动
// 对于创世seed以外的节点而言，启动流程：
// stage: RN -> RFB -> wait PotStart -> InPot -> PostPot -> InPot -> ...
// state: ----------------------------- witness -> learner -> 此时才开始推算该是什么状态
func (p *Pot) Init() (err error) {
	p.Infof("Init start")

	// 启动消息处理循环x
	go p.MsgHandleLoop()
	// 启动状态切换循环(没有clock触发)
	go p.StateMachineLoop()

	// 区块链的最新状态
	bc := p.bc.GetMaxIndex()

	// 是否是创世节点
	if p.genesisData != "" && bc == 0 {	// 本地已有区块链数据则不可再创世
		// 创世启动
		err = p.initForGenesis()
	} else {
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
	}
	if err != nil {
		return err
	}

	p.inited = true
	p.Infof("Init succ")
	return nil
}

func (p *Pot) Inited() bool {
	return p.inited
}

// Ok 判断依赖项是否准备好
func (p *Pot) Ok() bool {
	return true
}

// Close 关闭
func (p *Pot) Close() error {
	// 关闭自身工作循环
	close(p.done)
	// 关闭所依赖的其他组件
	p.net.Close()
	p.pit.Close()
	p.bc.Close()
	log.Close(p.id)
	return nil
}

// Closed 关闭
func (p *Pot) Closed() bool {
	return false
}

// SetMsgInChan 对外提供消息通道，用于数据向内传输
func (p *Pot) SetMsgInChan(bus chan *defines.Message) {
	p.msgin = bus
}

// HandleMsg 这个handlemsg的错误其实没啥用
func (p *Pot) HandleMsg(msg *defines.Message) error {
	// 检查消息格式与签名
	if msg.Type != defines.MessageType_LocalTxs {
		if err := msg.Verify(); err != nil {
			return err
		}
	}

	switch msg.Type {
	// 0-99 数据类
	case defines.MessageType_Blocks:
		return p.handleMsgBlocks(msg)
	case defines.MessageType_NewBlock:
		return p.handleMsgNewBlock(msg)
	case defines.MessageType_Txs:
		return p.handleMsgTxs(msg)
	case defines.MessageType_LocalTxs:
		return p.handleMsgLocalTxs(msg)
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
	for {
		select {
		case <-p.done:
			p.Infof("stateMachineLoop: return ...")
			return
		case moment := <-p.clock.Tick:
			p.Infof("[t%d] StateMachineLoop: clock tick: %s",
				math.RoundTickNo(moment.Time.UnixNano(), p.b1Time, TickMs), moment.String())
			p.handleTick(moment)
			p.Info(p.bc.Display())
		}
	}
}

func (p *Pot) MsgHandleLoop() {
	var err error
	for {
		select {
		case <-p.done:
			p.Infof("MsgHandleLoop: return ...")
			return
		case msg := <-p.msgin:
			go func(msg *defines.Message) {		// 另外处理
				p.Debugf("MsgHandleLoop: handle msg(%s) start: msg=%s", msg.Desc, msg)
				err = p.HandleMsg(msg)
				if err != nil {
					p.Errorf("MsgHandleLoop: handle msg(%s) fail: msg=%s,err=%s", msg.Desc, msg, err)
				}
			}(msg)
		}
	}
}

//// HeartbeatLoop 心跳循环
//// 每个节点启动之后都需要启动心跳循环，心跳循环的作用是使其他节点知道自己的存在而不删除自己
//func (p *Pot) HeartbeatLoop() {
//	var err error
//	tick := time.Tick(time.Duration(TickMs) * time.Millisecond * 10)		// 每10个TICK需要心跳一次
//
//	for {
//		select {
//		case <-p.done:
//			p.Infof("HeartbeatLoop: return ...")
//			return
//		case t := <-tick:
//			err = p.broadcastHeartbeat(t)
//			if err != nil {
//				p.Errorf("HeartbeatLoop: broadcastHeartbeat fail: t=%s,err=%s", t, err)
//			}
//		}
//	}
//}

func (p *Pot) handleTick(m Moment) {
	//ti := math.RoundTickNo(m.Time.UnixNano(), p.b1Time, TickMs)

	// 根据当前状态来进行状态变换
	stage := p.getStage()
	state := p.getState()
	bcReady := p.isSelfReady()
	duty := p.duty

	// RN阶段
	if stage == StageType_PreInited_RequestNeighbors {
		return // nothing
	}

	// RFB阶段
	if stage == StageType_PreInited_RequestFirstBlock {
		// 之所以使用这样一个PotStart时刻，
		// 是因为在从start clock到第1个PotStart，
		// 中间的过程如果直接从RFB尝试切换到InPot或者PostPOt情况未定，比较麻烦

		// 通过该chan向启动逻辑(init())传递时刻信号
		// 启动逻辑中会根据该信号，将stage切换至InPot
		// 并且将 RFB阶段的 除创世seed以外的其他节点 都切换为witness状态
		// 使进度未明（不清楚是否ready）的节点强制为witness，可以简化程序逻辑
		if m.Type == MomentType_PotStart {
			p.onceWaitFirstPotStart.Do(func() {
				p.potStartBeforeReady <- m
			})
		}
		return // nothing
	}

	// 非RN/RFB这两种特殊stage的情况下 （也就是InPot和PostPot两种stage）
	// 有以下状态变化的规则

	// PotStart到来
	if m.Type == MomentType_PotStart {
		p.setStage(StageType_InPot)

		// decide新区块 decide上一轮PoT竞争结束后得到的新区块
		p.decide(m)

		// 重置proofs
		latestBlock := p.bc.GetLatestBlock()
		p.proofs.Reset(m, latestBlock)

		// ready peer 在 PotStart时刻到来时成为 competitor
		if duty == defines.PeerDuty_Peer && bcReady {
			p.setState(StateType_Competitor)

			// 做competitor的事
			p.Info("start pot competetion. broadcast self proof")

			// 构造新区块并广播其证明，同时附带自身进度
			if err := p.broadcastSelfProof(false); err != nil {
				p.Errorf("start pot fail: %s", err)
			}
		} else { // not ready peer 以及 非peer的节点 在 PotStart时刻到来时成为 witness
			p.setState(StateType_Witness)

			// 成为witness时do nothing

			// 每次在PotSTart来临时，自己还是not ready，那么就请求区块。 （具体的取哪些区块 TODO 先直接要求所有的区块）
			if !bcReady {
				// TODO 增加RequestBlockVyIndexes([]int)
				p.broadcastRequestBlocksByIndex(2, 100, true, true)
			}
		}

		return
	}

	// PotOver到来
	if m.Type == MomentType_PotOver {
		p.setStage(StageType_PostPot)

		// 重置udbt，为后面接收新区块做准备
		latestBlock := p.bc.GetLatestBlock()
		if latestBlock == nil {
			p.udbt.Reset(0)
		} else {
			p.udbt.Reset(latestBlock.Index + 1)
		}

		// judge winner
		selfJudgeWinnerProof := p.proofs.JudgeWinner(m)
		p.Info(p.proofs.Display())
		p.Info(p.udbt.Display())

		if selfJudgeWinnerProof == nil { // proofs为空，则说明此时还没有共识节点加入进来，或者说没有节点能够参赛
			// do nothing
			p.Info("end pot competetion. judge winner, no winner")
			// 对于seed而言，还需要将本地最新区块广播出去
			if p.duty == defines.PeerDuty_Seed && bcReady {
				p.setState(StateType_Judger)
				p.broadcastNewBlock(p.bc.GetLatestBlock())
			} else {
				p.setState(StateType_Learner)
			}

		} else { // winner exists

			if selfJudgeWinnerProof.Id == p.id && state == StateType_Competitor { // i win
				p.Infof("end pot competetion. judge winner, i am winner(%s), broadcast new block(%s) now", selfJudgeWinnerProof.Short(), p.maybeNewBlock.ShortName())
				p.setState(StateType_Winner) // winner
				p.udbt.Add(p.maybeNewBlock)  // 将自己的新区块添加到未决区块表
				p.broadcastNewBlock(p.maybeNewBlock)
			} else { // 别人是胜者
				if p.duty == defines.PeerDuty_Seed { // 如果是种子节点，还要把种子节点自己判断的winner广播出去
					// 等待胜者区块
					p.Infof("end pot competetion. judge winner, wait winner(%s) and broadcast to all peers", selfJudgeWinnerProof.Short())
					p.setState(StateType_Judger) // judger
					p.proofs.AddProofRelayedBySeed(selfJudgeWinnerProof)
					p.broadcastProof(selfJudgeWinnerProof, false, true)
				} else { // 其他的话只需要等待
					// 等待胜者区块
					p.setState(StateType_Learner)
					p.Infof("end pot competetion. judge winner, wait winner(%s)", selfJudgeWinnerProof.Short())
				}
			}
		}
	}
}
