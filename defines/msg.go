/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 20:45
* @Description: The file is for
***********************************************************************/

package defines

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/azd1997/blockchain-consensus/utils/bufferpool"
)

const (
	MaxMessageLen             = 5 * 1024 * 1024 // 5MB
	MessageMagicNumber uint16 = 0xbcef          // 消息魔数，快速确定消息起始位置，校验协议是否匹配
)

type MessageType uint8

const (
	MessageType_None MessageType = 0 // 啥也不干的消息

	MessageType_Data MessageType = 1 // 一般的数据传输的消息

	MessageType_Req MessageType = 2 // 请求类消息
)

type Message struct {
	Version Version
	Type    MessageType

	Epoch uint64 // 纪元，指当前是基于哪一个区块的创建时间基准 从数量上等同于最新区块index

	From string
	To   string

	Entries []*Entry // 待传送的条目

	Reqs []*Request

	Desc string // 描述字段，必须是json序列化后的字符串，按照key:value的形式

	Sig []byte // 消息签名，以保证消息不被恶意篡改
}

// Len 获取Message序列化后的长度(包含魔数所占用的2B)
func (msg *Message) Len() int {
	length := 2 + 4 +
		1 + 1 + 8 +
		1 + (len(msg.From) + len(msg.To)) +
		1 + 1 + 2 + len(msg.Desc) + 2 + len(msg.Sig) // 还没加上Entries和Requests
	for _, ent := range msg.Entries {
		length += (2 + ent.Len())
	}
	for _, req := range msg.Reqs {
		length += (2 + req.Len())
	}
	return length
}

// Check 检查msg格式是否符合要求，允许序列化
func (msg *Message) Check() error {
	if msg == nil {
		return errors.New("nil Message")
	}

	if len(msg.From) != len(msg.To) {
		return errors.New("len(From) != len(To)")
	}

	if len(msg.Sig) == 0 {
		return errors.New("nil Sig")
	}

	return nil
}

// Encode 编码
func (msg *Message) Encode() ([]byte, error) {
	var err error

	// 检查msg格式是否有效
	if err = msg.Check(); err != nil {
		return nil, err
	}

	// 获取缓冲
	buf := bufferpool.Get()
	defer bufferpool.Return(buf)

	// 写入魔数 2B
	err = binary.Write(buf, binary.BigEndian, MessageMagicNumber)
	if err != nil {
		return nil, err
	}

	// 写入数据长度 4B
	length := uint32(msg.Len())
	err = binary.Write(buf, binary.BigEndian, length)
	if err != nil {
		return nil, err
	}

	/* 写入Message主体 */

	// 写入版本号 1B
	err = binary.Write(buf, binary.BigEndian, msg.Version)
	if err != nil {
		return nil, err
	}

	// 写入消息类型 1B
	err = binary.Write(buf, binary.BigEndian, msg.Type)
	if err != nil {
		return nil, err
	}

	// 写入Epoch 8B
	err = binary.Write(buf, binary.BigEndian, msg.Epoch)
	if err != nil {
		return nil, err
	}

	// 写入id长度 1B
	idlen := uint8(len(msg.From))
	err = binary.Write(buf, binary.BigEndian, idlen)
	if err != nil {
		return nil, err
	}

	// 写入From和To (2 * idlen)B
	// 注意binary.Write不能直接写string，需要转为[]byte，否则会报"invalid type string"
	err = binary.Write(buf, binary.BigEndian, []byte(msg.From))
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, []byte(msg.To))
	if err != nil {
		return nil, err
	}

	// 写入Entries长度和Requests长度 (2 * 1)B
	nEntry := uint8(len(msg.Entries))
	nRequest := uint8(len(msg.Reqs))
	err = binary.Write(buf, binary.BigEndian, nEntry)
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, nRequest)
	if err != nil {
		return nil, err
	}

	// 写入Entries
	for _, ent := range msg.Entries {
		// ent编码
		entBytes, err := ent.Encode()
		if err != nil {
			return nil, err
		}
		// 写入Entry长度 2B
		entlen := uint16(len(entBytes))
		err = binary.Write(buf, binary.BigEndian, entlen)
		if err != nil {
			return nil, err
		}
		// 写入Entry (entlen)B
		err = binary.Write(buf, binary.BigEndian, entBytes)
		if err != nil {
			return nil, err
		}
	}

	// 写入Reqs
	for _, req := range msg.Reqs {
		// req编码
		reqBytes, err := req.Encode()
		if err != nil {
			return nil, err
		}
		// 写入Req长度 2B
		reqlen := uint16(len(reqBytes))
		err = binary.Write(buf, binary.BigEndian, reqlen)
		if err != nil {
			return nil, err
		}
		// 写入Req (reqlen)B
		err = binary.Write(buf, binary.BigEndian, reqBytes)
		if err != nil {
			return nil, err
		}
	}

	// 写入Desc长度 2B
	desclen := uint16(len(msg.Desc))
	err = binary.Write(buf, binary.BigEndian, desclen)
	if err != nil {
		return nil, err
	}
	fmt.Printf("encode: desclen=%d\n", desclen)
	// 写入Desc (desclen)B
	err = binary.Write(buf, binary.BigEndian, []byte(msg.Desc))
	if err != nil {
		return nil, err
	}

	// 写入签名长度 2B
	siglen := uint16(len(msg.Sig))
	fmt.Printf("encode: siglen=%d\n", siglen)
	err = binary.Write(buf, binary.BigEndian, siglen)
	if err != nil {
		return nil, err
	}
	// 写入签名 (siglen)B
	err = binary.Write(buf, binary.BigEndian, msg.Sig)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decode 解码
// 调用Decode之前 msg := new(Message)  (注意: var msg *Message 得到的msg是nil，不能调用msg.Decode)
// 要注意的是：
// 		1. 传入的r是包含首部魔数和总长字段
// 		2. 长度为0的切片其值默认为nil，而非[]T{}
func (msg *Message) Decode(r io.Reader) error {

	// 读取魔数
	magic := uint16(0)
	err := binary.Read(r, binary.BigEndian, &magic)
	if err != nil {
		return err
	}
	if magic != MessageMagicNumber {
		return fmt.Errorf("wrong magic(%x)", magic)
	}

	// 读取总长度
	totallen := uint32(0)
	err = binary.Read(r, binary.BigEndian, &totallen)
	if err != nil {
		return err
	}
	totallen -= (2 + 4) // 减去魔数和自身
	if totallen <= 0 {
		return errors.New("not enough totallen")
	}

	// 读取版本号
	err = binary.Read(r, binary.BigEndian, &msg.Version)
	if err != nil {
		return err
	}
	//fmt.Printf("Decode Version: %d\n", msg.Version)
	totallen -= 1 // 减去Version
	if totallen <= 0 {
		return errors.New("not enough totallen")
	}

	// 读取消息类型
	err = binary.Read(r, binary.BigEndian, &msg.Type)
	if err != nil {
		return err
	}
	//fmt.Printf("Decode Type: %d\n", msg.Type)
	totallen -= 1 // 减去Type
	if totallen <= 0 {
		return errors.New("not enough totallen")
	}

	// 读取Epoch
	err = binary.Read(r, binary.BigEndian, &msg.Epoch)
	if err != nil {
		return err
	}
	totallen -= 8 // 减去Epoch
	if totallen <= 0 {
		return errors.New("not enough totallen")
	}

	// 读取ID长度
	idlen := uint8(0)
	err = binary.Read(r, binary.BigEndian, &idlen)
	if err != nil {
		return err
	}
	//fmt.Printf("Decode idlen: %d\n", idlen)
	totallen -= 1 // 减去idlen
	if totallen <= 0 {
		return errors.New("not enough totallen")
	}

	// 读取From/To
	fromto := make([]byte, idlen*2)
	err = binary.Read(r, binary.BigEndian, fromto)
	if err != nil {
		return err
	}
	msg.From = string(fromto[:idlen])
	msg.To = string(fromto[idlen:])
	//fmt.Printf("Decode From: %s\n", msg.From)
	//fmt.Printf("Decode To: %s\n", msg.To)
	totallen -= uint32(idlen * 2) // 减去idlen*2
	if totallen <= 0 {
		return errors.New("not enough totallen")
	}

	// 读取Entry数和Request数
	nEntry, nReq := uint8(0), uint8(0)
	err = binary.Read(r, binary.BigEndian, &nEntry)
	if err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &nReq)
	if err != nil {
		return err
	}
	totallen -= 2 // 减去2
	if totallen < 0 {
		return errors.New("not enough totallen")
	}
	//fmt.Printf("nEntry=%d, nReq=%d\n", nEntry, nReq)

	// 读取Entry
	if nEntry > 0 {
		msg.Entries = make([]*Entry, nEntry)
		entlen := uint16(0)
		for i := uint8(0); i < nEntry; i++ {
			// 读取entlen
			err = binary.Read(r, binary.BigEndian, &entlen)
			if err != nil {
				return err
			}
			// 读取ent
			ent := new(Entry)
			err = ent.Decode(r)
			if err != nil {
				return err
			}
			msg.Entries[i] = ent

			totallen -= uint32(2 + entlen)
			if totallen < 0 {
				return errors.New("not enough totallen")
			}
		}
	}

	// 读取Reqs
	if nReq > 0 {
		msg.Reqs = make([]*Request, nReq)
		reqlen := uint16(0)
		for i := uint8(0); i < nReq; i++ {
			// 读取reqlen
			err = binary.Read(r, binary.BigEndian, &reqlen)
			if err != nil {
				return err
			}
			// 读取req
			req := new(Request)
			err = req.Decode(r)
			if err != nil {
				return err
			}
			msg.Reqs[i] = req

			totallen -= uint32(2 + reqlen)
			if totallen < 0 {
				return errors.New("not enough totallen")
			}
		}
	}

	// 读取Desc长度
	desclen := uint16(0)
	err = binary.Read(r, binary.BigEndian, &desclen)
	if err != nil {
		return err
	}
	//fmt.Printf("desclen=%d\n", desclen)
	totallen -= 2
	if totallen < 0 {
		return errors.New("not enough totallen")
	}
	//fmt.Printf("totallen=%d\n", totallen)
	// 读取Desc
	descbytes := make([]byte, desclen)
	err = binary.Read(r, binary.BigEndian, descbytes)
	if err != nil {
		return err
	}
	totallen -= uint32(desclen)
	if totallen < 0 {
		return errors.New("not enough totallen")
	}
	msg.Desc = string(descbytes)

	// 读取签名长度
	siglen := uint16(0)
	err = binary.Read(r, binary.BigEndian, &siglen)
	if err != nil {
		return err
	}
	totallen -= 2
	if totallen < 0 {
		return errors.New("not enough totallen")
	}
	// 读取签名
	msg.Sig = make([]byte, siglen)
	err = binary.Read(r, binary.BigEndian, msg.Sig)
	if err != nil {
		return err
	}
	totallen -= uint32(siglen)
	if totallen < 0 {
		return errors.New("not enough totallen")
	}

	return nil
}

/*
	序列化后的Message格式：

	+--------------------------------------+
	| 魔数(2B) | 消息总长(4B) | 消息主体(变长) |
	+--------------------------------------+

	消息主体的格式：
	消息主体直接采用gob编码。使用binary编码的话随着成员变量的复杂提升，写起来比较啰嗦
	gob无痛序列化.
	gob的问题在于没办法在序列化之前得知消息的长度，用于消息长度控制
	尤其是可能需要不断检测序列化后长度
	所以还是得用binary编码

	+--------------------------------------+
	| 版本(1B)  |  消息类型(1B)  | Epoch(8B) |
	+--------------------------------------+
	|  ID长度(1B)  |  发送方ID  |  发送方ID   |
	+--------------------------------------+
	|   Entry数量(1B)    | Request数量(1B)   |
	+--------------------------------------+
	|   Entry1长度(2B)   |     Entry1       |
	+--------------------------------------+
	|        ...        |       ...        |	// Entry列表		// TODO: Entry 2B长度(65536B)似乎不够用
	+--------------------------------------+
	|   Entryi长度(2B)   |     Entryi       |
	+--------------------------------------+
	|  Request1长度(2B)  |    Request1      |
	+--------------------------------------+
	|        ...        |       ...        |	// Request列表
	+--------------------------------------+
	|  Requesti长度(2B)  |    Requesti      |
	+--------------------------------------+
	|  Requesti长度(2B)  |    Requesti      |
	+--------------------------------------+
	|   len(Desc)(2B)   |       Desc       |	// 描述字段，用来传递额外信息
	+--------------------------------------+
	|    签名长度(2B)    |     发送方签名     |
	+--------------------------------------+

	记消息长度(除)的预计算公式为 f(T), 则
	f(Message) = 2 + 4 + 1 + 1 + 8 + 1 + 2*idlen + 1 + 1 + nEntry * (2 + f(Entry)) + nRequest * (2 + f(Request)) + 2 + siglen
			   =

*/

// TODO

// Sign 生成签名
func (msg *Message) Sign() error {
	msg.Sig = []byte("signature")
	return nil
}

// Verify 验证基础格式与签名
func (msg *Message) Verify() error {
	if bytes.Equal(msg.Sig, []byte("signature")) {
		return nil
	}
	return errors.New("verify sig fail")
}

// String 字符串表示
func (msg *Message) String() string {
	b, err := json.Marshal(msg)
	if err != nil {
		return err.Error()
	}
	return string(b)
}
