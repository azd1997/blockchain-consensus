/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/11/20 10:26 PM
* @Description: 测试进一步封装的Conn的建立与消息收发
***********************************************************************/

package bnet

import (
	_default "github.com/azd1997/blockchain-consensus/requires/default"
	"reflect"
	"testing"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
)

type peer struct {
	id string
	addr string
	ln requires.Listener
	d requires.Dialer
}

func genPeer(id, addr string) (*peer, error) {
	ln, err := _default.ListenTCP(id, addr)
	if err != nil {
		return nil, err
	}
	d, err := _default.NewDialer(id, addr, 0)
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

// 测试连接的建立与消息的编解码
func TestConn(t *testing.T) {
	ida := "peerA"
	addra := "127.0.0.1:8083"
	peera, err := genPeer(ida, addra)
	if err != nil {
		t.Errorf("%s: %s\n", ida, err)
	}
	defer peera.ln.Close()

	idb := "peerB"
	addrb := "127.0.0.1:8084"
	peerb, err := genPeer(idb, addrb)
	if err != nil {
		t.Errorf("%s: %s\n", idb, err)
	}
	defer peerb.ln.Close()

	// 先启动节点A，再启动节点B，节点B主动连接节点A
	doneA, doneB := make(chan struct{}), make(chan struct{})

	// A是服务端
	go func(p *peer) {
		n := 1
		for n > 0 {
			c, err := p.ln.Accept()
			if err != nil {
				t.Errorf("%s: %s\n", p.id, err)
				return
			}
			t.Logf("%s: Conn(%s<->%s) established\n", p.id, c.LocalID(), c.RemoteID())

			////////////////////////////////
			// 将requires.Conn封装成bnet.Conn
			bconn := ToConn(c)
			go bconn.RecvLoop()		// bconn循环接收Message
			go func() {		// 启动协程处理接收到的Message(打日志并回写)
				for {
					select{
					case msg := <- bconn.MsgChan():
						t.Logf("%s: Conn(%s<->%s): Recv Msg: %v\n", p.id, c.LocalID(), c.RemoteID(), msg)
						if err = bconn.Send(msg); err != nil {
							t.Errorf("%s: Conn(%s<->%s): Send Msg fail: %s\n", p.id, c.LocalID(), c.RemoteID(), err)
						}
						// 回发完消息后
						close(doneA)
					case <-doneA:	// 关闭A时也要将此处退出
						t.Logf("%s: Conn(%s<->%s): Msg HandleLoop closed\n", p.id, c.LocalID(), c.RemoteID())
						return
					}
				}
			}()

			////////////////////////////////

			n--
		}
	}(peera)

	// B是客户端
	go func(p *peer) {
		// 和peerA建立连接，而后发送Hello
		c, err := p.d.Dial(addra, ida)
		if err != nil {
			t.Errorf("%s: %s\n", idb, err)
			return
		}
		t.Logf("%s: Conn(%s<->%s) established\n", p.id, c.LocalID(), c.RemoteID())

		////////////////////////////////
		// 将requires.Conn封装成bnet.Conn
		bconn := ToConn(c)
		go bconn.RecvLoop()		// bconn循环接收Message
		err = bconn.Send(testMsg)
		if err != nil {
			t.Errorf("%s: %s\n", idb, err)
			return
		}

		// 发送消息后，要等待对方把消息原样送回
		amsg := <-bconn.MsgChan()

		if !reflect.DeepEqual(testMsg, amsg) {
			t.Errorf("%s: testMsg != amsg(%v)\n", idb, amsg)
			return
		}

		// 关闭连接
		err = bconn.Close()
		if err != nil {
			t.Errorf("%s: %s\n", idb, err)
			return
		}

		t.Logf("%s: Conn(%s<->%s) closed\n", p.id, c.LocalID(), c.RemoteID())
		close(doneB)

		////////////////////////////////
	}(peerb)

	<-doneB
	<-doneA
}

var testMsg = &defines.Message{
	Version:defines.CodeVersion,
	Type:defines.MessageType_Data,
	From:"id_peerb",
	To:"id_peera",
	Sig:[]byte("Signature"),
	Entries:[]*defines.Entry{
		&defines.Entry{
			BaseIndex: 0,
			Base:      []byte("Base"),
			Type:      defines.EntryType_Block,
			Data:      []byte("block"),
		},
	},
}