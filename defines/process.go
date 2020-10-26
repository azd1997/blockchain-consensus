/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/19/20 5:34 PM
* @Description: 区块链进度
***********************************************************************/

package defines

import (
	"bytes"
	"encoding/gob"

	"github.com/azd1997/blockchain-consensus/utils/bufferpool"
)

// Process 进度
type Process struct {
	Index uint64
	Hash []byte
	LatestMaker string
	Id string		// 哪个节点的Process
}

func (p *Process) Encode() ([]byte, error) {
	buf := bufferpool.Get()
	defer bufferpool.Return(buf)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(p)
	return buf.Bytes(), err
}

// Decode 解码
// p := new(Process)
func (p *Process) Decode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))
	return dec.Decode(p)
}