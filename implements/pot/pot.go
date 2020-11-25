/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:49
* @Description: Pot结构体就是实现PoT共识的类，其只负责逻辑处理与状态转换
***********************************************************************/

package pot

import (
	"errors"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/peerinfo"
	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/utils/log"
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
type Option struct {
	Id      string
	Duty    defines.PeerDuty
	LogDest log.LogDest
	Pit     *peerinfo.PeerInfoTable
}

// Pot pot节点
type Pot struct {
	id   string           // 账户、节点、客户端共用一个ID
	duty defines.PeerDuty // 普通结点/种子节点/工人节点

	//latest bool // 本节点是否追上系统最新进度

	// epoch等效于网络中最新区块索引
	// 节点启动时必须Ready后判断当前处于哪一epoch
	epoch int64

	state StateType // 状态状态

	//processes map[string]*defines.Process
	//processesLock *sync.RWMutex
	processes *processTable

	// 用于p.loopBeforeReady
	nWait     int
	nWaitChan chan int

	// 对外提供的消息通道
	// 本机节点生成新交易时，也是构造成交易消息从msgin传入
	msgin  chan *defines.Message
	msgout chan *defines.MessageWithError

	clock *Clock // 滴答器，每次滴答时刻都需要根据当前的状态变量确定状态该如何变更

	// potStartBeforeReady 用于启动时
	potStartBeforeReady chan Moment

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
	txPool requires.TransactionPool

	// 区块链
	// 1. 添加区块，区块的校验由区块链负责
	// 2. 查询区块
	bc requires.BlockChain

	// 节点信息表，传入的需要是已初始化好的
	pit *peerinfo.PeerInfoTable

	////////////////////////// 本地依赖 /////////////////////////

	*log.Logger
	done chan struct{}

	///////////////////////// seeds peers ////////////////////
}

// New 新建Pot
func New(opt *Option) (*Pot, error) {
	p := &Pot{
		id:                  opt.Id,
		duty:                opt.Duty,
		state:               StateType_PreInited, // 初始状态
		processes:           newProcessTable(),
		msgin:               make(chan *defines.Message, DefaultMsgChanLen),
		msgout:              make(chan *defines.MessageWithError, DefaultMsgChanLen),
		potStartBeforeReady: make(chan Moment), //阻塞式
		proofs:              nil,
		udbt:                newUndecidedBlockTable(),
		done:                make(chan struct{}),
		Logger:              log.NewLogger(opt.LogDest, Module_Css, opt.Id),
	}

	if opt.Pit == nil {
		p.pit = peerinfo.Global()
	} else {
		if opt.Pit.Inited() {
			p.pit = opt.Pit
		} else {
			return nil, errors.New("PeerInfoTable should be inited")
		}
	}

	return p, nil
}

//////////////////////////// 实现接口 ///////////////////////////

// Close 关闭
func (p *Pot) Close() error {
	close(p.done)
	return nil
}

// OutMsgChan 对外提供消息通道，用于数据向外传输
func (p *Pot) OutMsgChan() chan *defines.MessageWithError {
	return p.msgout
}

// InMsgChan 对外提供消息通道，用于数据向内传输
func (p *Pot) InMsgChan() chan *defines.Message {
	return p.msgin
}
