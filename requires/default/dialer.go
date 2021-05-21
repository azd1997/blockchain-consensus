/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/10/20 4:53 PM
* @Description: 默认的Dialer实现(TCP)
***********************************************************************/

package _default

import (
	"errors"
	"log"
	"net"
	"time"

	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/utils/address"
	"github.com/azd1997/blockchain-consensus/utils/binary"
)

func NewDialer(id, addr string, timeout time.Duration) (*TCPDialer, error) {
	localAddr, err := address.ParseTCP4(addr)
	if err != nil {
		return nil, err
	}

	d := &net.Dialer{
		Timeout: timeout,
	}

	return &TCPDialer{
		d:               d,
		localListenAddr: localAddr,
		localId:         id,
	}, nil
}

type TCPDialer struct {
	d               *net.Dialer
	localListenAddr *net.TCPAddr
	localId         string
}

func (T *TCPDialer) ok() bool {
	return T != nil && T.d != nil && T.localListenAddr != nil
}

func (T *TCPDialer) Dial(addr, remoteId string) (requires.Conn, error) {
	if !T.ok() {
		return nil, errors.New("TCPDialer not ok")
	}

	conn, err := T.d.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	remoteAddr, err := address.ParseTCP4(addr)
	if err != nil {
		return nil, err
	}

	rc := &TCPConn{
		conn:             conn.(*net.TCPConn),
		localId:          T.localId,
		remoteId:         remoteId,
		localListenAddr:  T.localListenAddr,
		remoteListenAddr: remoteAddr,
	}

	// 握手
	err = T.handshake(rc)
	if err != nil {
		return nil, err
	}

	log.Printf("dialer dial {local: %s(%s), remote: %s(%s)} succ\n", T.localId, T.localListenAddr, remoteId, remoteAddr)
	return rc, nil
}

func (T *TCPDialer) Network() string {
	return "tcp"
}

func (T *TCPDialer) LocalID() string {
	return T.localId
}

func (T *TCPDialer) LocalListenAddr() net.Addr {
	return T.localListenAddr
}

// 握手
func (T *TCPDialer) handshake(c *TCPConn) error {
	// 建立连接后紧跟着发送 “idlen | id | addrlen | addr” 消息
	idlen := uint8(len(T.localId))
	localAddr := T.localListenAddr.String()
	addrlen := uint8(len(localAddr))
	err := binary.Write(c.conn, binary.BigEndian, idlen, []byte(T.localId), addrlen, []byte(localAddr))
	if err != nil {
		return err
	}
	return nil
}
