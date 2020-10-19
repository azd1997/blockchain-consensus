/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/30 17:43
* @Description: The file is for
***********************************************************************/

package bcc

import "github.com/azd1997/blockchain-consensus/requires"

// ConsensusLog 共识状态机日志持久化
type ConsensusLog struct {
	bc requires.BlockChain
	// 其他字段
}

func (cl *ConsensusLog) ReadConsensusState() []byte {
	return []byte{}
}

func (cl *ConsensusLog) SaveConsensusState(clog []byte) {

}

func (cl *ConsensusLog) Copy() *ConsensusLog {
	ncl := new(ConsensusLog)
	// 复制

	return ncl
}
