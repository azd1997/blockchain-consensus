/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/10/20 2:20 PM
* @Description: The file is for
***********************************************************************/

package _default

import (
	"encoding/binary"
	"errors"
	"net"

	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/utils/address"
)

type TCPListener struct {
	ln *net.TCPListener
	localId string
	localListenAddr *net.TCPAddr
}

// address ipv4地址 形如"127.0.0.1:80"
func ListenTCP(localid string, addr string) (*TCPListener, error) {
	// 解析地址
	tcpaddr, err := address.ParseTCP4(addr)
	if err != nil {
		return nil, err
	}

	// 建立监听套接字
	ln, err := net.ListenTCP("tcp", tcpaddr)
	if err != nil {
		return nil, err
	}

	return &TCPListener{
		ln:      ln,
		localId: localid,
		localListenAddr:tcpaddr,
	}, nil
}

func (T *TCPListener) ok() bool {
	return T != nil && T.ln != nil && T.localListenAddr != nil
}

func (T *TCPListener) Network() string {
	return "tcp"
}

func (T *TCPListener) Accept() (requires.Conn, error) {
	if !T.ok() {
		return nil, errors.New("TCPListener not ok")
	}

	c, err := T.ln.Accept()
	if err != nil {
		return nil, err
	}

	rc := &TCPConn{
		conn:     c.(*net.TCPConn),
		localId:T.localId,
		localListenAddr:T.localListenAddr,
	}

	// 握手
	err = T.handshake(rc)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func (T *TCPListener) Close() error {
	if !T.ok() {
		return errors.New("TCPListener not ok")
	}
	return T.ln.Close()
}

func (T *TCPListener) LocalID() string {
	return T.localId
}

func (T *TCPListener) LocalListenAddr() net.Addr {
	return T.localListenAddr
}

// 握手
func (T *TCPListener) handshake(c *TCPConn) error {
	// 要读取第一个包出来，解析出对方身份，这一步相当于握手
	// TODO: 这里先简单的约定，建立连接之后需要发送如下报文：
	// idlen | id | addrlen | addr
	idlen := uint8(0)
	err := binary.Read(c.conn, binary.BigEndian, &idlen)
	if err != nil {
		return err
	}
	from := make([]byte, idlen)
	err = binary.Read(c.conn, binary.BigEndian, from)
	if err != nil {
		return err
	}
	addrlen := uint8(0)
	err = binary.Read(c.conn, binary.BigEndian, &addrlen)
	if err != nil {
		return err
	}
	fromaddrbytes := make([]byte, addrlen)
	err = binary.Read(c.conn, binary.BigEndian, fromaddrbytes)
	if err != nil {
		return err
	}
	fromaddr, err := address.ParseTCP4(string(fromaddrbytes))
	if err != nil {
		return err
	}

	c.remoteId = string(from)
	c.remoteListenAddr = fromaddr
	return nil
}