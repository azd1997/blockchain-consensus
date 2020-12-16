/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/14/20 10:29 AM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"fmt"

	"github.com/azd1997/blockchain-consensus/defines"
)

// newUndecidedBlockTable 新建未决区块表
func newUndecidedBlockTable() *undecidedBlockTable {
	return &undecidedBlockTable{table: map[string]*undecidedBlock{}}
}

// undecidedBlock 未决区块
type undecidedBlock struct {
	count int
	b     *defines.Block
}

// 未决区块表
// 两种情境下使用：
//	1. 启动时收集最新区块
//  2. 正常运行周期中接收新区块
// 两种情况下均是其他途径没法确定谁是正确的（或者说必要的检验都通过了），
// 投入该表来进行多数获胜。
// 该表的生命周期为一次收集期，之后将会重置以给下次使用
type undecidedBlockTable struct {
	undecidedIndex int64
	table          map[string]*undecidedBlock // <hash_hex, v>
}

// Add 添加未决区块
func (udbt *undecidedBlockTable) Add(b *defines.Block) {
	if udbt.undecidedIndex > 0 && b.Index != udbt.undecidedIndex {
		return
	}

	k := fmt.Sprintf("%x", b.SelfHash)
	if udbt.table[k] == nil {
		udbt.table[k] = &undecidedBlock{
			count: 1,
			b:     b,
		}
	} else {
		udbt.table[k].count++
	}
}

// Get 根据区块的哈希查询区块
func (udbt *undecidedBlockTable) Get(bhash []byte) *defines.Block {
	k := fmt.Sprintf("%x", bhash)
	return udbt.table[k].b
}

// Major 判定多数的区块
// 仅在需要依赖多数获胜原则时使用
func (udbt *undecidedBlockTable) Major() *defines.Block {
	maxcount := 0
	maxk := ""
	for k, udb := range udbt.table {
		if udb.count > maxcount {
			maxcount = udb.count
			maxk = k
		}
	}
	return udbt.table[maxk].b
}

// Reset index指此次收集的未决区块索引，若为0，表示不确定; 若为负数，也是不确定
// -1表示最新区块
func (udbt *undecidedBlockTable) Reset(index int64) {
	udbt.undecidedIndex = index
	udbt.table = map[string]*undecidedBlock{}
}
