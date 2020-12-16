/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 11:35 AM
* @Description: The file is for
***********************************************************************/

package defines

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
)

// Transaction 交易
type Transaction struct {
	TxHash      []byte
	From        string
	To          string
	Amount      int64
	Fields      map[string][]byte // 交易过程中的一些必要的字段，由业务侧解释。 []byte表示范围最广，数字也可以用其表示。 约定第1个字节标志后面的字节如何翻译
	Sig         []byte
	Description string // 一般性的描述
}

func (tx *Transaction) Key() string {
	if tx == nil || tx.TxHash == nil {
		return ""
	}
	return fmt.Sprintf("%x", tx.TxHash)
}

// ShortName 取区块哈希十六进制字符串的前6个字符作为短名
func (tx *Transaction) ShortName() string {
	if k := tx.Key(); k == "" {
		return ""
	} else {
		return k[:6]
	}
}

// Encode 编码
func (tx *Transaction) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(tx)
	return buf.Bytes(), err
}

// Decode 解码
// tx := new(Transaction)
func (tx *Transaction) Decode(data []byte) error {
	r := bytes.NewReader(data)
	return gob.NewDecoder(r).Decode(tx)
}

// Sign 签名
func (tx *Transaction) Sign() error {
	return nil
}

// 验证基础格式与签名
func (tx *Transaction) Verify() error {
	return nil
}

// Hash 为区块生成哈希或者查询其哈希
func (tx *Transaction) Hash() error {
	if tx == nil {
		return errors.New("nil block")
	}

	if tx.TxHash == nil {
		if tx.Sig != nil {
			return errors.New("non-nil sig when hash")
		}
		txBytes, err := tx.Encode()
		if err != nil {
			return err
		}
		h := sha256.Sum256(txBytes)
		tx.TxHash = h[:]
	}
	return nil
}

///////////////////////

// NewTransaction 构造新区块
func NewTransactionAndSign(from, to string, amount int64, fields map[string][]byte, description string) (*Transaction, error) {
	tx := &Transaction{
		TxHash:      nil,
		From:        from,
		To:          to,
		Amount:      amount,
		Fields:      fields,
		Sig:         nil,
		Description: description,
	}
	// 生成selfhash
	if err := tx.Hash(); err != nil {
		return nil, err
	}
	// 最后签名
	if err := tx.Sign(); err != nil {
		return nil, err
	}

	return tx, nil
}
