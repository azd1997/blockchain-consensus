/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/11/20 11:51 PM
* @Description: The file is for
***********************************************************************/

package bcc

import (
	"io"
	"testing"

	"github.com/azd1997/blockchain-consensus/requires"
)

type peer struct {
	id string
	addr string
	ln requires.Listener
	d requires.Dialer
}

func genPeer(id, addr string) (*peer, error) {
	ln, err := DefaultListener(id, addr)
	if err != nil {
		return nil, err
	}
	d, err := DefaultDialer(id, addr)
	if err != nil {
		return nil, err
	}
	return &peer{
		id:   id,
		addr: addr,
		ln:   ln,
		d:    d,
	}, nil
}

// 测试两个TCP peer间连接的建立和基本的消息发送（echo回显）
func TestTCP(t *testing.T) {
	ida := "peerA"
	addra := "127.0.0.1:8085"
	peera, err := genPeer(ida, addra)
	if err != nil {
		t.Errorf("%s: %s\n", ida, err)
	}
	defer peera.ln.Close()

	idb := "peerB"
	addrb := "127.0.0.1:8086"
	peerb, err := genPeer(idb, addrb)
	if err != nil {
		t.Errorf("%s: %s\n", idb, err)
	}
	defer peerb.ln.Close()

	// 先启动节点A，再启动节点B，节点B主动连接节点A
	doneA, doneB := make(chan struct{}), make(chan struct{})

	go func(p *peer) {
		n := 1
		for n > 0 {
			c, err := p.ln.Accept()
			if err != nil {
				t.Errorf("%s: %s\n", p.id, err)
				return
			}
			t.Logf("%s: Conn(%s<->%s) established\n", p.id, c.LocalID(), c.RemoteID())

			done := make(chan struct{})

			// 将读到的内容写到in中
			go func(c requires.Conn){	// 回显，直到c被对端关闭，然后自己这边也关闭
				io.Copy(c, c)	// 回显
				c.Close()		// 关闭连接
				t.Logf("%s: Conn(%s<->%s) closed\n", p.id, c.LocalID(), c.RemoteID())

				close(done)
			}(c)

			<- done
			n--
		}
		close(doneA)
	}(peera)

	go func(p *peer) {
		// 和peerA建立连接，而后发送Hello
		c, err := p.d.Dial(addra, ida)
		if err != nil {
			t.Errorf("%s: %s\n", idb, err)
			return
		}
		t.Logf("%s: Conn(%s<->%s) established\n", p.id, c.LocalID(), c.RemoteID())
		_, err = c.Write([]byte("Hello world!"))
		if err != nil {
			t.Errorf("%s: %s\n", idb, err)
			return
		}

		// 读取回显数据
		buf := make([]byte, 20)
		n, err := c.Read(buf)
		if err != nil {
			t.Errorf("%s: %s\n", idb, err)
			return
		}
		t.Logf("%s: Conn(%s<->%s) echo: %s\n", p.id, c.LocalID(), c.RemoteID(), string(buf[:n]))
		if string(buf[:n]) != "Hello world!" {
			t.Errorf("%s: echostr(%s) != 'Hello world!'\n", idb, string(buf[:n]))
		}

		c.Close()
		t.Logf("%s: Conn(%s<->%s) closed\n", p.id, c.LocalID(), c.RemoteID())
		close(doneB)
	}(peerb)

	<-doneA
	<-doneB
}
