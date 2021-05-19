package btcp

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/modules/bnet/conn_net"
)

func New(id string, addr string, logger *log.Logger,
	msgchan chan *defines.Message) (*conn_net.Net, error) {

	return conn_net.New(id, addr, logger, msgchan, nil, nil)
}