/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/11/20 11:21 AM
* @Description: bcc库内部的Net实现
***********************************************************************/

package conn_net

import (
	"errors"
	"fmt"
	"sync"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/requires"
	_default "github.com/azd1997/blockchain-consensus/requires/default"
)

/*
	bcc内采用“连接”式的传输通道，默认是tcp的实现方式。
	因此需要类比于 监听套接字 和 连接套接字 的两个接口和封装类
	具体底层使用其他通信协议时，需要按该接口进行封装使用


	Listener 监听器(监听套接字)
	Conn 连接(连接套接字)


*/

/*
	Net(网络模块)							SM(逻辑模块)
->	msgout	(网络接收数据)	--->			msgin    而后逻辑处理后写到SM.msgout
<-	msgin (向网络发送)		<--				msgout

	注意: Net.msgout是从SM.msgin引用过来； Net.msgin是从SM.msgout引用过来
*/

// Net 网络模块
// Net的定位是一个只负责数据的网络收发模块，其发送和接收都通过channel来与外界沟通
// Net只负责消息收发和连接管理，其他一概不负责
// 假设还有个逻辑处理模块 SM，SM与Net的启动顺序应该是: Net SM
// 关闭顺序：SM Net
// （这里要注意：两个模块关闭时都是丢弃所有channel中数据的，原因是没必要去处理善后，每次节点重启都会同步自身到最新
// SM关闭时需要关闭msgin，并停止处理msgout内数据， 接下来调用Net.Close()，而后Net放弃msgin所有数据丢弃，并关闭msgout）
type Net struct {
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
	conns     map[string]*Conn
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

	//消息读入chan
	//msgin chan *defines.MessageWithError
	// 消息传出chan
	msgout chan *defines.Message

	// 节点信息表
	//pit *memorypit.PeerInfoTable

	/*自定义的Net启动执行的任务*/
//	customInitFunc func(n *Net) error
	/*自定义的消息处理函数*/
//	customMsgHandleFunc func(n *Net, msg *defines.Message) error

	// 标记Net是否准备好？
	inited bool
	// 是否关闭
	closed bool

	// 日志器
	*log.Logger

	// 关闭信号
	done chan struct{}
}

// NewNewNet 新建
func NewNet(id string, addr string, logger *log.Logger,
		msgchan chan *defines.Message,
		ln requires.Listener, d requires.Dialer) (*Net, error) {

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

	return &Net{
		id:        id,
		addr:      addr,
		ln:        ln,
		d:         d,
		conns:     make(map[string]*Conn),
		msgout:    msgchan,
		inited:    false,
		closed:    false,
		Logger:    logger,
		done:      make(chan struct{}),
	}, nil
}

func (n *Net) ID() string {
	return n.id
}

func (n *Net) Addr() string {
	return n.addr
}

// Network 网络协议
func (n *Net) Network() string {
	return n.network
}

// Init 初始化
// 整体的启动顺序：
// 		1. 开启监听循环
//		2. 启动消息处理循环（如果有）
// 		3. 与已有节点建立连接
//		4. 执行启动任务（如果有）
func (n *Net) Init() error {

	n.Infof("Init: id(%s), addr(%s): start", n.id, n.addr)

	// 启动监听循环
	go n.listenLoop()
	//// 启动消息处理循环
	//if n.customMsgHandleFunc != nil {
	//	go n.msgHandleLoop()
	//}
	//// 启动发送循环
	//go n.msgSendLoop()

	//f := func(peer *defines.PeerInfo) error {
	//	// 建立连接
	//	_, err := n.connect(peer.Id)
	//	if err != nil {
	//		n.Errorf("Init: connect to peer (%s,%s) fail: %s", peer.Id, peer.Addr, err)
	//		return err
	//	}
	//	return nil
	//}
	//
	//// 和所有seeds建立连接，发送getNeighbors消息
	//n.pit.RangeSeeds(f)
	//// 如果peers非空，也建立连接
	//n.pit.RangePeers(f)
	//
	//// 如果已经设置CustomInitFunc，那么执行
	//if n.customInitFunc != nil {
	//	n.Infof("Init: CustomInitFunc: ready")
	//	err := n.customInitFunc(n)
	//	if err != nil {
	//		// 这里如果出错不会退出，而是打日志
	//		n.Errorf("Init: CustomInitFunc: %s", err)
	//	}
	//	n.Infof("Init: CustomInitFunc: finish")
	//}

	n.inited = true
	n.Infof("Init: id(%s), addr(%s): finish", n.id, n.addr)
	return nil
}

// Inited 是否已初始化
func (n *Net) Inited() bool {
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
func (n *Net) Close() error {

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
func (n *Net) Ok() bool {
	return n != nil && n.ln != nil && n.d != nil
}

func (n *Net) Closed() bool {
	return n.closed
}

func (n *Net) Send(id, raddr string, msg *defines.Message) error {
	err := n.send(id, raddr, msg)
	if err != nil {
		n.Errorf("ConnNet send msg(%s) fail. raddr=%s, err=%s, msg=%v", msg.Type.String(), raddr, err, msg)
		return err
	}
	n.Debugf("ConnNet send msg(%s) succ. raddr=%s, n=%d, msg=%v", msg.Type.String(), raddr, n, msg)
	return nil
}

func (n *Net) SetMsgOutChan(bus chan *defines.Message) {
	if n.msgout == nil {
		n.msgout = bus
	}
}

func (n *Net) RecvLoop() {
	// 空实现
}


func (n *Net) DisplayAllConns(brief bool) string {
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


// NewNet
//func NewNet(opt *Option) (*Net, error) {
//
//	n := &Net{}
//
//	// 日志器
//	n.Logger = log.NewLogger(Module_Net, opt.Id)
//	if n.Logger == nil {
//		return nil, errors.New("nil logger, please init logger first")
//	}
//
//	// 节点表
//	if opt.Pit == nil {
//		n.pit = pitable.Global()
//	} else {
//		if opt.Pit.Inited() {
//			n.pit = opt.Pit
//		} else {
//			return nil, errors.New("PeerInfoTable should be inited")
//		}
//	}
//
//	// 检查opt.Listener和opt.Dialer
//	if opt.Listener == nil && opt.Dialer == nil {
//		ln, err := _default.ListenTCP(opt.Id, opt.Addr)
//		if err != nil {
//			return nil, err
//		}
//		d, err := _default.NewDialer(opt.Id, opt.Addr, 0)
//		if err != nil {
//			return nil, err
//		}
//		n.ln, n.d = ln, d
//		n.id, n.addr, n.network = opt.Id, opt.Addr, ln.Network()
//	} else if opt.Listener != nil && opt.Dialer != nil {
//		ln, d := opt.Listener, opt.Dialer
//		// 检查ln和d是否协议匹配
//		if ln.Network() != d.Network() {
//			return nil, fmt.Errorf("dismatch network: Listener(%s) and Dialer(%s)", ln.Network(), d.Network())
//		}
//		n.ln, n.d = ln, d
//		n.id, n.addr, n.network = ln.LocalID(), ln.LocalListenAddr().String(), ln.Network()
//	} else { // 只有1个非空
//		return nil, errors.New("require both-nil or both-non-nil Listener and Dialer")
//	}
//
//	// 检查msgin/msgout
//	if opt.MsgIn == nil || opt.MsgOut == nil {
//		return nil, errors.New("require non-nil msgin and msgout")
//	}
//	n.msgin, n.msgout = opt.MsgIn, opt.MsgOut
//
//	// 这两个函数是可空的
//	n.customInitFunc = opt.CustomInitFunc
//	n.customMsgHandleFunc = opt.CustomMsgHandleFunc
//
//	n.conns = make(map[string]*Conn)
//	n.connsLock = new(sync.RWMutex)
//	n.done = make(chan struct{})
//
//	return n, nil
//}





//// Stop 停止Net模块
//func (n *Net) Stop() error {
//	err := n.ln.Close()
//	if err != nil {
//		return err
//	}
//	close(n.done)
//	return nil
//}

// Addr 获取本机监听的地址
//func (tn *Net) Addr() net.Addr {
//	return tn.ln.LocalListenAddr()
//}

// connect 连接某个id的节点
// 如果连接已存在，则返回该连接Conn
// 如果链接不存在，则创建
func (n *Net) connect(to, raddr string) (*Conn, error) {
	if !n.Ok() {
		return nil, errors.New("Net is not ok for Connect")
	}

	// to是自己
	if n.id == to {
		return nil, nil
	}

	// 连接存在，直接返回
	n.connsLock.RLock()
	c1, exists := n.conns[to]
	n.connsLock.RUnlock()
	if exists && c1.status == ConnStatus_SendRecv {
		return c1, nil
	}


	// 连接不存在(或原来的连接已经停止了)，创建连接 和to建立连接
	c, err := n.d.Dial(raddr, to)
	if err != nil {
		return nil, err
	}
	conn := ToConn(c, n.msgout) // c传输过来的消息会写到msgout传出去
	// 记录连接
	n.connsLock.Lock()
	n.conns[to] = conn
	n.connsLock.Unlock()
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

// send 向对端节点发送消息
// 如果net.id与to之间已经有Conn，那么通过该Conn发送消息
// 如果没有Conn，那么建立新Conn，并发送消息。
// TODO: 关于连接数量的控制
func (n *Net) send(id, raddr string, msg *defines.Message) error {
	// 获取或创建连接
	conn, err := n.connect(id, raddr)
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

////////////////////////////// 工作循环 ///////////////////////////////
// 所有的工作循环统一通过done退出

// 消息处理循环
// 仅当Net.msgHandleFunc设置时生效
//func (tn *Net) msgHandleLoop() {
//	if tn.customMsgHandleFunc == nil {
//		tn.Error("msgHandleLoop: nil msgHandleFunc, return")
//		return
//	}
//
//	var err error
//
//	for {
//		select {
//		case <-tn.done:
//			tn.Infof("msgHandleLoop: returned...")
//			return
//		case msg := <-tn.msgout:
//			err = tn.customMsgHandleFunc(tn, msg)
//			if err != nil {
//				tn.Errorf("msgHandleLoop: handle msg: {msg:%v, err:%s}", msg, err)
//			}
//		}
//	}
//}

// msgSendLoop 发送循环
// 不断读msgin的消息，然后发送出去
//func (tn *Net) msgSendLoop() {
//	for {
//		select {
//		case <-tn.done:
//			tn.Infof("msgSendLoop: returned...")
//			return
//		case msg := <-tn.msgin:
//			// 检查
//			if err := msg.Check(); err != nil {
//				tn.Errorf("msgSendLoop: recv invalid msg(%s): msg=%v, err=%s", msg.Msg.Desc, msg.Msg, err)
//				continue
//			} else {
//				//n.Infof("msgSendLoop: recv msg(%v) from local\n", msg)
//			}
//			// 发送
//			if err := tn.send(msg.Msg.To, msg.Msg); err != nil {
//				tn.Errorf("msgSendLoop: send msg(%s) fail: msg=%v, err=%s", msg.Msg.Desc, msg.Msg, err)
//				msg.Err <- err // 回传发送时错误信息
//			} else {
//				tn.Debugf("msgSendLoop: send msg(%s) succ: msg=%v", msg.Msg.Desc, msg.Msg)
//				msg.Err <- nil
//			}
//		}
//	}
//}

// listenLoop 监听循环
func (n *Net) listenLoop() {

	for {
		select {
		case <-n.done:
			n.Infof("listenLoop: returned...")
			return
		default:
			// 接受连接
			conn, err := n.ln.Accept()
			if err != nil {
				n.Errorf("listenLoop: accept fail: %s", err)
				continue
			}
			// 启动连接，循环接收消息
			c := ToConn(conn, n.msgout)
			// 记录连接
			n.connsLock.Lock()
			n.conns[conn.RemoteID()] = c
			n.connsLock.Unlock()
			// 启动连接
			n.startConn(c)
		}
	}
}

//func (n *Net) Send(req *defines.Request, rsp *Response) {
//
//}
