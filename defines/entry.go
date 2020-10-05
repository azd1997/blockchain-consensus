/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/21 1:42
* @Description: The file is for
***********************************************************************/

package defines

// EntryType 条目类型，没有做预定义，具体解释由各个实现自行定义、解析
type EntryType = uint8

const (
	EntryType_Block EntryType = 0
	EntryType_Proof EntryType = 1
)

type Entry struct {
	BaseIndex uint64	// 当前区块编号（高度），相当于任期
	Base []byte	// 当前消息构建时所基于的区块的Hash，当启用严格检查时，该项应被设置

	Type EntryType	// 指示Entry内存放的内容
	Data []byte		// 区块/证明/交易 等序列化的数据
}

