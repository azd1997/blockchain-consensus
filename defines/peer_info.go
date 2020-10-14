/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 20:39
* @Description: 节点信息
***********************************************************************/

package defines

import (
	"encoding/gob"
	"encoding/json"
	"io"

	"github.com/azd1997/blockchain-consensus/utils/bufferpool"
)

//// Peer 节点信息
//type Peer struct {
//	ID string
//	Addr string
//	Data []byte
//}

// PeerInfo 节点信息
type PeerInfo struct {
	Id string
	Addr string
	Data []byte
}

// String
func (pi *PeerInfo) String() string {
	b, err := json.Marshal(pi)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// Encode 编码
func (pi *PeerInfo) Encode() ([]byte, error) {
	buf := bufferpool.Get()
	defer bufferpool.Return(buf)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(pi)
	return buf.Bytes(), err
}

// Decode 解码
// pi := new(PeerInfo)
func (pi *PeerInfo) Decode(r io.Reader) error {
	dec := gob.NewDecoder(r)
	return dec.Decode(pi)
}
