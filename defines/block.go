/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:47
* @Description: 公式需要的定义，现在不管结构体内容，先定义在此。之后考虑直接做类型重命名
***********************************************************************/

package defines

//

// Block 区块
type Block struct {
	Index uint64
	SelfHash []byte
	PrevHash []byte
	Merkle []byte
	Txs [][]byte
	Sig []byte
}

func (b *Block) Encode() []byte {
	return nil	// TODO
}

// Transaction 交易
type Transaction struct {

}