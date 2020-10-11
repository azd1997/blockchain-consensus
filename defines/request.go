/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/22 10:59
* @Description: 请求(Request)
***********************************************************************/

package defines

import (
	"encoding/binary"
	"io"

	"github.com/azd1997/blockchain-consensus/utils/bufferpool"
)

type RequestType uint8

const (
	RequestType_Blocks RequestType = 0
	RequestType_Neighbors RequestType = 1
)

// Request 请求
type Request struct {
	Type RequestType

	// 根据index区间请求
	IndexStart uint64
	IndexCount uint64

	// 根据哈希请求
	Hashes [][]byte
}

// Check 检查格式
func (req *Request) Check() error {
	return nil
}

// Len 获取序列化后长度
func (req *Request) Len() int {
	length := 1 + 8 + 8 + 4
	hashnum := len(req.Hashes)
	if hashnum == 0 {
		return length
	} else {
		hashlen := len(req.Hashes[0])
		return length + 4 + hashnum * hashlen
	}
}

// Encode 编码
func (req *Request) Encode() ([]byte, error) {
	var err error

	// 检查格式
	if err = req.Check(); err != nil {
		return nil, err
	}

	// 获取缓冲
	buf := bufferpool.Get()
	defer bufferpool.Return(buf)

	/* 序列化 */

	// Type 1B
	err = binary.Write(buf, binary.BigEndian, req.Type)
	if err != nil {
		return nil, err
	}

	// IndexStart 8B
	err = binary.Write(buf, binary.BigEndian, req.IndexStart)
	if err != nil {
		return nil, err
	}

	// IndexCount 8B
	err = binary.Write(buf, binary.BigEndian, req.IndexCount)
	if err != nil {
		return nil, err
	}

	// hashnum
	hashnum := uint32(len(req.Hashes))
	err = binary.Write(buf, binary.BigEndian, hashnum)
	if err != nil {
		return nil, err
	}
	if hashnum == 0 {
		return buf.Bytes(), nil
	}
	// hashlen
	hashlen := uint32(len(req.Hashes[0]))
	err = binary.Write(buf, binary.BigEndian, hashlen)
	if err != nil {
		return nil, err
	}
	// Hashes
	for i:=uint32(0); i<hashnum; i++ {
		err = binary.Write(buf, binary.BigEndian, req.Hashes[i])
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// Decode 解码
func (req *Request) Decode(r io.Reader) error {
	// Type 1B
	err := binary.Read(r, binary.BigEndian, &req.Type)
	if err != nil {
		return err
	}

	// IndexStart 8B
	err = binary.Read(r, binary.BigEndian, &req.IndexStart)
	if err != nil {
		return err
	}

	// IndexCount 8B
	err = binary.Read(r, binary.BigEndian, &req.IndexCount)
	if err != nil {
		return err
	}

	// hashnum
	hashnum := uint32(0)
	err = binary.Read(r, binary.BigEndian, &hashnum)
	if err != nil {
		return err
	}
	if hashnum == 0 {
		return nil
	}
	// hashlen
	hashlen := uint32(0)
	err = binary.Read(r, binary.BigEndian, &hashlen)
	if err != nil {
		return err
	}
	// Hashes
	req.Hashes = make([][]byte, hashnum)
	for i:=uint32(0); i<hashnum; i++ {
		req.Hashes[i] = make([]byte, hashlen)
		err = binary.Read(r, binary.BigEndian, req.Hashes[i])
		if err != nil {
			return err
		}
	}

	return nil
}

/*
	序列化后的Request格式：
	(暂时不考虑字节对齐问题)

	+--------------------------------------+
	| Type(1B) |  Start(8B)  |   Count(8B)   |
	+--------------------------------------+
	|   hashnum(4B)   |    hashlen(4B)     |
	+--------------------------------------+
	|              Hashi(hashlenB)         |	// Hashes
	+--------------------------------------+

	长度计算公式：
	f(Request) = 1 + 8 + 8 + 4 + 4 + hashlen * hashnum
*/