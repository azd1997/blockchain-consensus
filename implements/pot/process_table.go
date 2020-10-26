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


// 进度表
// 使用两个哈希表实现：
//		1. 快速查询修改某id节点的进度
//		2. 快速获取随机若干个最新进度的节点
// 		3. 快速判断自身是否跟上最新进度
// 还有一个需求：
//		所谓latest可能出现新latest与旧latest的问题，需要三个哈希表
type processTable struct {
	processes map[string]*defines.Process
	latest map[string]bool
	oldlatest map[string]bool
	lock *sync.RWMutex
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