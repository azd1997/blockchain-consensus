/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/11/20 11:21 AM
* @Description: bcc库内部的Net实现
***********************************************************************/

package bnet

import (
	"errors"
	"fmt"
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
// Net的定位是一个只负责数据的网络收发模块，其发送和接收都通过channel来与外界沟通
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

	// 消息读入chan
	msgin <-chan *defines.Message
	// 消息传出chan
	msgout chan<- *defines.Message

	// 关闭信号
	done chan struct{}
}

// NewNet
func NewNet(ln requires.Listener, d requires.Dialer, kv requires.Store,
	msgin <-chan *defines.Message, msgout chan<- *defines.Message) (*Net, error) {

	if ln == nil || d == nil {
		return nil, errors.New("require non-nil Listener and Dialer")
	}

	// 检查ln和d是否协议匹配
	if ln.Network() != d.Network() {
		return nil, fmt.Errorf("dismatch network: Listener(%s) and Dialer(%s)", ln.Network(), d.Network())
	}

	// 检查Store
	if kv == nil {
		return nil, errors.New("require non-nil kv store")
	}

	// 检查msgin/msgout
	if msgin == nil || msgout == nil {
		return nil, errors.New("require non-nil msgin and msgout")
	}

	// 加载pit
	pit := NewPeerInfoTable(kv)
	err := pit.Init()
	if err != nil {
		return nil, fmt.Errorf("init PeerInfoTable failed: %s", err)
	}

	n := &Net{}
	n.network = ln.Network()
	n.ln = ln
	n.d = d
	n.pit = pit
	n.conns = make(map[string]*Conn)
	n.done = make(chan struct{})

	return n, nil
}

// Init 初始化
// 根据pit的seeds和peers的情况：
// 		不管有没有peers，都向seeds节点发送getNeighbors消息，保持自身有集群最多的节点
func (n *Net) Init() error {
	// 和所有seeds建立连接，发送getNeighbors消息
	err := n.pit.RangeSeeds(func(seed *defines.PeerInfo) error {
		// 建立连接

		// 发送getNeighbors消息

		return nil
	})

	// 如果peers非空，也建立连接
	err = n.pit.RangePeers(func(peer *defines.PeerInfo) error {
		// 建立连接

		// 请求对方的进度，从而构建进度表


	})

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
func (n *Net) Stop() error {
	err := n.ln.Close()
	if err != nil {
		return err
	}
	close(n.done)
	return nil
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
	toPeerInfo, err := n.pit.Get(to)
	if err != nil {
		return nil, err
	}
	// 和to建立连接
	c, err := n.d.Dial(toPeerInfo.Addr, n.id)
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