/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/14/20 10:29 AM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"fmt"
	"sort"

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

	//moment Moment

	//baseBlock *defines.Block

	table          map[string]*undecidedBlock // <hash_hex, v>
}

// Add 添加未决区块
func (udbt *undecidedBlockTable) Add(b *defines.Block) {

	if b == nil {
		return
	}

	fmt.Printf("udbt add: ub(%s)=%v\n", b.ShortName(), b)

	//if udbt.baseBlock != nil && b.Index != udbt.baseBlock.Index + 1 {
	//	return
	//}

	if udbt.undecidedIndex > 0 && b.Index != udbt.undecidedIndex {
		return
	}

	k := b.Key()
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
	ub := udbt.table[k]
	if ub == nil {
		return nil
	}
	return ub.b
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
	if maxk  == "" {
		return nil
	}
	return udbt.table[maxk].b
}

// Reset latestBlock指目前本地有的最新的区块
// 若latestBlock为nil，说明本地没有最新区块
// 有的话，需要校验
func (udbt *undecidedBlockTable) Reset(undecidedIndex int64) {
	udbt.undecidedIndex = undecidedIndex
	udbt.table = map[string]*undecidedBlock{}
}

// 展示当前情况
func (udbt *undecidedBlockTable) Display() string {
	var str string
	str = fmt.Sprintf("udbt(%d): { ", udbt.undecidedIndex)
	ubs := make([]*undecidedBlock, 0, len(udbt.table))
	for k := range udbt.table {
		ubs = append(ubs, udbt.table[k])
	}
	sort.Slice(ubs, func(i, j int) bool {
		return ubs[i].count > ubs[j].count
	})
	for i:=0; i<len(ubs); i++ {
		substr := fmt.Sprintf("%s(%s,%d) ", ubs[i].b.ShortName(), ubs[i].b.Maker, ubs[i].count)
		str += substr
	}
	str += "}\n"
	return str
}