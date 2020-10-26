/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 7:10 PM
* @Description: 处理Entry
***********************************************************************/

package pot

import (
	"bytes"
	"errors"

	"github.com/azd1997/blockchain-consensus/defines"
)

// 处理EntryBlocks
//
func (p *Pot) handleEntryBlock(ent *defines.Entry) error {
	// 检查ent.BaseIndex和Base
	process := p.processes.get(p.id)
	if ent.BaseIndex < process.Index {
		return nil
	}
	if ent.BaseIndex == process.Index && !bytes.Equal(ent.Base, process.Hash) {
		return errors.New("mismatched ent.BaseIndex or ent.Base")
	}
	// 解码
	block := new(defines.Block)
	err := block.Decode(ent.Data)
	if err != nil {
		return err
	}
	// 检查区块本身格式/签名的正确性
	err = block.Verify()
	if err != nil {
		return err
	}
	// 检查block与ent中携带的Base/BaseIndex信息是否一致
	if !bytes.Equal(block.PrevHash, ent.Base) || block.Index != ent.BaseIndex + 1 {
		return errors.New("mismatched block.PreHash or block.Index")
	}
	//// 如果其序号 = 本地的index+1，那么检查其有效性
	//if block.Index == ent.BaseIndex + 1 {
	//	// 将Block传递到检查器进行检查
	//
	//	// 检查通过后追加到本地分支
	//
	//} else if block.Index == ent.BaseIndex + 1 {
	//	// 如果序号是不连续的，暂时先保留在blocksCache
	//	p.addBlock(block)
	//}

	// 尝试添加到区块链中
	return p.bc.AddBlock(block)
}

func (p *Pot) handleEntryProof(ent *defines.Entry, from string) error {
	// 解码
	var proof Proof
	if err := proof.Decode(ent.Data); err != nil {
		return err
	}
	// 检查ent的Base信息是否合理
	process := p.processes.get(p.id)
	if ent.BaseIndex != process.Index || !bytes.Equal(ent.Base, process.Hash) {
		return errors.New("mismatched ent.Base or ent.BaseIndex")
	}
	// 如果证明信息有效，用以更新本地winner
	p.setProof(from,  proof)
	if proof.GreaterThan(p.proofs[p.winner]) {
		p.winner = from
	}
	return nil
}

// 处理新区块
// 新区块的话，得检查是否与之前的证明信息匹配
func (p *Pot) handleEntryNewBlock(ent *defines.Entry) error {
	block := new(defines.Block)

	// 解码
	if err := block.Decode(ent.Data); err != nil {
		return err
	}

	// 检查区块本身格式/签名的正确性
	if err := block.Verify(); err != nil {
		return err
	}

	// 检查区块是否是winner的，以及是否证明有效
	// TODO

	return p.bc.AddBlock(block)
}

// 处理交易
func (p *Pot) handleEntryTransaction(ent *defines.Entry) error {

	// 解码
	txbytes := ent.Data

	// 尝试添加到本地交易池
	return p.txPool.AddTransaction(txbytes)
}

// 处理邻居节点信息
// TODO: 考虑节点恶意
// 目前直接相信这个节点信息，添加到本地节点信息表
func (p *Pot) handleEntryNeighbor(ent *defines.Entry) error {
	pi := new(defines.PeerInfo)
	err := pi.Decode(bytes.NewReader(ent.Data))
	if err != nil {
		return err
	}
	return p.pit.Set(pi)
}

// 处理Process
func (p *Pot) handleEntryProcess(from string, ent *defines.Entry) error {
	p.processes.set(from, &defines.Process{
		Index:       ent.BaseIndex,
		Hash:        ent.Base,
		LatestMaker: string(ent.Data),
	})

	return nil
}

