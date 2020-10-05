/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 20:41
* @Description: 状态机状态
***********************************************************************/

package pot

type StateType uint32

const (

	// 进度区块链进度没有和邻居节点们保持一致（达到最新），称为“NotReady”
	// NotReady时不能参与竞争，只能等待新区块
	StateType_NotReady = 0

	// 同步到最新进度，可以参与竞争
	StateType_ReadyCompete = 1

	// 竞争
	StateType_Competing = 2

	// 竞争结束，结果出来，是胜者或者不是胜者
	StateType_CompeteOver = 3

	// 竞赛胜者，需要出块
	StateType_CompeteWinner = 4		// 广播完新区块后切换为ReadyCompete

	// 竞赛负者，需要等待新区块
	StateType_CompeteLoser = 5		// 收到新区块无误后切换为ReadyCompete，否则切换为Competing，重新竞争

	//
)


