package cs_net

import (
	"time"

	"github.com/azd1997/blockchain-consensus/requires"
	_default "github.com/azd1997/blockchain-consensus/requires/default"
)

//var (
//	listener requires.Listener
//	dialer requires.Dialer
//)
//
//func init() {
//	listener ,err := _default.ListenTCP("id", "listenAddr")
//
//	dialer, err := _default.NewDialer("id", "addr", 5*time.Second)
//}

func listener(serverId, listenAddr string) requires.Listener {
	ln ,err := _default.ListenTCP(serverId, listenAddr)
	if err != nil {
		panic(err)
	}
	return ln
}

// 尽管传入了dialer端的listenAddr，但实际没用
func dialer(dialerId, listenAddr string) requires.Dialer {
	d, err := _default.NewDialer(dialerId, listenAddr, 5*time.Second)
	if err != nil {
		panic(err)
	}
	return d
}