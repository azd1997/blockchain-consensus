/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/21 1:42
* @Description: The file is for
***********************************************************************/

package defines

import (
	"bytes"
	"encoding/binary"
	"io"
)

// EntryType 条目类型，没有做预定义，具体解释由各个实现自行定义、解析
type EntryType uint8

const (
	EntryType_Block       EntryType = 0 // 区块同步 Base BaseIndex Type Data
	EntryType_Proof       EntryType = 1 // 证明 Type Data (proof本身包含了Base/BaseIndex信息)
	EntryType_NewBlock    EntryType = 2 // 新区块 Base BaseIndex Type Data
	EntryType_Transaction EntryType = 3 // 交易	Type Data
	EntryType_Neighbor    EntryType = 4 // 邻居节点信息	Type Data
	EntryType_Process     EntryType = 5 // 进度	Type Data
)

func (et EntryType) String() string {
	switch et {
	case EntryType_Block:
		return "EntryBlock"
	case EntryType_Proof:
		return "EntryProof"
	case EntryType_NewBlock:
		return "EntryNewBlock"
	case EntryType_Transaction:
		return "EntryTransaction"
	case EntryType_Neighbor:
		return "EntryNeighbor"
	case EntryType_Process:
		return "EntryProcess"
	default:
		return "EntryUnknown"
	}
}

// Entry 条目
type Entry struct {
	BaseIndex int64  // 当前区块编号（高度），相当于任期
	Base      []byte // 当前消息构建时所基于的区块的Hash，当启用严格检查时，该项应被设置

	Type EntryType // 指示Entry内存放的内容
	Data []byte    // 区块/证明/交易 等序列化的数据
}

// Check 检查格式
func (e *Entry) Check() error {
	return nil
}

// Len 获取序列化后的长度
func (e *Entry) Len() int {
	return 8 + 2 + len(e.Base) + 1 + 4 + len(e.Data)
}

// Encode 编码
func (e *Entry) Encode() ([]byte, error) {
	var err error

	// 检查格式
	if err = e.Check(); err != nil {
		return nil, err
	}

	// 获取缓冲
	buf := new(bytes.Buffer)

	//fmt.Println("buf: ", buf.Bytes())
	//fmt.Println("e.Data: ", e.Data)

	/* 序列化Entry */

	// BaseIndex 8B
	err = binary.Write(buf, binary.BigEndian, e.BaseIndex)
	if err != nil {
		return nil, err
	}

	//fmt.Println("buf: ", buf.Bytes())
	//fmt.Println("e.Data: ", e.Data)

	// baselen 2B
	baselen := uint16(len(e.Base))
	err = binary.Write(buf, binary.BigEndian, baselen)
	if err != nil {
		return nil, err
	}

	//fmt.Println("buf: ", buf.Bytes())
	//fmt.Println("e.Data: ", e.Data)

	if baselen > 0 {
		// Base (baselen)B
		err = binary.Write(buf, binary.BigEndian, e.Base)
		if err != nil {
			return nil, err
		}
	}

	//fmt.Println("buf: ", buf.Bytes())
	//fmt.Println("e.Data: ", e.Data)

	// Type 1B
	err = binary.Write(buf, binary.BigEndian, e.Type)
	if err != nil {
		return nil, err
	}

	//fmt.Println("buf: ", buf.Bytes())
	//fmt.Println("e.Data: ", e.Data)

	// datalen 4B
	datalen := uint32(len(e.Data))
	//fmt.Println("encode datalen: ", datalen)
	err = binary.Write(buf, binary.BigEndian, datalen)
	if err != nil {
		return nil, err
	}

	//fmt.Println("buf: ", buf.Bytes())
	//fmt.Println("e.Data: ", e.Data)

	if datalen > 0 {
		// Data (datalen)B
		//fmt.Println("e.Data: ", e.Data)
		err = binary.Write(buf, binary.BigEndian, e.Data)
		if err != nil {
			return nil, err
		}
	}

	//fmt.Println("buf: ", buf.Bytes())

	return buf.Bytes(), nil
}

// Decode 解码
func (e *Entry) Decode(r io.Reader) error {

	// BaseIndex
	err := binary.Read(r, binary.BigEndian, &e.BaseIndex)
	if err != nil {
		return err
	}

	// baselen
	baselen := uint16(0)
	err = binary.Read(r, binary.BigEndian, &baselen)
	if err != nil {
		return err
	}
	// Base
	if baselen > 0 {
		e.Base = make([]byte, baselen)
		err = binary.Read(r, binary.BigEndian, e.Base)
		if err != nil {
			return err
		}
	}

	// Type
	err = binary.Read(r, binary.BigEndian, &e.Type)
	if err != nil {
		return err
	}

	// datalen
	datalen := uint32(0)
	err = binary.Read(r, binary.BigEndian, &datalen)
	if err != nil {
		return err
	}
	//fmt.Println("decode datalen: ", datalen)
	if datalen > 0 {
		// Data
		e.Data = make([]byte, datalen)
		err = binary.Read(r, binary.BigEndian, e.Data)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
	序列化后的Entry格式：
	(暂时不考虑字节对齐问题)

	+--------------------------------------+
	| BaseIndex(8B) | len(Base)(2B) | Base |
	+--------------------------------------+
	| Type(1B) | len(Data)(4B) |    Data   |
	+--------------------------------------+

	长度计算公式：
	f(Entry) = 8 + 2 + len(Base) + 1 + 4 + len(Data)
*/
