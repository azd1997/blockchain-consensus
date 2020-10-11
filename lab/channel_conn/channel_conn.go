/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/8/20 10:35 AM
* @Description: 基于channel实现的Conn
***********************************************************************/

package channel_conn

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"net"
	"time"
)

// ChannelConn 基于channel的Conn
type ChannelConn struct {
	local string
	remote string
	reqChan chan defines.Message
}

func (c ChannelConn) Read(b []byte) (n int, err error) {
	panic("implement me")
}

func (c ChannelConn) Write(b []byte) (n int, err error) {
	panic("implement me")
}

func (c ChannelConn) Close() error {
	panic("implement me")
}

func (c ChannelConn) LocalAddr() net.Addr {
	panic("implement me")
}

func (c ChannelConn) RemoteAddr() net.Addr {
	panic("implement me")
}

func (c ChannelConn) SetDeadline(t time.Time) error {
	panic("implement me")
}

func (c ChannelConn) SetReadDeadline(t time.Time) error {
	panic("implement me")
}

func (c ChannelConn) SetWriteDeadline(t time.Time) error {
	panic("implement me")
}

