/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 5/21/21 8:48 AM
* @Description: The file is for
***********************************************************************/

package conn_net

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
	_default "github.com/azd1997/blockchain-consensus/requires/default"
)

var testMsg = &defines.Message{
	Version: defines.CodeVersion,
	Type:    defines.MessageType_Blocks,
	From:    "id_peerb",
	To:      "id_peera",
	Sig:     []byte("Signature"),
}

type peer struct {
	id   string
	addr string
	ln   requires.Listener
	d    requires.Dialer
}

func genPeer(id, addr string) (*peer, error) {
	ln, err := _default.ListenTCP(id, addr)
	if err != nil {
		return nil, err
	}
	d, err := _default.NewDialer(id, addr, 0)
	if err != nil {
		return nil, err
	}
	return &peer{
		id:   id,
		addr: addr,
		ln:   ln,
		d:    d,
	}, nil
}

