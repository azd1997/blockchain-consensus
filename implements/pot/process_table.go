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

// pnode 进度节点
// type pnode struct {

// }


// 进度表更新发生在decide(bn)(PotStart)，使用完旧的最新进度表之后就重置。然后自己率先更新自己的进度，接下来到PotOver，其他节点陆陆续续把自己进度发过来
// 更新了其他节点的进度
// 进度表记录的空洞，没了之后不代表进度完整，还要满足BC中没有离散区块
// 同样的，BC中没有离散区块也不足以充分说明进度完整。两个条件需要合一

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
	id string

	//counts map[uint64]map[string]map[string]*defines.Process	// <index, <hex_hash, <id, process> > >
	maxIndex int64 // 收到的最大的区块的索引，借助maxIndex和maxIndex-1可以得到latest/oldlatest

	// processes 存的就是最新进度。process和proof会合并起来在
	processes map[string]*defines.Process
	//latest    map[string]bool
	//oldlatest map[string]bool
	lock *sync.RWMutex

	//total      int // 统计的总节点数
	//totalAlive int // 进度在n或n-1的节点认为是在不断同步，其他可能是挂掉或刚起来，不算在alive内

	// 自己的进度是否存在空洞？空洞情况
	// “不重叠区间” 按left升序
	holes [][2]int64 // 空洞，还欠缺的区块区间 [2]uint64{left, right}，[left, right]这个区间的区块还没有同步到
}

// newProcessTable 新建进度表
func newProcessTable() *processTable {
	return &processTable{
		//counts: map[uint64]map[string]map[string]*defines.Process{},
		processes: make(map[string]*defines.Process),
		//latest:    make(map[string]bool),
		//oldlatest: make(map[string]bool),
		lock:  new(sync.RWMutex),
		holes: [][2]int64{},
	}
}

// 更新，将maxIndex刷新
// refresh的调用发生在自己决定了新区块的时候
func (pt *processTable) refresh(latestBlock *defines.Block) {
	if latestBlock.Index == pt.maxIndex {
		return
	}

	if latestBlock.Index == pt.maxIndex + 1 {
		pt.maxIndex++
		pt.lock.Lock()
		pt.processes = map[string]*defines.Process{}	// 刷新
		pt.processes[pt.id] = &defines.Process{
			Index:       latestBlock.Index,
			Hash:        latestBlock.SelfHash,
			LatestMaker: latestBlock.Maker,
			Id:          pt.id,
			NoHole:      len(pt.holes) == 0,
		}
		pt.lock.Unlock()
	}
}

// set 更新某个节点的进度
func (pt *processTable) set(id string, process *defines.Process) {
	pt.lock.Lock()
	// 更新process
	if pt.processes == nil || pt.processes[id] == nil ||
		process == nil || pt.processes[id].Index >= process.Index {
		return
	}
	pt.processes[id] = process
	pt.lock.Unlock()
}

// get 查询某个节点的进度
func (pt *processTable) get(id string) *defines.Process {
	pt.lock.RLock()
	defer pt.lock.RUnlock()
	// 更新process
	if pt.processes == nil || pt.processes[id] == nil {
		return &defines.Process{}
	} else {
		return pt.processes[id]
	}
}

// nLatestPeers 随机返回n个最新进度的节点的id
// 如果输入的n=0，则返回所有最新进度的节点id
// 如果输入的n比总的最新进度的节点数大，那么返回所有
func (pt *processTable) nLatestPeers(n int) []string {
	pt.lock.RLock()
	defer pt.lock.RUnlock()
	c := len(pt.processes)
	l := c // 返回的数量
	if n > 0 && n < c {
		l = n
	}
	all := make([]string, 0, c)
	for id := range pt.processes {
		all = append(all, id)
	}
	return all[:l]
}

// isLatest 检查某个节点是否是最新进度
// 注意：NoHole这项，通常不被使用到，因为非Ready状态的节点不能广播proof及process
func (pt *processTable) isLatest(id string) bool {
	pt.lock.RLock()
	defer pt.lock.RUnlock()
	p, ok := pt.processes[id]
	if !ok {
		return false
	}
	return p.Index == pt.maxIndex && p.NoHole // < 则不是最新； 不可能大于
}

// isSelfReady 判断自己是否准备好（所有区块都得到，并且紧跟最新进度）
func (pt *processTable) isSelfReady() bool {
	return pt.isLatest(pt.id) && len(pt.holes) == 0
}

// latest查询当前最新区块
func (pt *processTable) latest() {

}

// totalAlive 所有状况正常的节点数
func (pt *processTable) totalAlive() int {
	pt.lock.RLock()
	defer pt.lock.RUnlock()
	return len(pt.processes)
}

// fill 本机节点获得中间的区块，用以填补空缺 （fill hole）
func (pt *processTable) fill(bIndex int64) {
	// 首先通过二分查找定位到bIndex **可能** 属于哪一个hole（“区间”）
	mayIdx := binarySearch(pt.holes, bIndex)
	if mayIdx >= 0 { // 起码说明有意义
		hole := pt.holes[mayIdx]
		// [l, r]	bIndex可能在区间内或右
		if hole[0] == hole[1] && bIndex <= hole[1] { // 区间长度为1
			pt.holes = append(pt.holes[:mayIdx], pt.holes[mayIdx+1:]...)
			return
		}
		if bIndex == hole[0] {
			pt.holes[mayIdx] = [2]int64{hole[0] + 1, hole[1]}
		} else if bIndex == hole[1] {
			pt.holes[mayIdx] = [2]int64{hole[0], hole[1] - 1}
		} else if bIndex > hole[0] && bIndex < hole[1] {
			pt.holes[mayIdx] = [2]int64{hole[0], bIndex - 1}
			var holes [][2]int64
			holes = append(holes, pt.holes[:mayIdx+1]...)
			holes = append(holes, [2]int64{bIndex + 1, hole[1]})
			holes = append(holes, pt.holes[mayIdx+1:]...)
			pt.holes = holes
		}
	}
}

// binarySearch 找出所属区间的下标
func binarySearch(holes [][2]int64, target int64) int {
	l, r := 0, len(holes)-1
	for l <= r {
		mid := (l + r) / 2
		if holes[mid][0] > target { // l  t  mid   r
			r = mid - 1
		} else { // <= 		// l  mid t r
			l = mid + 1
		}
	}

	return r
}
