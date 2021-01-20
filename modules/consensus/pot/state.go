/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/20/21 8:55 PM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"sync/atomic"
)

type StateType uint32

const (
	// 进入RLB阶段及之后才有state概念，之前都是None
	StateType_None StateType = iota
	StateType_Witness
	StateType_Competitor
	StateType_Winner
	StateType_Judger
	StateType_Learner
	)

var stateMap = map[StateType]string{
	StateType_None:"none",
	StateType_Witness: "witness",
	StateType_Competitor: "competitor",
	StateType_Winner : "winner",
	StateType_Judger : "judger",
	StateType_Learner : "learner",
}

func (st StateType) String() string {
	if str, ok := stateMap[st]; ok {
		return str
	}
	return "unknown"
}

// pot_state = state_switch ( bc_state, moment, duty )
// stage是moment的另一种表现
// stage_RLB具有一些特殊性

// 根据输入确定当前应该为何种potstate
// 应当在每次Tick到来时调用
// 如果只考虑InPot和PostPot两个阶段的话，这个转换公式可以只要stage或者moment
func (p *Pot) stateSwitch(moment Moment) {

	duty := p.duty
	bcReady := p.isSelfReady()
	stage := p.getStage()

	// 这时没有状态的概念
	if stage == StageType_PreInited_RequestNeighbors || stage == StageType_PreInited_RequestFirstBlock {
		return
	}

	// potstart到来

	if duty == defines.PeerDuty_Peer && bcReady && moment.Type == MomentType_PotStart {
		p.setState(StateType_Competitor)
		return
	}

	if moment.Type == MomentType_PotStart &&
		(duty != defines.PeerDuty_Peer || (duty == defines.PeerDuty_Peer && !bcReady)) {
		p.setState(StateType_Witness)
		return
	}

	// potover到来
	if moment.Type == MomentType_PotStart &&
		(p.getState() == StateType_Competitor )
}

// 查看当前状态
func (p *Pot) getState() StateType {
	return StateType(atomic.LoadUint32((*uint32)(&p.state)))
}

// 更新当前状态
func (p *Pot) setState(newState StateType) {
	atomic.StoreUint32((*uint32)(&p.state), uint32(newState))
}