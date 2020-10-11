/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/10/20 1:54 PM
* @Description: The file is for
***********************************************************************/

package _default

import (
	"errors"
	"net"
	"time"
)

// 建立连接时需要告知接受连接那一方自己的ID

type TCPConn struct {
	conn *net.TCPConn
	localId, remoteId string
	localListenAddr, remoteListenAddr *net.TCPAddr
}

func (T *TCPConn) ok() bool {
	return T.conn != nil && T.localId != "" && T.remoteId != ""
}

func (T *TCPConn) Read(b []byte) (n int, err error) {
	if !T.ok() {
		return 0, errors.New("conn not ok")
	}
	return T.conn.Read(b)
}

func (T *TCPConn) Write(b []byte) (n int, err error) {
	if !T.ok() {
		return 0, errors.New("conn not ok")
	}
	return T.conn.Write(b)
}

func (T *TCPConn) Close() error {
	if !T.ok() {
		return errors.New("conn not ok")
	}
	return T.conn.Close()
}

func (T *TCPConn) LocalAddr() net.Addr {
	if !T.ok() {
		return nil
	}
	return T.conn.LocalAddr()
}

func (T *TCPConn) RemoteAddr() net.Addr {
	if !T.ok() {
		return nil
	}
	return T.conn.RemoteAddr()
}

func (T *TCPConn) SetDeadline(t time.Time) error {
	if !T.ok() {
		return errors.New("conn not ok")
	}
	return T.conn.SetDeadline(t)
}

func (T *TCPConn) SetReadDeadline(t time.Time) error {
	if !T.ok() {
		return errors.New("conn not ok")
	}
	return T.conn.SetReadDeadline(t)
}

func (T *TCPConn) SetWriteDeadline(t time.Time) error {
	if !T.ok() {
		return errors.New("conn not ok")
	}
	return T.conn.SetWriteDeadline(t)
}

func (T *TCPConn) Network() string {
	return "tcp"
}

func (T *TCPConn) LocalID() string {
	return T.localId
}

func (T *TCPConn) RemoteID() string {
	return T.remoteId
}

func (T *TCPConn) LocalListenAddr() net.Addr {
	return T.localListenAddr
}

func (T *TCPConn) RemoteListenAddr() net.Addr {
	return T.remoteListenAddr
}


