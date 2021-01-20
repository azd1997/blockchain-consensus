/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 20:41
* @Description: 状态机状态
***********************************************************************/

package pot

// StageType pot共识中的阶段类型
type StageType uint32

const (

	// StateType_PreInited (初始化之前)的状态，启动之后到进入到NotReady之前的阶段
	StageType_PreInited_RequestNeighbors StageType = iota
	StageType_PreInited_RequestFirstBlock
	//StageType_PreInited_RequestLatestBlock
	//// StateType_NotReady 状态: 区块链有缺失或者网络中能正常连接且均无缺失的节点>=3个
	// NotReady仅仅用于在启动结束时的一个较短过程
	//StateType_NotReady
	// StageType_InPot Pot竞赛的阶段，位于PotStart到PotOver之间
	StageType_InPot
	// StageType_PostPot Pot竞赛结束之后的阶段，位于PotOver到PotStart之间
	StageType_PostPot
)

var stageMap = map[StageType]string{
	//StateType_Init_GetNeighbors: "[State_Init_GetNeighbors]",
	//StateType_Init_GetProcesses: "[StateType_Init_GetProcesses]",
	//StateType_Init_GetLatestBlock:    "[StateType_Init_GetLatestBlock]",
	//StateType_NotReady:          "[State_NotReady]",
	//StateType_ReadyCompete:      "[State_ReadyCompete]",
	//StateType_Competing:         "[State_Competing]",
	//StateType_CompeteOver:       "[State_CompeteOver]",
	//StateType_CompeteWinner:     "[State_CompeteWinner]",
	//StateType_CompeteLoser:      "[State_CompeteLoser]",

	StageType_PreInited_RequestNeighbors:   "[Stage_PreInited_RN]",
	StageType_PreInited_RequestFirstBlock:  "[Stage_PreInited_RFB]",
	//StageType_PreInited_RequestLatestBlock: "[Stage_PreInited_RLB]",
	//StateType_NotReady:                     "[State_NotReady]",
	StageType_InPot:   "[Stage_InPot]",
	StageType_PostPot: "[Stage_PostPot]",
}

func (st StageType) String() string {
	if v, ok := stageMap[st]; ok {
		return v
	} else {
		return "[Stage_Unknown]"
	}
}
