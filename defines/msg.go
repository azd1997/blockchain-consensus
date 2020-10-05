/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 20:45
* @Description: The file is for
***********************************************************************/

package defines

type MessageType uint8

const (
	MessageType_None MessageType = 0		// 啥也不干的消息

	MessageType_Data MessageType = 1		// 一般的数据传输的消息

	MessageType_Req MessageType = 2			// 请求类消息
)

type Message struct {
	Version Version
	Type MessageType
	From string
	To string

	Sig []byte	// 消息签名，以保证消息不被恶意篡改

	Entries []*Entry	// 待传送的条目

	Reqs []*Request
}


