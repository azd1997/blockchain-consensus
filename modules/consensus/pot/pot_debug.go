package pot

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
	"github.com/azd1997/blockchain-consensus/modules/pitable"
	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/test"
	"github.com/azd1997/blockchain-consensus/utils/math"
	"sort"
	"sync"
)

// 用于测试版本的pot

type DebugPot struct {
	*Pot

	// 测试用的一些变量
	test         bool // test=true时才考虑定时关闭和cheat
	testOnce     sync.Once
	shutdownAtTi int   // ti时定时关闭，若<=0则不生效
	cheatAtTi    []int // ti时广播虚假的证明，将自己的证明数报至一个非常大的数：99999999

	tm *test.TxMaker
}

// NewDebugPot 新建Pot
func NewDebugPot(id string, duty defines.PeerDuty,
	pit pitable.Pit, bc requires.BlockChain,
	net bnet.BNet, msgchan chan *defines.Message,
	monitorId, monitorHost string,
//genesisId, genesisAddr string,
	genesisData ...string) (*DebugPot, error) {

	pot, err := New(id, duty, pit, bc, net, msgchan, monitorId, monitorHost, genesisData...)
	if err != nil {
		return nil, err
	}
	dp := &DebugPot{
		Pot:          pot,
		tm:           nil,
	}
	return dp, nil
}

////////////////////////// 测试阶段用于修改行为 //////////////////////////////

// ShutdownAt 定时关闭。i为ti，或者说全网第几个TICK
func (p *DebugPot) ShutdownAt(i int) {
	p.testOnce.Do(func() {
		if i > 0 {
			p.test = true
			p.shutdownAtTi = i
		}
	})
}

// CheatAt 设置定时作弊
func (p *DebugPot) CheatAt(is ...int) {
	p.testOnce.Do(func() {
		sort.Ints(is)
		idx := 0
		for j, i := range is {
			if i > 0 {
				idx = j
				break
			}
		}
		is = is[idx:]
		if len(is) > 0 {
			p.test = true
			p.cheatAtTi = is
		}
	})
}

// CheatShutdownAt 设置定时作弊与关闭
func (p *DebugPot) CheatShutdownAt(shutdownAt int, cheatAt ...int) {
	p.testOnce.Do(func() {
		if shutdownAt > 0 {
			p.test = true
			p.shutdownAtTi = shutdownAt
		}

		sort.Ints(cheatAt)
		idx := 0
		for j, i := range cheatAt {
			if i > 0 {
				idx = j
				break
			}
		}
		cheatAt = cheatAt[idx:]
		if len(cheatAt) > 0 {
			p.test = true
			p.cheatAtTi = cheatAt
		}
	})
}

// 覆盖原本的handleTick()
func (p *DebugPot) handleTick(m Moment) {
	ti := math.RoundTickNo(m.Time.UnixNano(), p.b1Time, TickMs)

	// test模式下才检查定时关闭
	if p.test {
		if p.shutdownAtTi > 0 && ti >= p.shutdownAtTi {
			p.Infof("close at t[%d]", ti)
			p.Close()
			return
		}
	}

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

		// decide新区块
		p.decide(m)

		// 重置proofs
		latestBlock := p.bc.GetLatestBlock()
		p.proofs.Reset(m, latestBlock)

		// ready peer 在 PotStart时刻到来时成为 competitor
		if duty == defines.PeerDuty_Peer && bcReady {
			p.setState(StateType_Competitor)

			// 做competitor的事
			p.Info("start pot competetion. broadcast self proof")

			// test模式下检查是否有定时cheat
			cheatNow := false
			if p.test {
				idx := 0
				for i, elem := range p.cheatAtTi {
					if ti == elem {
						idx = i + 1
						cheatNow = true
						break
					}
					if ti > elem {
						idx = i
						break
					}
				}
				p.cheatAtTi = p.cheatAtTi[idx:]
			}
			// 构造新区块并广播其证明，同时附带自身进度
			if err := p.broadcastSelfProof(cheatNow); err != nil {
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
		p.setStage(StageType_InPot)

		// 重置udbt
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
