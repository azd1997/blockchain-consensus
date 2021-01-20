/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 20:41
* @Description: 状态机状态
***********************************************************************/

package pot

// StateType pot共识中的阶段类型
type StateType uint32

const (

	// StateType_PreInited (初始化之前)的状态，启动之后到进入到NotReady之前的阶段
	StateType_PreInited_RequestNeighbors StateType = iota
	StateType_PreInited_RequestFirstBlock
	StateType_PreInited_RequestLatestBlock
	// StateType_NotReady 状态: 区块链有缺失或者网络中能正常连接且均无缺失的节点>=3个
	StateType_NotReady
	// StateType_InPot Pot竞赛的阶段，位于PotStart到PotOver之间
	StateType_InPot
	// StateType_PostPot Pot竞赛结束之后的阶段，位于PotOver到PotStart之间
	StateType_PostPot
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

	StateType_PreInited_RequestNeighbors:   "[State_PreInited_RN]",
	StateType_PreInited_RequestFirstBlock:  "[State_PreInited_RFB]",
	StateType_PreInited_RequestLatestBlock: "[State_PreInited_RLB]",
	StateType_NotReady:                     "[State_NotReady]",
	StateType_InPot:                        "[State_InPot]",
	StateType_PostPot:                      "[State_PostPot]",
}

func (st StateType) String() string {
	if v, ok := stateMap[st]; ok {
		return v
	} else {
		return "[State_Unknown]"
	}
}
