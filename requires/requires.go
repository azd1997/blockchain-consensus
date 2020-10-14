/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:18
* @Description: requires 包含区块链共识bcc的外部调用方需要实现的一些接口。
***********************************************************************/

package requires

import (
	"log"
	"net"

	"github.com/azd1997/blockchain-consensus/defines"
)

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
	validate(data []byte) *ValidateResult
}

// ValidateResult 验证结果
type ValidateResult struct {
	validate bool
	reason []byte	// 无效的原因
}

// Dialer 连接器
// 由于Conn是两个节点间的概念，并不是单例模式，不太适合直接作为接口进行替换
// 因此，为了支持外部自定义“连接”，本库对外提供Dialer接口，Conn的话则直接使用net.Conn接口
// 外部需要自定义Conn时，只需要自定义Dialer，自定义conn实现net.Conn接口
type Dialer interface {
	// Dial 发起连接
	// 需要注意的是：必须向对方传输自己的id
	// 假如改造TCP协议的话，可以在TCP包头添加自定义的ID字段，以表明身份
	// 如果不改造TCP协议的话，则可以在建立连接之后，要求链接发起方必须立马发送一个数据包，
	// 其内包含自身的id、签名等信息
	// 实际实现的Dialer必须提供localId，localListenAddr
	Dial(address, remoteId string) (Conn, error)
	Network() string	// 返回基于的网络协议
	LocalID() string
	LocalListenAddr() net.Addr
}

// Listener 外部也需要传入Listener监听器，以使共识节点有能力建立连接
type Listener interface {
	Network() string
	Accept() (Conn, error)
	Close() error
	LocalID() string
	LocalListenAddr() net.Addr
}

// Conn 连接
// Conn代表的是原生的Conn连接(无论是朴素的tcp还是其他自定义的“连接”)
// 但是Conn的实现结构必须包含连接双方节点的id
// bcc.Conn则是在requires.Conn上的进一步封装
type Conn interface {
	net.Conn
	Network() string	// 网络协议
	LocalID() string	// 获取自己的ID
	RemoteID() string	// 获取对端的ID
	LocalListenAddr() net.Addr	// 获取自己的监听地址
	RemoteListenAddr() net.Addr   // 获取对方的监听地址
}

//////////////////////////////////////////////////

// Store KV存储接口
// 实现时，不管到底是通过内存实现，还是数据库，对于bcc库而言，只按Store使用
type Store interface {
	Open() error	// 开启（其实是建立连接）
	Close() error	// 关闭（其实是关闭连接）
	Get(cf CF, key []byte) ([]byte, error)
	Set(cf CF, key, value []byte) error
	Del(cf CF, key []byte) error

	// Store的上层使用者需要有各自的前缀，
	// 注册Prefix时若报prefix已存在或不可用，则换个前缀使用
	// 通常在Store的使用者启动时，通过该API检查
	// prefix固定为2个字节
	//
	// CF ColumnFamily 列族，相当于“Database”的概念
	// 实现时是通过key前缀实现的
	// cf有长度限制，约定最大为6(6B表示的可能数足够多了)，长度不足将用" "补全
	// cf必须是连续字符
	RegisterCF(cf CF) error

	// 遍历某个CF
	RangeCF(cf CF, f func(key, value []byte) error) error
}

const CFLen = 6

// CF 列族名
type CF [CFLen]byte

// String2CF 字符串转换为列族名
// 由String2CF的空格补全知，自定义的CF名不能以' '结尾
func String2CF(in string) CF {
	n := len(in)
	if n == 0 || n > CFLen || in[n-1] == ' ' {
		log.Fatalln("incompatible cf string")
	}
	cf := CF{}
	for i:=0; i<len(in); i++ {
		cf[i] = in[i]
	}
	for i:=len(in); i<CFLen; i++ {
		cf[i] = ' '
	}
	return cf
}

