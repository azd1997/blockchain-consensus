/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:18
* @Description: requires 包含区块链共识需要实现的一些接口。
***********************************************************************/

package requires

import "github.com/azd1997/blockchain-consensus/defines"

// 以下这些接口是实现一个支持共识协议的完整区块链必须的一些模块
//

// 本地可直接处理的部分直接组合进Consensus接口
// 网络通信部分需要借助Transport的则和Consensus并行，通过channel双向通信


// BlockChain 区块链接口，包括内存存储及持久化相关的内容，逻辑上是一条哈希链式结构
type BlockChain interface {

}

// TransactionPool 交易池接口，其内部实现必须提供UBTXP,TBTXP,UCTXP这三类交易池
type TransactionPool interface {
	GenBlock() *defines.Block
}

// Validator 本地验证器，负责验证账户/区块/交易/证明的有效性
type Validator interface {

}

// Transport 传输器。用以支持传输协议的替换
type Transport interface {

}

//////////////////////////////////////////////////


