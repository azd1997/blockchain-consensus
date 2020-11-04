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

// Proof 证明
type Proof struct {
	Id        string
	TxsNum    uint64 // 收集的交易数量
	TxsMerkle []byte // 收集的所有交易组织成的merkle树的根
	Base      []byte // 基于的区块的哈希
	BaseIndex uint64 // 基于的区块的序号
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

func (p *Proof) GreaterThan(ap *Proof) bool {
	// TODO
	return true
}

// Match 检查block和proof是否匹配
func (p *Proof) Match(block *defines.Block) bool {
	return p.Id == block.Maker &&
		p.TxsNum == uint64(len(block.Txs)) &&
		bytes.Equal(p.TxsMerkle, block.Merkle) &&
		p.BaseIndex == block.Index-1 &&
		bytes.Equal(p.Base, block.PrevHash)
}
