/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:47
* @Description: 公式需要的定义，现在不管结构体内容，先定义在此。之后考虑直接做类型重命名
***********************************************************************/

package defines

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"time"
)

//

// Block 区块
type Block struct {
	Index       int64
	Maker       string
	Timestamp   int64
	SelfHash    []byte
	PrevHash    []byte
	Merkle      []byte
	Txs         []*Transaction
	Description string
	Sig         []byte
}

// Encode 编码
func (b *Block) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(b)
	return buf.Bytes(), err
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

// Hash 为区块生成哈希或者查询其哈希
func (b *Block) Hash() error {
	if b == nil {
		return errors.New("nil block")
	}

	if b.SelfHash == nil {
		if b.Sig != nil {
			return errors.New("non-nil sig when hash")
		}
		blockBytes, err := b.Encode()
		if err != nil {
			return err
		}
		h := sha256.Sum256(blockBytes)
		b.SelfHash = h[:]
	}
	return nil
}

// MerkleTxs 为区块内包含的交易列表生成默克尔数根哈希
func (b *Block) MerkleTxs() error {
	return nil
}

// Key 区块的键
func (b *Block) Key() string {
	if b == nil || b.SelfHash == nil {
		return ""
	}
	return fmt.Sprintf("%x", b.SelfHash)
}

// ShortName 取区块哈希十六进制字符串的前6个字符作为短名
func (b *Block) ShortName() string {
	if k := b.Key(); k == "" {
		return ""
	} else {
		return k[:6]
	}
}

///////////////////////

// NewBlock 构造新区块
func NewBlockAndSign(index int64, id string, prevHash []byte, txs []*Transaction, desc string) (*Block, error) {
	b := &Block{
		Index:       index,
		Maker:       id,
		Timestamp:   time.Now().UnixNano(),
		SelfHash:    nil,
		PrevHash:    prevHash,
		Merkle:      nil,
		Txs:         txs,
		Description: desc,
		Sig:         nil,
	}
	// 先生成merkle
	if err := b.MerkleTxs(); err != nil {
		return nil, err
	}
	// 再生成selfhash
	if err := b.Hash(); err != nil {
		return nil, err
	}
	// 最后签名
	if err := b.Sign(); err != nil {
		return nil, err
	}

	return b, nil
}
