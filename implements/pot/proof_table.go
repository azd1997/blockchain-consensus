/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/18/20 2:49 PM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"bytes"
	"sync"
)

type proofTable struct {
	baseIndex int64	// 基于哪个区块开始的竞争
	base []byte		//
	table map[string]*Proof	// <id, *Proof>
	sync.RWMutex	// 保护table
	winner *Proof	// 胜者
}

func (proofs *proofTable) Add(p *Proof) {
	if p.BaseIndex != proofs.baseIndex || !bytes.Equal(p.Base, proofs.base) {
		return 		// TODO: 是否返回错误，让上层回复来信方？
	}

	proofs.Lock()
	proofs.table[p.Id] = p
	proofs.Unlock()

	if p.GreaterThan(proofs.winner) {
		proofs.winner = p
	}
}

// Winner 获胜者的proof
func (proofs *proofTable) Winner() *Proof {
	return proofs.winner
}

// Reset
func (proofs *proofTable) Reset() {
	if proofs.winner == nil {
		return 	// 重置失败
	}
	proofs.baseIndex = proofs.winner.BaseIndex+1
	proofs.base = proofs.winner.BlockHash
	proofs.winner = nil
	proofs.Lock()
	proofs.table = map[string]*Proof{}
	proofs.Unlock()
}

// 使用最新的区块去创建证明表
func newProofTable(latestBlockIndex int64, latestBlockHash []byte) *proofTable {
	return &proofTable{
		baseIndex: latestBlockIndex,
		base:      latestBlockHash,
		table: map[string]*Proof{},
	}
}
