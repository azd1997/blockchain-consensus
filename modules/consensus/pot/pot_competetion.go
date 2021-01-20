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

	// 重置proofs
	latestBlock := p.bc.GetLatestBlock()
	p.proofs.Reset(moment, latestBlock)

	// 对于已经准备好的共识节点，此时还需要广播自身的证明消息
	if p.duty == defines.PeerDuty_Peer && p.isSelfReady() {
		p.Info("start pot competetion. broadcast self proof")
		// 构造新区块并广播其证明，同时附带自身进度
		if err := p.broadcastSelfProof(); err != nil {
			p.Errorf("start pot fail: %s", err)
		}
	} else { // 其他所有的情况都是见证
		p.Info("start pot competetion. witness")
	}
	// 所有角色还需要广播自身进度
}

// endPot 结束一轮Pot竞争
// 	1. 根据proofs得到本轮pot竞赛的自己判定的winner
func (p *Pot) endPot(moment Moment) {

	// 刷新udbt
	latestBlock := p.bc.GetLatestBlock()
	if latestBlock == nil {
		p.udbt.Reset(0)
	} else {
		p.udbt.Reset(latestBlock.Index + 1)
	}

	// 通过proofs裁决winner
	selfJudgeWinnerProof := p.proofs.JudgeWinner(moment)
	p.Info(p.proofs.Display())
	//p.Info(p.udbt.Display())
	if selfJudgeWinnerProof == nil { // proofs为空，则说明此时还没有共识节点加入进来
		// do nothing
		p.Info("end pot competetion. judge winner, no winner")
	} else if selfJudgeWinnerProof.Id == p.id { // 自己是胜者
		p.Infof("end pot competetion. judge winner, i am winner(%s), broadcast new block(%s) now", selfJudgeWinnerProof.Short(), p.maybeNewBlock.ShortName())
		p.udbt.Add(p.maybeNewBlock) // 将自己的新区块添加到未决区块表
		p.broadcastNewBlock(p.maybeNewBlock)
	} else { // 别人是胜者
		if p.duty == defines.PeerDuty_Seed { // 如果是种子节点，还要把种子节点自己判断的winner广播出去
			// 等待胜者区块
			p.Infof("end pot competetion. judge winner, wait winner(%s) and broadcast to all peers", selfJudgeWinnerProof.Short())
			p.proofs.AddProofRelayedBySeed(selfJudgeWinnerProof)
			p.broadcastProof(selfJudgeWinnerProof, true)
		} else { // 其他的话只需要等待
			// 等待胜者区块
			p.Infof("end pot competetion. judge winner, wait winner(%s)", selfJudgeWinnerProof.Short())
		}
	}
}

// decide 决定新区块
func (p *Pot) decide(moment Moment) {
	p.Info("decide new block now")
	p.Info(p.proofs.Display())
	p.Info(p.udbt.Display())
	// 决定谁是胜者
	decidedWinnerProof := p.proofs.DecideWinner(moment)
	if decidedWinnerProof != nil {
		// 从未决区块表中取出胜者
		decidedWinnerBlock := p.udbt.Get(decidedWinnerProof.BlockHash)

		// 这说明没收到胜者的区块。
		if decidedWinnerBlock == nil {
			p.Errorf("proof decided, but decided block not found")
			return
		}

		// 拿到decided proof 和 decided block 后，需要校验
		if !decidedWinnerProof.Match(decidedWinnerBlock) {
			p.Errorf("proof decided, but decided block doesn't match the decided proof")
			// TODO: decided block的构造者作假，惩罚
			return
		}

		// TODO: 对decidedWinnerBlock内容的校验

		// 更新时钟(如果允许时间纠偏的话)
		if err := p.clock.Trigger(decidedWinnerBlock); err != nil {
			p.Errorf("Trigger clock fail: %s", err)
		}
		// 将胜者区块保存起来
		if err := p.bc.AddNewBlock(decidedWinnerBlock); err != nil {
			p.Errorf("BC add block fail: %s", err)
		}
		// 刷新进度表并更新自己进度 （暂时没使用）
		//p.processes.refresh(decidedWinnerBlock)
	} else { // decided为nil说明，此时proofs表一个证明都没收到，正常情况下只有seed启动时会遇到。 异常情况下则是自己掉线了
		// 啥也不用干
		p.Debug("decide winner proof but no winner found")
	}

}
