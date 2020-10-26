/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:47
* @Description: 公式需要的定义，现在不管结构体内容，先定义在此。之后考虑直接做类型重命名
***********************************************************************/

package defines

import (
	"bytes"
	"encoding/gob"

	"github.com/azd1997/blockchain-consensus/utils/bufferpool"
)

//

// Block 区块
type Block struct {
	Index    uint64
	Maker string
	Timestamp uint64
	SelfHash []byte
	PrevHash []byte
	Merkle   []byte
	Txs      [][]byte
	Sig      []byte
}

// Encode 编码
func (b *Block) Encode() ([]byte, error) {
	buf := bufferpool.Get()
	defer bufferpool.Return(buf)
	err := gob.NewEncoder(buf).Encode(b)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode 解码
// b := new(Block)
func (b *Block) Decode(data []byte) error {
	r := bytes.NewReader(data)
	return gob.NewDecoder(r).Decode(b)
}

// Sign 签名
func (b *Block) Sign() error {
	return nil
}

// 验证基础格式与签名
func (b *Block) Verify() error {
	return nil
}


