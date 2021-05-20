package conn_net

import (
	"fmt"
	"log"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/utils/binary"
)

type DualConn struct {
	sendConn, recvConn requires.Conn	// 两条实际连接，一条负责发送，一条负责接收

	// msgChan为Net也就是多个Conn的管理结构所提供，在多个Conn间共享该msgChan
	// Conn只负责写msgChan，Net会另起goroutine循环读msgChan
	msgChan chan<- *defines.Message // Message channel 从recv读到该chan

	status  ConnStatus    // Conn当前状态
	timeout time.Duration // 超时
}

// ToDualConn 将requires.Conn封装成DualConn
func ToDualConn(send, recv requires.Conn, recvmsg chan<- *defines.Message) *DualConn {
	if send == nil && recv == nil {		// 不能都空
		return nil
	}
	c := &DualConn{
		sendConn:send,
		recvConn:recv,
		msgChan: recvmsg,
		status:  ConnStatus_Ready,
		timeout: DefaultConnTimeout,
	}
	return c
}

// Status 报告Conn状态
func (c *DualConn) Status() string {
	return c.status.String()
}

// Name Conn名称
func (c *DualConn) Name() string {
	if c.sendConn != nil {
		return fmt.Sprintf("[%s]<->[%s]", c.sendConn.LocalID(), c.sendConn.RemoteID())
	} else if c.recvConn != nil {
		return fmt.Sprintf("[%s]<->[%s]", c.recvConn.LocalID(), c.recvConn.RemoteID())
	} else {
		return ""
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

	return fmt.Sprintf("Conn info: {Network: %s, From: %s(%s,%s), To: %s(%s,%s), Status: %s, Timeout: %s}",
		availableConn.Network(),
		availableConn.LocalID(), availableConn.LocalListenAddr().String(), availableConn.LocalAddr().String(),
		availableConn.RemoteID(), availableConn.RemoteListenAddr().String(), availableConn.RemoteAddr().String(),
		c.status.String(), c.timeout.String())
}

// Close 关闭连接
func (c *DualConn) Close() error {
	if c.sendConn != nil {c.sendConn.Close()}
	if c.recvConn != nil {c.recvConn.Close()}
	c.status = ConnStatus_Closed
	return nil
}

// Send 发送消息
func (c *DualConn) Send(msg *defines.Message) error {
	if c.sendConn == nil {
		log.Printf("Conn(%s) Send quit: send_conn == nil \n", c.Name())
		return nil
	}

	// 序列化为字节数组
	b, err := msg.Encode()
	if err != nil {
		return err
	}

	// 发送
	_, err = c.sendConn.Write(b)
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
		log.Printf("Conn(%s) RecvLoop quit: recv_conn == nil \n", c.Name())
		return
	}

	c.status = ConnStatus_Running
	log.Printf("Conn(%s) running\n", c.Name())

	var err error
	magic, msglen := uint32(0), uint32(0)
	for {
		// 循环读取数据包，解析成Message
		err = binary.Read(c.recvConn, binary.BigEndian, &magic)
		if err != nil {
			log.Printf("Conn(%s) met error: %s\n", c.Name(), err)

			continue
		}
		err = binary.Read(c.recvConn, binary.BigEndian, &msglen)
		if err != nil {
			log.Printf("Conn(%s) met error: %s\n", c.Name(), err)
			continue
		}
		msgbytes := make([]byte, msglen)
		err = binary.Read(c.recvConn, binary.BigEndian, msgbytes)
		if err != nil {
			log.Printf("Conn(%s) met error: %s\n", c.Name(), err)
			continue
		}

		msg := new(defines.Message)
		err = msg.Decode(msgbytes)
		if err != nil { // 遇到错误就断开连接
			log.Printf("Conn(%s) met error: %s\n", c.Name(), err)
			c.status = ConnStatus_Closed
			return
		}
		// 塞到msgChan(来自Net.msgout)
		c.msgChan <- msg
	}
}

