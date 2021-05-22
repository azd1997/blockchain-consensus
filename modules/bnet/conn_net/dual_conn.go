package conn_net

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/utils/binary"
)

// DualConn 双链接
// 注意：在任何对sendConn/recvConn新建\销毁等操作上必须将其与状态ConnStatus绑定，此外recvConn还需要和RecvLoop绑定
// 不要将这些关系留给使用者去做！
type DualConn struct {
	sendConn, recvConn requires.Conn	// 两条实际连接，一条负责发送，一条负责接收
	sendLock, recvLock sync.RWMutex		// 由于这两个连接中途有可能更换或者销毁之类的，所以要上锁，避免出现更换与使用同时出现的情形
	// 这两把锁保护的只是sendConn,recvConn这两个指针，对于指针指向的内容不关心
	// 并且status的变更也要放到临界区以保证和这两个连接的变更同步

	// msgChan为Net也就是多个Conn的管理结构所提供，在多个Conn间共享该msgChan
	// Conn只负责写msgChan，Net会另起goroutine循环读msgChan
	msgChan chan<- *defines.Message // Message channel 从recv读到该chan

	status  ConnStatus    // Conn当前状态
	timeout time.Duration // 超时
}

// ToDualConn 将requires.Conn封装成DualConn
func ToDualConn(send, recv requires.Conn, recvmsg chan<- *defines.Message) *DualConn {
	dc := &DualConn{
		msgChan:recvmsg,
		timeout: DefaultConnTimeout,
		status:ConnStatus_Closed,
	}

	if send != nil {
		dc.sendLock.Lock()
		dc.sendConn = send
		dc.status.EnableSend()
		dc.sendLock.Unlock()
	}
	if recv != nil {
		dc.recvLock.Lock()
		dc.recvConn = recv
		dc.status.EnableRecv()
		go dc.RecvLoop()	// 启动接收循环
		dc.recvLock.Unlock()		// 将go RecvLoop这句包含进去
	}
	return dc
}

// SetSendConn 向已存在的DualConn更新sendConn
func (c *DualConn) SetSendConn(sendConn requires.Conn) {
	c.sendLock.Lock()
	c.sendConn = sendConn
	c.status.EnableSend()
	c.sendLock.Unlock()
}

// SetRecvConn 向已存在的DualConn更新recvConn
func (c *DualConn) SetRecvConn(recvConn requires.Conn) {
	//c.recvLock.Lock()
	// 分三种情况：recv之前没有；之前有但没有启动RecvLoop；之前有并且已启动RecvLoop
	// 通过ToDualConn中将recvConn与RecvLoop绑定 并且加锁 ，可以保证情况2不会出现
	// 因此，对于情况1，需要更新recvConn并启动RecvLoop，对于情况3，直接更新
	//c.recvConn = recvConn
	//if !c.status.CanRecv() {
	//  c.status.EnableRecv()
	//	go c.RecvLoop()
	//}
	//c.recvLock.Unlock()
	// 其实这里只要把原有的conn关闭，那么原本的Recv必将退出，也可以实现想要的功能

	// NOTICE: recvConn是由对端主动请求之后才会建立，所以新的recvConn只会在旧的recvConn关闭之后建立
	// 因此，这里采取将recvConn生命周期绑定RecvLoop的做法
	// 不管原先是否有RecvLoop，直接重新go RecvLoop()
	// 旧的recvConn关闭之后，其绑定的RecvLoop也将关闭
	c.recvLock.Lock()
	c.recvConn = recvConn
	c.status.EnableRecv()
	go c.RecvLoop()
	c.recvLock.Unlock()
}

// NOTICE: 由于Status()/String()/Name()这样的用于打印的方法并不重要，所以不加读锁，直接用
// 对于 只读 的方法， 如果不是很重要、会影响关键逻辑，不加读锁
// 对于 写 的方法，必须加写锁

// Status 报告Conn状态
func (c *DualConn) Status() string {
	return c.status.String()
}

// Name Conn名称
func (c *DualConn) Name() string {
	if c.sendConn != nil && c.recvConn != nil {
		return fmt.Sprintf("[%s]<->[%s](s:%s,r:%s)",
			c.sendConn.LocalID(), c.sendConn.RemoteID(),
			c.sendConn.HashCode(),  c.recvConn.HashCode())
	} else if c.recvConn == nil {
		return fmt.Sprintf("[%s]->[%s](s:%s,r:%s)",
			c.sendConn.LocalID(), c.sendConn.RemoteID(),
			c.sendConn.HashCode(), "_")
	} else if c.sendConn == nil {
		return fmt.Sprintf("[%s]<-[%s](s:%s,r:%s)",
			c.recvConn.LocalID(), c.recvConn.RemoteID(),
			"_", c.recvConn.HashCode())
	} else {
		return "_"
	}
}

// String 描述
func (c *DualConn) String() string {

	var availableConn requires.Conn
	if c.sendConn != nil {
		availableConn = c.sendConn
	} else if c.recvConn != nil {
		availableConn = c.recvConn
	} else {
		return ""
	}

	//localId, localListenAddr, localAddr, remoteId, remoteListenAddr, remoteAddr := "", "", "", "", "", ""
	//localId = availableConn.LocalID()
	//localAddr = availableConn.LocalAddr().String()
	//localListenAddr = availableConn.LocalListenAddr().String()
	//remoteId = availableConn.RemoteID()
	//remoteListenAddr = availableConn.RemoteListenAddr().String()
	//remoteAddr = availableConn.RemoteAddr().String()

	localAddr, remoteAddr := "", ""
	if c.sendConn != nil {
		localAddr = c.sendConn.LocalAddr().String()
	}
	if c.recvConn != nil {
		remoteAddr = c.recvConn.RemoteAddr().String()
	}

	return fmt.Sprintf("DualConn %s info: {Network: %s, Local: %s(%s,%s), Remote: %s(%s,%s), Status: %s, Timeout: %s}",
		c.Name(),
		availableConn.Network(),
		availableConn.LocalID(), availableConn.LocalListenAddr().String(), localAddr,
		availableConn.RemoteID(), availableConn.RemoteListenAddr().String(), remoteAddr,
		c.status.String(), c.timeout.String())
}

// Close 关闭连接
func (c *DualConn) Close() error {
	c.sendLock.Lock()
	if c.sendConn != nil {c.sendConn.Close()}
	c.status.DisableSend()
	c.sendLock.Unlock()

	c.recvLock.Lock()
	if c.recvConn != nil {c.recvConn.Close()}
	c.status.DisableRecv()
	c.recvLock.Unlock()

	//c.status = ConnStatus_Closed
	return nil
}

// Send 发送消息
func (c *DualConn) Send(msg *defines.Message) error {
	if c.sendConn == nil {
		log.Printf("DualConn(%s) Send quit: send_conn == nil \n", c.Name())
		return nil
	}

	// 序列化为字节数组
	b, err := msg.Encode()
	if err != nil {
		return err
	}

	// 给实际内容前面封装magic和消息长度
	length := uint32(len(b))
	buf2 := new(bytes.Buffer)
	err = binary.Write(buf2, binary.BigEndian, defines.MessageMagicNumber, length, b)
	if err != nil {
		return err
	}

	// 发送
	_, err = c.sendConn.Write(buf2.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// RecvLoop 接收循环，在goroutine内循环读取并解码成Message
// 这里涉及到一个问题：单个消息的大小需要有一定限制，不然以特殊边界符来控制读取停止的话，
// 可能不太安全。不容易预设一开始的缓冲区大小
// 因此：
// 这里采取限制消息最大尺寸的做法，这样容易实现，预设合理的缓冲区大小也有利于执行性能
//
// 关于RecvLoop退出：
// c.Close()后主动关闭，而后c.RecvLoop()会因为“use of closed network connection”而退出
// 对端的conn则会收到“EOF”而退出
// TODO: 链接关闭时退出
func (c *DualConn) RecvLoop() {
	if c.recvConn == nil {
		log.Printf("DualConn(%s) RecvLoop quit: recv_conn == nil \n", c.Name())
		return
	}

	//if c.status == ConnStatus_OnlySend {
	//	c.status = ConnStatus_SendRecv
	//} else if c.status == ConnStatus_Closed {
	//	c.status = ConnStatus_OnlyRecv
	//} else {	// OnlyRecv || SendRecv
	//	//fmt.Printf("%s: 4444\n", c.recvConn.LocalID())
	//	return
	//}

	log.Printf("DualConn(%s) RecvLoop running\n", c.Name())

	var err error
	//var recv requires.Conn
	magic, msglen := uint32(0), uint32(0)
	for {

		//// 为了支持在RecvLoop中更换recvConn，需要在这里将其备份一次
		//c.recvLock.RLock()
		//recv = c.recvConn
		//c.recvLock.RUnlock()
		//// 但是这引入了新的问题：如果换的这一次：recv(旧)读不到数据，就一直堵着；读到数据但不完整出问题，就退出循环；

		log.Printf("DualConn(%s) RecvLoop read new msg start\n", c.Name())

		// 每次都置位
		c.recvLock.Lock()
		c.status.EnableRecv()
		c.recvLock.Unlock()

		// 循环读取数据包，解析成Message
		err = binary.Read(c.recvConn, binary.BigEndian, &magic)
		if err != nil {
			log.Printf("DualConn(%s) RecvLoop met error: %s\n", c.Name(), err)
			c.closeRecvConn()
			return	// 退出接收循环
		}
		if magic != defines.MessageMagicNumber {	// 说明对面不按格式发消息
			log.Printf("DualConn(%s) RecvLoop met error: magic(%x) != MessageMagicNumber(%x)\n",
				c.Name(), magic, defines.MessageMagicNumber)
			c.closeRecvConn()
			return
		}
		err = binary.Read(c.recvConn, binary.BigEndian, &msglen)
		if err != nil {
			log.Printf("DualConn(%s) RecvLoop met error: %s\n", c.Name(), err)
			c.closeRecvConn()
			return
		}
		msgbytes := make([]byte, msglen)
		err = binary.Read(c.recvConn, binary.BigEndian, msgbytes)
		if err != nil {
			log.Printf("DualConn(%s) RecvLoop met error: %s\n", c.Name(), err)
			c.closeRecvConn()
			return
		}

		msg := new(defines.Message)
		err = msg.Decode(msgbytes)
		if err != nil { // 遇到错误就断开连接
			log.Printf("DualConn(%s) RecvLoop met error: %s\n", c.Name(), err)
			c.closeRecvConn()
			return
		}
		log.Printf("DualConn(%s) RecvLoop: recv msg: %s\n", c.Name(), msg)
		// 塞到msgChan(来自Net.msgout)
		c.msgChan <- msg
		//fmt.Print("XXXXXXXXXXXXXXXXX")
	}
}

func (c *DualConn) closeRecvConn() {
	c.recvLock.Lock()
	c.recvConn.Close()
	c.status.DisableRecv()
	//c.recvConn = nil // 绝对不要写这句！
	c.recvLock.Unlock()
	// 如果recvConn已切换，旧的recvConn运行到此处，将新recvConn的status给复位了，
	// 解决办法就是，RecvLoop()每次循环开始都置位一次
	log.Printf("DualConn(%s) RecvLoop closed\n", c.Name())
}