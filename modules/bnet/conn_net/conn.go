/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/11/20 11:23 AM
* @Description: bcc库内部封装的Conn
***********************************************************************/

package conn_net

import (
	"bytes"
	"fmt"
	"github.com/azd1997/blockchain-consensus/utils/binary"
	"log"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
)

const (
	// 默认连接中的读写操作超时
	DefaultConnTimeout time.Duration = 5 * time.Second
)



// Conn 连接
/*
	用法：
	1. 获取c(requires.Conn)
	2. 调用bconn := ToConn(c)
	3. go bconn.RecvLoop()
	4. 需要发送Message时调bconn.Send()
	5. Message处理方需要读bconn.MsgChan()

	// 注意：由于基于TCP C/S架构实现P2P，为了避免连接建立的冲突，将连接视为 【单向连接】

*/
type Conn struct {
	conn requires.Conn

	// msgChan为Net也就是多个Conn的管理结构所提供，在多个Conn间共享该msgChan
	// Conn只负责写msgChan，Net会另起goroutine循环读msgChan
	msgChan chan<- *defines.Message // Message channel

	status  ConnStatus    // Conn当前状态
	timeout time.Duration // 超时
}

// ToConn 将requires.Conn封装成bcc.Conn
func ToConn(conn requires.Conn, recvmsg chan<- *defines.Message) *Conn {
	if conn == nil {
		return nil
	}
	c := &Conn{
		conn:    conn,
		msgChan: recvmsg,
		status:  ConnStatus_OnlySend,
		timeout: DefaultConnTimeout,
	}
	return c
}

// Status 报告Conn状态
func (c *Conn) Status() string {
	return c.status.String()
}

// Name Conn名称
func (c *Conn) Name() string {
	return fmt.Sprintf("[%s]<->[%s](%s)", c.conn.LocalID(), c.conn.RemoteID(), c.conn.HashCode())
}

// String 描述
func (c *Conn) String() string {
	return fmt.Sprintf("Conn %s info: {Network: %s, From: %s(%s), To: %s(%s), Status: %s, Timeout: %s}",
		c.Name(),
		c.conn.Network(), c.conn.LocalID(), c.conn.LocalAddr().String(),
		c.conn.RemoteID(), c.conn.RemoteAddr().String(), c.status.String(), c.timeout.String())
}

// Close 关闭连接
func (c *Conn) Close() error {
	return c.conn.Close()
}

// Send 发送消息
func (c *Conn) Send(msg *defines.Message) error {
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
	_, err = c.conn.Write(buf2.Bytes())
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
func (c *Conn) RecvLoop() {
	c.status = ConnStatus_SendRecv
	log.Printf("Conn(%s) running\n", c.Name())

	var err error
	magic, msglen := uint32(0), uint32(0)
	for {
		// 循环读取数据包，解析成Message
		err = binary.Read(c.conn, binary.BigEndian, &magic)
		if err != nil {
			log.Printf("Conn(%s) met error: %s\n", c.Name(), err)
			c.closeConn()
			return	// 退出接收循环
		}
		if magic != defines.MessageMagicNumber {	// 说明对面不按格式发消息
			log.Printf("Conn(%s) RecvLoop met error: magic(%x) != MessageMagicNumber(%x)\n",
				c.Name(), magic, defines.MessageMagicNumber)
			c.closeConn()
			return
		}
		err = binary.Read(c.conn, binary.BigEndian, &msglen)
		if err != nil {
			log.Printf("Conn(%s) met error: %s\n", c.Name(), err)
			c.closeConn()
			return	// 退出接收循环
		}
		msgbytes := make([]byte, msglen)
		err = binary.Read(c.conn, binary.BigEndian, msgbytes)
		if err != nil {
			log.Printf("Conn(%s) met error: %s\n", c.Name(), err)
			c.closeConn()
			return	// 退出接收循环
		}

		msg := new(defines.Message)
		err = msg.Decode(msgbytes)
		if err != nil { // 遇到错误就断开连接
			log.Printf("Conn(%s) met error: %s\n", c.Name(), err)
			c.closeConn()
			return	// 退出接收循环
		}
		log.Printf("Conn(%s) RecvLoop: recv msg: %s\n", c.Name(), msg)
		// 塞到msgChan(来自Net.msgout)
		c.msgChan <- msg
	}
}

func (c *Conn) closeConn() {
	c.conn.Close()
	c.status = ConnStatus_Closed
	log.Printf("Conn(%s) RecvLoop closed\n", c.Name())
}

// MsgChan 对外提供只读的msgChan
//func (c *Conn) MsgChan() <-chan *defines.Message {
//	return c.msgChan
//}
