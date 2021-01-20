/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 3:21 PM
* @Description: The file is for
***********************************************************************/

package budp

import (
	"bytes"
	"errors"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/utils/binary"
	"net"

	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/utils/netutil"
)

// uninit -> inited (running) -> closed
// network: udp
type UDPNet struct {
	id   string
	addr string

	listener *net.UDPConn

	inited bool
	closed bool

	msgout chan []byte

	done chan struct{}

	*log.Logger
}

func NewUDPNet(id string, addr string, logger *log.Logger, msgchan chan []byte) (*UDPNet, error) {
	return &UDPNet{
		id:       id,
		addr:     addr,
		listener: nil,
		inited:   false,
		closed:   false,
		msgout:   msgchan,
		done:     make(chan struct{}),
		Logger:   logger,
	}, nil
}

func (un *UDPNet) ID() string {
	return un.id
}

func (un *UDPNet) Network() string {
	return "udp"
}

func (un *UDPNet) Addr() string {
	return un.addr
}

func (un *UDPNet) Init() error {
	// init流程
	lip, lport := netutil.ParseIPPort(un.addr)
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: lip, Port: lport})
	if err != nil {
		return err
	}
	un.listener = listener

	// 启动接收循环
	go un.RecvLoop()

	un.inited = true
	return nil
}

func (un *UDPNet) Inited() bool {
	return un.inited
}

func (un *UDPNet) Ok() bool {
	return true
}

func (un *UDPNet) Close() error {
	// close流程
	close(un.done)

	un.closed = true
	return nil
}

func (un *UDPNet) Closed() bool {
	return un.closed
}

func (un *UDPNet) Send(raddr string, msg []byte) error {
	if !un.inited || un.closed {
		return errors.New("UDPNet is not running")
	}

	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, defines.MessageMagicNumber, uint32(len(msg)), msg); err != nil {
		return err
	}

	rip, rport := netutil.ParseIPPort(raddr)
	n, err := un.listener.WriteTo(buf.Bytes(), &net.UDPAddr{IP: rip, Port: rport})
	if err != nil {
		un.Errorf("UDPNet send msg fail. raddr=%s, err=%s, msg=%v", raddr, err, msg)
		return err
	}
	un.Debugf("UDPNet send msg succ. raddr=%s, n=%d, msg=%v", raddr, n, msg)
	return nil
}

func (un *UDPNet) MsgOut() chan []byte {
	return un.msgout
}

func (un *UDPNet) RecvLoop() {
	magic, msglen := uint32(0), uint32(0)
	for {
		buf := make([]byte, defines.MaxMessageLen)
		n, raddr, err := un.listener.ReadFrom(buf)
		if err != nil {
			un.Errorf("n=%d, raddr=%s, err=%s", n, raddr, err)
			continue
		}

		r := bytes.NewReader(buf[:n])

		if err := binary.Read(r, binary.BigEndian, &magic); err != nil {
			un.Errorf("UDPNet(%s) met error: %s\n", un.addr, err)
			continue
		}
		if magic != defines.MessageMagicNumber {
			continue
		}
		if err := binary.Read(r, binary.BigEndian, &msglen); err != nil {
			un.Errorf("UDPNet(%s) met error: %s\n", un.addr, err)
			continue
		}
		msgbytes := make([]byte, msglen)
		if err := binary.Read(r, binary.BigEndian, msgbytes); err != nil {
			un.Errorf("UDPNet(%s) met error: %s\n", un.addr, err)
			continue
		}
		un.Debugf("UDPNet(%s) recv msg from (%s): msg=%v\n", un.addr, raddr, msgbytes)
		un.msgout <- msgbytes
	}
}
