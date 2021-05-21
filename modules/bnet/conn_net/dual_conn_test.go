/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 5/21/21 8:50 AM
* @Description: The file is for
***********************************************************************/

package conn_net

import (
	"reflect"
	"testing"

	"github.com/azd1997/blockchain-consensus/defines"
)

// TODO: 这个测试逻辑不适用于DualConn

// 测试连接的建立与消息的编解码
func TestDualConn(t *testing.T) {
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

	msgchanA := make(chan *defines.Message, 10)
	msgchanB := make(chan *defines.Message, 10)

	// A是服务端
	go func(p *peer) {
		n := 1
		for n > 0 {
			recvConn, err := p.ln.Accept()
			if err != nil {
				t.Errorf("%s: %s\n", p.id, err)
				return
			}
			t.Logf("%s: Conn(%s<->%s) established\n", p.id, recvConn.LocalID(), recvConn.RemoteID())

			////////////////////////////////
			// 将requires.Conn封装成bnet.Conn
			bconn := ToDualConn(nil, recvConn, msgchanA)
			go bconn.RecvLoop() // bconn循环接收Message
			go func() {         // 启动协程处理接收到的Message(打日志并回写)
				for {
					select {
					case msg := <-msgchanA:
						t.Logf("%s: Conn(%s<->%s): Recv Msg: %v\n", p.id, recvConn.LocalID(), recvConn.RemoteID(), msg)
						if err = bconn.Send(msg); err != nil {
							t.Errorf("%s: Conn(%s<->%s): Send Msg fail: %s\n", p.id, recvConn.LocalID(), recvConn.RemoteID(), err)
						}
						// 回发完消息后
						close(doneA)
					case <-doneA: // 关闭A时也要将此处退出
						t.Logf("%s: Conn(%s<->%s): Msg HandleLoop closed\n", p.id, recvConn.LocalID(), recvConn.RemoteID())
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
		bconn := ToConn(c, msgchanB)
		go bconn.RecvLoop() // bconn循环接收Message
		err = bconn.Send(testMsg)
		if err != nil {
			t.Errorf("%s: %s\n", idb, err)
			return
		}

		// 发送消息后，要等待对方把消息原样送回
		amsg := <-msgchanB

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


