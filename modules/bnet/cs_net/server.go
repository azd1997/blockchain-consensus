package cs_net

import (
	"fmt"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/modules/bnet/conn_net"
)

type Server struct {
	*conn_net.Net	// 继承Net的所有方法，对一些不想暴露的方法应该覆盖为空
}

// NewServer 新建Server	//，启动监听循环
func NewServer(id, addr string, msgout chan *defines.Message, logger *log.Logger) (*Server, error) {
	cn, err := conn_net.NewNet(id, addr, logger, msgout, nil, nil)	// ln和d都使用Net默认的，等同于使用listen()和dialer()
	if err != nil {
		return nil, fmt.Errorf("NewServer fail: %s", err)
	}
	server := &Server{Net:cn}
	return server, nil
}


