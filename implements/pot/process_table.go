/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/20/20 1:05 PM
* @Description: 节点进度表
***********************************************************************/

package pot

import (
	"sync"

	"github.com/azd1997/blockchain-consensus/defines"
)

// 一轮决胜负：
// 每个节点只会遵从符合条件的/大多数选择的区块进度
// 按照pot竞争，在竞争阶段，就选定了自身节点能接受的新区块，
// 但是：由于节点所收集的证明消息可能有偏差，可能存在网络之中有多个新区块被若干个节点接受
// 也就是网络中节点们产生了分歧，在收集众节点进度时，同一个进度index下会有不同的hash，
//
// 1. proof开始
// 2. proof结束，出区块/等区块
// 3.

type pnode struct {

}

// 进度表
// 使用两个哈希表实现：
//		1. 快速查询修改某id节点的进度
//		2. 快速获取随机若干个最新进度的节点
// 		3. 快速判断自身是否跟上最新进度
// 还有一个需求：
//		所谓latest可能出现新latest与旧latest的问题，需要三个哈希表
// n -> n-1 -> n-2 -> m
//	 \    \      \     \
//	ids   ids    ids    ids

type processTable struct {
	processes map[string]*defines.Process
	latest map[string]bool
	oldlatest map[string]bool
	lock *sync.RWMutex

	total int	// 统计的总节点数
	totalAlive int	// 进度在n或n-1的节点认为是在不断同步，其他可能是挂掉或刚起来，不算在alive内

}

func newProcessTable() *processTable {
	return &processTable{
		processes: make(map[string]*defines.Process),
		latest:    make(map[string]bool),
		oldlatest: make(map[string]bool),
		lock:      new(sync.RWMutex),
	}
}

// 更新某个节点的进度
func (pt *processTable) set(id string, process *defines.Process) {
	pt.lock.Lock()
	// 更新process
	if pt.processes == nil || pt.processes[id] == nil ||
			process == nil || pt.processes[id].Index >= process.Index {
		return
	}
	pt.processes[id] = process
	// 是否要更新latest
	pt.lock.Unlock()
}

// 查询某个节点的进度
func (pt *processTable) get(id string) *defines.Process {

}

// 随机返回n个最新进度的节点的id
// 如果输入的n=0，则返回所有最新进度的节点id
// 如果输入的n比总的最新进度的节点数大，那么返回所有
func (pt *processTable) nLatestPeers(n int) []string {
	var res []string

	return res
}

// 检查某个节点是否是最新进度
func (pt *processTable) isLatest(id string) bool {

	return false
}