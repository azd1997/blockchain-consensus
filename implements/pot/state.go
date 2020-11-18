/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 20:41
* @Description: 状态机状态
***********************************************************************/

package pot

type StateType uint32

const (

	//// 此状态为节点启动时需要向已有的节点（数量为n，当前值）请求节点表
	//// 当n个Neighbors消息到达之后（或者超时），进入下个阶段NotReady，追赶进度
	//StateType_Init_GetNeighbors StateType = iota
	//
	//// 此状态为节点启动时需要向已有的节点（数量为n，当前值）请求进度
	//// 当n个进度消息到达之后（或者超时），进入下个阶段NotReady，追赶进度
	//StateType_Init_GetProcesses
	//
	//// 此状态为节点启动时需要向已有的节点（数量为n，当前值）请求区块
	//// 当追赶进度至最新后，开始切换为ReadyCompete状态
	//StateType_Init_GetLatestBlock
	//
	//// 进度区块链进度没有和邻居节点们保持一致（达到最新），称为“NotReady”
	//// NotReady时不能参与竞争，只能等待新区块
	//StateType_NotReady
	//
	//// 同步到最新进度，可以参与竞争
	//StateType_ReadyCompete
	//
	//// 竞争
	//StateType_Competing
	//
	//// 竞争结束，结果出来，是胜者或者不是胜者
	//StateType_CompeteOver
	//
	//// 竞赛胜者，需要出块
	//StateType_CompeteWinner // 广播完新区块后切换为ReadyCompete
	//
	//// 竞赛负者，需要等待新区块
	//StateType_CompeteLoser // 收到新区块无误后切换为ReadyCompete，否则切换为Competing，重新竞争

	// seed状态

	// PreInited(初始化之前)的状态，启动之后到进入到NotReady之前的阶段
	StateType_PreInited StateType = 10
	// NotReady状态: 区块链有缺失或者网络中能正常连接且均无缺失的节点>=3个
	StateType_NotReady StateType = 11
	// Pot竞赛的阶段，位于PotStart到PotOver之间
	StateType_InPot StateType = 12
	// Pot竞赛结束之后的阶段，位于PotOver到PotStart之间
	StateType_PostPot StateType = 13
)

var stateMap = map[StateType]string{
	//StateType_Init_GetNeighbors: "[State_Init_GetNeighbors]",
	//StateType_Init_GetProcesses: "[StateType_Init_GetProcesses]",
	//StateType_Init_GetLatestBlock:    "[StateType_Init_GetLatestBlock]",
	//StateType_NotReady:          "[State_NotReady]",
	//StateType_ReadyCompete:      "[State_ReadyCompete]",
	//StateType_Competing:         "[State_Competing]",
	//StateType_CompeteOver:       "[State_CompeteOver]",
	//StateType_CompeteWinner:     "[State_CompeteWinner]",
	//StateType_CompeteLoser:      "[State_CompeteLoser]",
}

func (st StateType) String() string {
	if v, ok := stateMap[st]; ok {
		return v
	} else {
		return "[State_Unknown]"
	}
}
