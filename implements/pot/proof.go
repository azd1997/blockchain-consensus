/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 20:37
* @Description: Proof
***********************************************************************/

package pot

// Proof 证明
type Proof struct {
	//Base []byte	// 基于的区块的哈希
	//BaseIndex uint64	// 基于的区块的序号
	TxsNum uint64	// 收集的交易数量
	TxsMerkle []byte	// 收集的所有交易组织成的merkle树的根
}

func (p *Proof) Encode() []byte {
	// TODO
	return nil
}

// var p Proof
func (p *Proof) Decode() error {
	// TODO
	return nil
}

func (p *Proof) GreaterThan(ap *Proof) bool {
	// TODO
	return true
}