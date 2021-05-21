package conn_net

import (
	"errors"
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/requires"
	_default "github.com/azd1997/blockchain-consensus/requires/default"
	"sync"
)

type DualNet struct {
	// 唯一标识
	id string
	// 监听地址
	addr string
	// 网络协议名
	network string
	// 监听套接字
	ln requires.Listener
	// 建连器
	d requires.Dialer

	// 连接表
	// 与对端结点连接异常时，或者后需考虑连接数量控制，会需要删除一些连接
	conns     map[string]*DualConn
	connsLock sync.RWMutex

	/*
		消息的流动顺序：
		Conn从网络中接收Message
		（Conn共享外部使用者提供的msgout）
		外部使用者处理收到的Message，作出相应
		响应的Message通过msgin传递到Net

		Conn的注意：
		Conn会在两种情况下建立：
		1. Net自身监听后接收连接
		2. Net主动连接别的节点的网络模块
	*/

	// 消息传出chan
	msgout chan *defines.Message

	// 标记Net是否准备好？
	inited bool
	// 是否关闭
	closed bool

	// 日志器
	*log.Logger

	// 关闭信号
	done chan struct{}
}

// NewDualNet 新建
func NewDualNet(id string, addr string, logger *log.Logger,
	msgchan chan *defines.Message,
	ln requires.Listener, d requires.Dialer) (*DualNet, error) {

	var (
		err error
	)

	// 检查ln和d
	if ln == nil && d == nil {
		ln, err = _default.ListenTCP(id, addr)
		if err != nil {
			return nil, err
		}
		d, err = _default.NewDialer(id, addr, 0)
		if err != nil {
			return nil, err
		}
	} else if ln != nil && d != nil {
		// 检查ln和d是否协议匹配
		if ln.Network() != d.Network() {
			return nil, fmt.Errorf("dismatch network: Listener(%s) and Dialer(%s)", ln.Network(), d.Network())
		}
		// 检查ln本机监听地址是否等于addr
		if ln.LocalListenAddr().String() != addr {
			return nil, fmt.Errorf("dismatch addr: Listener(%s) and addr(%s)", ln.LocalListenAddr(), addr)
		}
	} else { // 只有1个非空
		return nil, errors.New("require both-nil or both-non-nil Listener and Dialer")
	}

	return &DualNet{
		id:        id,
		addr:      addr,
		ln:        ln,
		d:         d,
		conns:     make(map[string]*DualConn),
		msgout:    msgchan,
		inited:    false,
		closed:    false,
		Logger:    logger,
		done:      make(chan struct{}),
	}, nil
}

func (n *DualNet) ID() string {
	return n.id
}

func (n *DualNet) Addr() string {
	return n.addr
}

// Network 网络协议
func (n *DualNet) Network() string {
	return n.network
}

// Init 初始化
// 整体的启动顺序：
// 		1. 开启监听循环
//		2. 启动消息处理循环（如果有）
// 		3. 与已有节点建立连接
//		4. 执行启动任务（如果有）
func (n *DualNet) Init() error {

	n.Infof("Init: id(%s), addr(%s): start", n.id, n.addr)

	// 启动监听循环
	go n.listenLoop()

	n.inited = true
	n.Infof("Init: id(%s), addr(%s): finish", n.id, n.addr)
	return nil
}

// Inited 是否已初始化
func (n *DualNet) Inited() bool {
	return n.inited
}

// Close 关闭
// 关闭流程：
// 		0. Close发生在外部逻辑模块关闭msgin之后，但是这边也没必要处理msgin剩余数据，直接放弃。但是要立刻关闭发送循环
//		1. 关闭监听循环，停止接入新连接
//		2. 关闭所有现有连接，停止从网络接收新数据，停止写msgout
//		3. (不关闭msgout，只是直接丢弃，因为做conns和Net间goroutine的同步没太大必要，直接放弃msgout即可)
//		*. 上面三者其实没有严格顺序，因为所有可能的存在的新连接、新消息都不会被处理，即便处理也没有关系
// 		目前直接使用done通知关闭
func (n *DualNet) Close() error {

	// 关闭监听器
	err := n.ln.Close()
	if err != nil {
		return err
	}

	// 关闭发送循环，放弃msgin的读取
	// 关闭监听循环
	// 关闭消息处理循环（如果有）
	close(n.done)

	// 关闭所有现有链接
	n.connsLock.RLock()
	for id := range n.conns {
		n.conns[id].Close()
	}
	n.connsLock.RUnlock()

	n.inited = false
	return nil
}

// Ok 判断Net是否准备好
func (n *DualNet) Ok() bool {
	return n != nil && n.ln != nil && n.d != nil
}

func (n *DualNet) Closed() bool {
	return n.closed
}

func (n *DualNet) Send(id, raddr string, msg *defines.Message) error {
	err := n.send(id, raddr, msg)
	if err != nil {
		n.Errorf("ConnNet send msg(%s) fail. raddr=%s, err=%s, msg=%v", msg.Type.String(), raddr, err, msg)
		return err
	}
	n.Debugf("ConnNet send msg(%s) succ. raddr=%s, n=%d, msg=%v", msg.Type.String(), raddr, n, msg)
	return nil
}

func (n *DualNet) SetMsgOutChan(bus chan *defines.Message) {
	if n.msgout == nil {
		n.msgout = bus
	}
}

func (n *DualNet) RecvLoop() {
	// 空实现
}

func (n *DualNet) DisplayAllConns(brief bool) string {
	str := ""
	if brief { // 打印名字
		str += fmt.Sprintf("\nAll Conns of [%s]: (brief version)\n", n.id)
		n.connsLock.RLock()
		for _, v := range n.conns {
			str += fmt.Sprintf("%s\n", v.Name())
		}
		n.connsLock.RUnlock()
		str += "\n"
	} else {
		str += fmt.Sprintf("\nAll Conns of [%s]: \n", n.id)
		n.connsLock.RLock()
		for _, v := range n.conns {
			str += fmt.Sprintf("%s\n", v.String())
		}
		n.connsLock.RUnlock()
		str += "\n"
	}
	return str
}


// connect 返回 由本机向某个id的节点发起的连接
// 如果连接已存在，则返回该连接Conn
// 如果链接不存在，则创建
func (n *DualNet) connect(to, raddr string) (*DualConn, error) {
	if !n.Ok() {
		return nil, errors.New("Net is not ok for Connect")
	}

	// to是自己
	if n.id == to {
		return nil, nil
	}

	// 连接存在，直接返回
	var dc *DualConn
	var exists bool

	// 注意：检测->更改
	// 这个逻辑应该锁在一起，不然当主动创建连接和被动接收连接同时时，会出现问题，“吞掉”其中一个

	n.connsLock.RLock()
	dc, exists = n.conns[to]
	n.connsLock.RUnlock()

	// 如果连接存在，并且其中的send_conn可用，那么直接返回该连接
	if exists &&
		(dc.status == ConnStatus_SendRecv || dc.status == ConnStatus_OnlySend) {
		return dc, nil
	}

	// 否则的话，新键send_conn
	// 连接不存在(或原来的连接已经停止了)，创建连接 和to建立连接
	sendConn, err := n.d.Dial(raddr, to)
	if err != nil {
		return nil, err
	}
	n.connsLock.Lock()
	if !exists || dc==nil {		// 这是旧的检测结果，此时应该重新检测一次
		dc, exists = n.conns[to]
		if !exists || dc == nil {
			dc = ToDualConn(sendConn, nil, n.msgout) // c传输过来的消息会写到msgout传出去
			n.conns[to] = dc
		} else {
			dc.sendConn = sendConn
		}
	} else {	// 旧的检测结果已经检测出存在了
		dc.sendConn = sendConn
	}
	n.connsLock.Unlock()

	return dc, nil
}

// startConn 启动连接
func (n *DualNet) startConn(c *DualConn) {
	if c == nil {
		return
	}
	// 启动其接收循环
	go c.RecvLoop()
}

// send 向对端节点发送消息
// 如果net.id与to之间已经有Conn，那么通过该Conn发送消息
// 如果没有Conn，那么建立新Conn，并发送消息。
// TODO: 关于连接数量的控制
func (n *DualNet) send(id, raddr string, msg *defines.Message) error {
	// 获取或创建连接
	conn, err := n.connect(id, raddr)	// 这里要求的connect是send_conn
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

// listenLoop 监听循环
func (n *DualNet) listenLoop() {

	for {
		select {
		case <-n.done:
			n.Infof("listenLoop: returned...")
			return
		default:
			// 接受连接
			recvConn, err := n.ln.Accept()
			if err != nil {
				n.Errorf("listenLoop: accept fail: %s", err)
				continue
			}

			// 两种情况
			var dc *DualConn
			var exists bool

			n.connsLock.Lock()	// 枷锁
			dc, exists = n.conns[recvConn.RemoteID()]
			// 1. 原先不存在该DualConn
			if !exists || dc==nil {
				// 启动连接，循环接收消息
				c := ToDualConn(nil, recvConn, n.msgout)
				// 记录连接
				n.connsLock.Lock()
				n.conns[recvConn.RemoteID()] = c
				n.connsLock.Unlock()
				// 启动连接
				n.startConn(c)
			} else {// 该连接原先已创建 dc!=nil
				// 不管原先的recv是否存在都替换
				// 原因是：recv是对端发起的连接，对端先关闭之后立马又建立，
				// 这种情况下本机节点可能还没有将recv清除
				dc.recvConn = recvConn	// 将新建立的连接conn放入dualconn中，并启动
				if dc.status == ConnStatus_Closed || dc.status == ConnStatus_OnlySend {
					n.startConn(dc)
				}	// 避免重复 go dc.RecvLoop()
			}
			n.connsLock.Unlock()	// 解锁
		}
	}
}