package btcp_dual

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/modules/bnet/conn_net"
)

func New(id string, addr string, logger *log.Logger,
	msgchan chan *defines.Message) (*conn_net.DualNet, error) {

	return conn_net.NewDualNet(id, addr, logger, msgchan, nil, nil)
}