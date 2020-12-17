/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/18/20 2:49 PM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"bytes"
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"sync"
)

// 证明表
type proofTable struct {
	// 如果有最新区块，那么在使用proofs需要校验最新区块
	// 如果没有，则是在两种特殊情况下：1. RLB阶段； 2. 本地区块链发现自己遇到了错误的部分，丢弃，重新RLB
	HasLatestBlockNow bool

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
	//oldDecided *Proof				// 上一轮的决胜者
}

// Add 添加
func (proofs *proofTable) Add(p *Proof) {

	if p == nil {
		return
	}

	if proofs.HasLatestBlockNow {
		if p.BaseIndex != proofs.baseIndex || !bytes.Equal(p.Base, proofs.base) {
			return // TODO: 是否返回错误，让上层回复来信方？
		}
	}	// 当前没有最新区块就不做此检查


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

		proofs.end = moment

		proofs.Judged = proofs.winner

		if proofs.Judged != nil {
			fmt.Printf("JudgeWinner: winner(%s)(%d, %x)", proofs.Judged.Id, proofs.Judged.TxsNum, proofs.Judged.BlockHash)
		}

		return proofs.Judged
	}
	return nil
}

// DecideWinner 决定winner
// 必须在PotStart时刻调用
// Decide 是在收到种子们的决定之后再综合决定新区块是哪个
// 调用完Decide之后必须检查decided block
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
// 传入的latestBlock应该是bc的最新区块
func (proofs *proofTable) Reset(moment Moment, latestBlock *defines.Block) {

	if !moment.Time.After(proofs.end.Time) {
		return
	}
	proofs.start = moment

	// 这种做法被取消，是因为这里只做了决定谁的区块被选择，但是这个所谓的decided可能没收到，也可能区块内容本身无效
	// 所以要考虑重新pot
	//if proofs.Decided != nil {	// 这代表上一轮竞赛自己有参与或见证，得到了最新区块
	//	proofs.baseIndex = proofs.Decided.BaseIndex + 1
	//	proofs.base = proofs.Decided.BlockHash
	//	proofs.HasLatestBlockNow = true
	//} else {	// 本地第一次使用proofTable，没有winner
	//	// 再看本地是否有最新区块。事实上对于“最新区块”而言，一定是先有pot.winner，再有最新区块。
	//	// 所以如果pot.winner == nil，那么其实本地也没有真正的最新区块
	//	proofs.HasLatestBlockNow = false
	//}

	if latestBlock == nil {		// 这说明本地不知道最新的区块是什么
		proofs.HasLatestBlockNow = false
	} else {
		proofs.HasLatestBlockNow = true
		proofs.baseIndex = latestBlock.Index + 1
		proofs.base = latestBlock.SelfHash
	}

	proofs.winner = nil
	proofs.Decided = nil
	proofs.Judged = nil
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
