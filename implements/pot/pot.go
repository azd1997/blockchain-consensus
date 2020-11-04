/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:49
* @Description: Pot结构体就是实现PoT共识的类，其只负责逻辑处理与状态转换
***********************************************************************/

package pot

import (
	"errors"
	"sync"
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
	epoch uint64

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
	msgout chan *defines.Message

	clock *time.Timer // 滴答器，每次滴答时刻都需要根据当前的状态变量确定状态该如何变更

	// 节点信息表，传入的需要是已初始化好的
	pit *peerinfo.PeerInfoTable

	// proofs表可能会因为某些节点出现恶意行为而将其删除
	proofs          map[string]*Proof // 收集的其他共识节点的证明进度
	proofsLock      *sync.RWMutex
	winner          string
	maybeNewBlock   *defines.Block
	waitingNewBlock *defines.Block // 等待的新区块，没等到就一直是nil

	// 区块缓存的事交给Blockchain去做，这里不管
	//blocksCache map[string]*defines.Block // 同步到本机节点的区块，但尚未排好序的。也就是序列化没有接着本地最高区块后边的
	//blocksLock *sync.RWMutex

	////////////////////////// 本地依赖 /////////////////////////

	// 交易池
	// 1. 添加交易。交易的检验由交易池负责
	// 2. 生成新区块
	txPool requires.TransactionPool

	// 区块链
	// 1. 添加区块，区块的校验由区块链负责
	// 2. 查询区块
	bc requires.BlockChain

	////////////////////////// 本地依赖 /////////////////////////

	*log.Logger
	done chan struct{}
}

// New 新建Pot
func New(opt *Option) (*Pot, error) {
	p := &Pot{
		id:         opt.Id,
		duty:       opt.Duty,
		state:      StateType_Init_GetNeighbors, // 初始状态
		processes:  newProcessTable(),
		msgin:      make(chan *defines.Message, DefaultMsgChanLen),
		msgout:     make(chan *defines.Message, DefaultMsgChanLen),
		proofs:     map[string]*Proof{},
		proofsLock: new(sync.RWMutex),
		done:       make(chan struct{}),
		Logger:     log.NewLogger(opt.LogDest, Module_Css, opt.Id),
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

// Init 初始化
// 启动流程：
//		1. 启动消息处理循环
//		2. 先后请求Neighbors/Processes/Blocks，直至自身节点进入Ready状态
// 		3. 在取到“最新区块”的时候，按照自身时间戳与最新区块的构建时间/索引，以及响应侧epoch，判断当前处于何种阶段，启动定时器
//		4. 启动状态切换循环
//
// 种子节点启动和非种子节点启动有所不同:
// 		1. 种子节点启动时需要向其他种子节点与所有一直在线的非种子节点获取信息
//		2. 非种子节点启动时可以只通过种子节点来获取信息直至到达最新状态
//
// ** 整个网络启动时一定是先启动种子节点而后启动非种子节点；种子节点允许中间重启
//
func (p *Pot) Init() error {

	// 启动消息处理循环
	go p.msgHandleLoop()

	// 阻塞直到追上最新进度
	p.loopBeforeReady()

	// 切换状态，准备进入状态切换循环
	p.setState(StateType_ReadyCompete)

	// 启动世界时钟
	p.clock = time.NewTimer(time.Second)

	// 启动状态切换循环
	go p.stateMachineLoop()

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
