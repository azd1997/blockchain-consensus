/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/11/20 11:21 AM
* @Description: bcc库内部的Net实现
***********************************************************************/

package bnet

import (
	"errors"
	"log"
	"net"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
)
/*
	bcc内采用“连接”式的传输通道，默认是tcp的实现方式。
	因此需要类比于 监听套接字 和 连接套接字 的两个接口和封装类
	具体底层使用其他通信协议时，需要按该接口进行封装使用


	Listener 监听器(监听套接字)
	Conn 连接(连接套接字)


*/



// Net 网络模块
type Net struct {
	// 唯一标识
	id string
	// 网络协议名
	network string
	// 监听套接字
	ln requires.Listener
	// 建连器
	d requires.Dialer
	// 连接表
	conns map[string]*Conn
	// 节点信息表
	pit *PeerInfoTable

	// 关闭信号
	done chan struct{}
}

// NewNet
func NewNet(ln requires.Listener, d requires.Dialer) *Net {
	if ln == nil || d == nil {
		log.Printf("require non-nil Listener and Dialer\n")
		return nil
	}

	// 检查ln和d是否协议匹配
	if ln.Network() != d.Network() {
		log.Printf("dismatch network: Listener(%s) and Dialer(%s)\n", ln.Network(), d.Network())
		return nil
	}

	// 加载pit
	pit := &PeerInfoTable{}
	err := pit.Load()
	if err != nil {
		log.Printf("load PeerInfoTable failed: %s\n", err)
		return nil
	}

	n := &Net{}
	n.network = ln.Network()
	n.ln = ln
	n.d = d
	n.pit = pit
	n.conns = make(map[string]*Conn)
	n.done = make(chan struct{})

	return n
}

// Ok 判断Net是否准备好
func (n *Net) Ok() bool {
	return n != nil && n.ln != nil && n.d != nil
}

// Network 网络协议
func (n *Net) Network() string {
	return n.network
}

// Stop 停止Net模块
func (n *Net) Stop() {
	n.ln.Close()
	close(n.done)
}

// ListenLoop 监听循环
func (n *Net) ListenLoop() {
	for {
		// 接受连接
		conn, err := n.ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		// 启动连接，循环接收消息
		c := ToConn(conn)
		// 记录连接
		n.conns[conn.RemoteID()] = c
		// 启动连接
		n.startConn(c)
	}
}

// Send 向对端节点发送消息
// 如果net.id与to之间已经有Conn，那么通过该Conn发送消息
// 如果没有Conn，那么建立新Conn，并发送消息。
// TODO: 关于连接数量的控制
func (n *Net) Send(to string, msg *defines.Message) error {
	// 获取或创建连接
	conn, err := n.Connect(to)
	if err != nil {
		return err
	}

	// 发送消息
	err = conn.Send(msg)
	if err != nil {
		return err
	}

	return nil
}

// Addr 获取本机监听的地址
func (n *Net) Addr() net.Addr {
	return n.ln.LocalListenAddr()
}

// Connect 连接某个id的节点
// 如果连接已存在，则返回该连接Conn
// 如果链接不存在，则创建
func (n *Net) Connect(to string) (*Conn, error) {
	if !n.Ok() {
		return nil, errors.New("Net is not ok for Connect")
	}

	// 连接存在，直接返回
	if n.conns[to] != nil {
		return n.conns[to], nil
	}

	// 连接不存在，创建连接
	toAddr := n.pit.Get(to)
	// 和to建立连接
	c, err := n.d.Dial(toAddr, n.id)
	if err != nil {
		return nil, err
	}
	conn := ToConn(c)
	// 记录连接
	n.conns[to] = conn
	// 启动连接
	n.startConn(conn)

	return conn, nil
}

// startConn 启动连接
func (n *Net) startConn(c *Conn) {
	if c == nil {
		return
	}
	// 启动其接收循环
	go c.RecvLoop()
}