/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 20:37
* @Description: Proof
***********************************************************************/

package pot

import (
	"bytes"
	"encoding/gob"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/utils/bufferpool"
)

// 这里将证明的含义作下说明：本质上节点构造区块是在PotOver时，这时将区块哈希和区块包含的有效交易数量作为证明

// Proof 证明
type Proof struct {
	Id        string
	TxsNum    int64  // 收集的交易数量
	BlockHash []byte // 自己构造的区块的哈希
	Base      []byte // 基于的区块的哈希
	BaseIndex int64  // 基于的区块的序号
}

// Encode 编码
func (p *Proof) Encode() ([]byte, error) {
	buf := bufferpool.Get()
	defer bufferpool.Return(buf)
	err := gob.NewEncoder(buf).Encode(p)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode 解码
// p := new(Proof)
func (p *Proof) Decode(data []byte) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(p)
}

// GreaterThan 两个证明间的比较
func (p *Proof) GreaterThan(ap *Proof) bool {
	// TODO
	if ap == nil {
		return true
	}

	return true
}

// Match 检查block和proof是否匹配
func (p *Proof) Match(block *defines.Block) bool {
	return p.Id == block.Maker &&
		p.TxsNum == int64(len(block.Txs)) &&
		bytes.Equal(p.BlockHash, block.SelfHash) &&
		p.BaseIndex == block.Index-1 &&
		bytes.Equal(p.Base, block.PrevHash)
}
