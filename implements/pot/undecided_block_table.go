/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/14/20 10:29 AM
* @Description: The file is for
***********************************************************************/

package pot

// 未决区块表
// 两种情境下使用：
//	1. 启动时收集最新区块
//  2. 正常运行周期中接收新区块
// 两种情况下均是其他途径没法确定谁是正确的（或者说必要的检验都通过了），
// 投入该表来进行多数获胜。
// 该表的生命周期为一次收集期，之后将会重置以给下次使用
type undecidedBlockTable struct {
	tables map[uint64]map[string]
}
