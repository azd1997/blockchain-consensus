package cs_net

import (
	"errors"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/bnet/conn_net"
	"github.com/azd1997/blockchain-consensus/requires"
)

// 使用逻辑：
// server启动 ->  ----------------------------------------------- -> recvLoop
// 				 新建client -> client Dial建立连接 -> client Send

// Client 针对Net模块作为Server的客户端
// 为了避免增加一份新的从conn到net的代码，这里实现C/S架构时仍保留id字段，以保持兼容性
// 暂时Client绑定一个确定的Server，不支持更换，并且支持默认版本的TCP（_default包）
type Client struct {
	d requires.Dialer	// 发起连接
	conn *conn_net.Conn	// 维护连接

	id string	// 本机节点id
	addr string	// 无用
	serverId string // 无用
	serverAddr string		// 服务器地址
}

func Dial(serverId, serverListenAddr, clientId, clientListenAddr string, msgout chan *defines.Message) (*Client, error) {
	d := dialer(clientId, clientListenAddr)
	rawC, err := d.Dial(serverListenAddr, serverId)
	if err != nil {
		return nil, err
	}
	conn := conn_net.ToConn(rawC, msgout)

	// 启动接收循环
	go conn.RecvLoop()

	return &Client{
		d:          d,
		conn:       conn,
		id:         clientId,
		addr:       clientListenAddr,
		serverId:   serverId,
		serverAddr: serverListenAddr,
	}, nil
}

func (c *Client) Send(msg *defines.Message) error {
	if c.conn == nil {
		return errors.New("c,conn == nil")
	}

	return c.conn.Send(msg)
}
