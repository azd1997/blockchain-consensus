/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 7:09 PM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"fmt"
	"sync/atomic"
)

// 查看当前状态
func (p *Pot) getStage() StageType {
	return StageType(atomic.LoadUint32((*uint32)(&p.stage)))
}

// 更新当前状态
func (p *Pot) setStage(newStage StageType) {
	atomic.StoreUint32((*uint32)(&p.stage), uint32(newStage))
}

// 查看当前状态和duty
func (p *Pot) DutyStageState() string {
	return fmt.Sprintf("%s-%s-%s", p.duty.String(), p.getStage().String(), p.getState().String())
}

//// 获取某个id的进度
//func (p *Pot) getProcess(id string) defines.Process {
//	var process defines.Process
//	p.processesLock.RLock()
//	process = *(p.processes[id])
//	p.processesLock.RUnlock()
//	return process
//}
//
//// 设置某个id的进度
//func (p *Pot) setProcess(id string, process defines.Process) {
//	p.processesLock.Lock()
//	p.processes[id] = &process
//	p.processesLock.Unlock()
//}

// // 获取某个id的proof
// func (p *Pot) getProof(id string) Proof {
// 	var proof Proof
// 	p.proofsLock.RLock()
// 	proof = *(p.proofs[id])
// 	p.proofsLock.RUnlock()
// 	return proof
// }

// // 设置某个id的proof
// func (p *Pot) setProof(id string, proof Proof) {
// 	p.proofsLock.Lock()
// 	p.proofs[id] = &proof
// 	p.proofsLock.Unlock()
// }

//
//// 获取某个区块缓存
//func (p *Pot) getBlock(hash []byte) *defines.Block {
//	key := fmt.Sprintf("%x", hash)
//	var block *defines.Block
//	p.blocksLock.RLock()
//	block = p.blocksCache[key]
//	p.blocksLock.RUnlock()
//	return block
//}
//
//// 添加某个区块缓存
//func (p *Pot) addBlock(block *defines.Block) {
//	key := fmt.Sprintf("%x", block.SelfHash)
//	p.blocksLock.Lock()
//	p.blocksCache[key] = block
//	p.blocksLock.Unlock()
//}

///////////////////////////////////////////////////////

// 检查是否追上最新进度
//func (p *Pot) latest() bool {
//	return p.processes.isLatest(p.id)
//}

// 判断自身记录是否准备好（最新 + 完整）
func (p *Pot) isSelfReady() bool {
	// return p.processes.isSelfReady() && !p.bc.Discontinuous()
	return !p.bc.Discontinuous()
}

// Epoch 查看当前处于哪一个纪元
func (p *Pot) Epoch() int64 {
	return p.epoch
}

// NextEpoch 新纪元开启
func (p *Pot) NextEpoch() {
	p.epoch++
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}
