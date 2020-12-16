/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/18/20 2:49 PM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"fmt"
	"bytes"
	"sync"
)

// 证明表
type proofTable struct {
	baseIndex    int64             // 基于哪个区块开始的竞争
	base         []byte            //
	table        map[string]*Proof // <id, *Proof>
	sync.RWMutex                   // 保护table
	winner       *Proof            // 胜者

	relayed map[string]int // seed转发的winnerproof的投票数
	// relayedWinner *Proof		 	// seed公认的winner，众数比较

	// 当PotStart时决定新区块时，需要依赖relayedWinner, 如果没有relayed，则相信自己的winner

	start, end Moment // 本轮竞争的PotStart/PotEnd时刻

	Judged  *Proof // 自己判定的胜者
	Decided *Proof // 每轮竞争确定的winner，综合自己判定和种子转发
}

// Add 添加
func (proofs *proofTable) Add(p *Proof) {
	if p.BaseIndex != proofs.baseIndex || !bytes.Equal(p.Base, proofs.base) {
		return // TODO: 是否返回错误，让上层回复来信方？
	}

	proofs.Lock()
	proofs.table[p.Id] = p
	proofs.Unlock()

	if p.GreaterThan(proofs.winner) {
		proofs.winner = p
	}
}

// AddProofRelayedBySeed 添加seed转发的proof
func (proofs *proofTable) AddProofRelayedBySeed(relayed *Proof) {
	proofs.Add(relayed)
	proofs.relayed[relayed.Id]++
}

// JudgeWinner 获胜者的proof
// 必须在PotOver时调用
// Judge是自己判定的
func (proofs *proofTable) JudgeWinner(moment Moment) *Proof {
	if moment.Type == MomentType_PotOver && moment.Time.After(proofs.start.Time) {
		
		proofs.Judged = proofs.winner
		fmt.Println("JudgeWinner: ", proofs.Judged)
		return proofs.Judged
	}
	return nil
}

// DecideWinner 决定winner
// 必须在PotStart时刻调用
// Decide 是在收到种子们的决定之后再综合决定新区块是哪个
func (proofs *proofTable) DecideWinner(moment Moment) *Proof {
	// 是PotStart时刻并且比本轮开始时的PotStart大
	if moment.Type == MomentType_PotStart && moment.Time.After(proofs.start.Time) {
		// 确定出seed承认的winner
		var seedRelayWinnerProof *Proof
		var seedRelayWinner string
		max := 0
		for id, count := range proofs.relayed {
			if count > max {
				max = count
				seedRelayWinner = id
			}
		}
		proofs.RLock()
		seedRelayWinnerProof = proofs.table[seedRelayWinner]
		proofs.RUnlock()

		// 确定自己承认的winner
		selfJudgeWinnerProof := proofs.Judged

		if seedRelayWinnerProof != nil {
			proofs.Decided = seedRelayWinnerProof
		} else {
			proofs.Decided = selfJudgeWinnerProof
		}
		return proofs.Decided
	}
	return nil
}

// Reset 重置
func (proofs *proofTable) Reset() {
	if proofs.winner == nil {
		return // 重置失败
	}
	proofs.baseIndex = proofs.winner.BaseIndex + 1
	proofs.base = proofs.winner.BlockHash
	proofs.winner = nil
	proofs.Lock()
	proofs.table = map[string]*Proof{}
	proofs.Unlock()
}

// newProofTable 使用最新的区块去创建证明表
func newProofTable(latestBlockIndex int64, latestBlockHash []byte) *proofTable {
	return &proofTable{
		baseIndex: latestBlockIndex,
		base:      latestBlockHash,
		table:     map[string]*Proof{},
		relayed:   map[string]int{},
	}
}
