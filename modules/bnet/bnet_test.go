/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 6:19 PM
* @Description: The file is for
***********************************************************************/

package bnet

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"net"
	"sync"
	"testing"
	"time"
)

func TestBNet(t *testing.T) {

	addr1 := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9981}
	addr2 := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9982}
	id1, id2 := "peerA", "peerB"
	log.InitGlobalLogger(id1, true, true)
	log.InitGlobalLogger(id2, true, true)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		peerA, _ := NewBNet(id1, "udp", addr1.String(), make(chan *defines.Message, 100))
		if err := peerA.Init(); err != nil {
			t.Error(err)
		}
		time.Sleep(1 * time.Second)
		if err := peerA.Send(addr2.String(), &defines.Message{
			Desc: "ping to #2: " + addr2.String(),
		}); err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	go func() {
		peerB, _ := NewBNet(id2, "udp", addr2.String(), make(chan *defines.Message, 100))
		if err := peerB.Init(); err != nil {
			t.Error(err)
		}
		time.Sleep(1 * time.Second)
		if err := peerB.Send(addr1.String(), &defines.Message{
			Desc: "ping to #1: " + addr1.String(),
		}); err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	wg.Wait()

	time.Sleep(10 * time.Millisecond)
}
