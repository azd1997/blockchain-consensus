/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/18/20 2:54 PM
* @Description: pot竞争时节点的动作
***********************************************************************/

package pot

import "github.com/azd1997/blockchain-consensus/defines"

// startPot 开始一轮Pot竞争
// 	1. 根据本地的交易池创建一个新区块，并且得到相应的proof
// 	2. 将proof广播
//  3. 重置proofs
func (p *Pot) startPot(moment Moment) {
	if p.duty == defines.PeerDuty_Peer {
		// 构造新区块并广播其证明，同时附带自身进度
		if err := p.broadcastSelfProof(); err != nil {
			p.Errorf("start pot fail: %s", err)
		}
	}
	// 所有角色还需要广播自身进度
}

// endPot 结束一轮Pot竞争
// 	1. 根据proofs得到本轮pot竞赛的自己判定的winner
func (p *Pot) endPot(moment Moment) {
	selfJudgeWinnerProof := p.proofs.JudgeWinner(moment)
	if selfJudgeWinnerProof == nil {	// proofs为空，则说明此时还没有共识节点加入进来
		// do nothing
	} else if selfJudgeWinnerProof.Id == p.id { // 自己是胜者
		p.broadcastNewBlock(p.maybeNewBlock)
	} else { // 别人是胜者
		// 等待胜者区块
	}
}

// decide 决定新区块
func (p *Pot) decide(moment Moment) {
	// 决定谁是胜者
	decidedWinnerProof := p.proofs.DecideWinner(moment)
	if decidedWinnerProof != nil {
		// 从未决区块表中取出胜者
		decidedWinnerBlock := p.udbt.Get(decidedWinnerProof.BlockHash)
		// 更新时钟(如果允许时间纠偏的话)
		if err := p.clock.Trigger(decidedWinnerBlock); err != nil {
			p.Errorf("Trigger clock fail: %s", err)
		}
		// 将胜者区块保存起来
		if err := p.bc.AddNewBlock(decidedWinnerBlock); err != nil {
			p.Errorf("BC add block fail: %s", err)
		}
		// 刷新进度表并更新自己进度
		p.processes.refresh(decidedWinnerBlock)
	} else {	// decided为nil说明，此时proofs表一个证明都没收到，正常情况下只有seed启动时会遇到。 异常情况下则是自己掉线了
		// 啥也不用干
	}

}
