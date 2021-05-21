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

func testBNet(t *testing.T, network NetType) {
	addr1 := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9981}
	addr2 := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9982}
	id1, id2 := "peerA", "peerB"
	log.InitGlobalLogger(id1, true, true)
	log.InitGlobalLogger(id2, true, true)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		peerA, _ := NewBNet(id1, network, addr1.String(), make(chan *defines.Message, 100))
		if err := peerA.Init(); err != nil {
			t.Error(err)
		}
		time.Sleep(1 * time.Second)
		if err := peerA.Send(id2, addr2.String(), &defines.Message{
			Desc: "ping to #2: " + addr2.String(),
		}); err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	go func() {
		peerB, _ := NewBNet(id2, network, addr2.String(), make(chan *defines.Message, 100))
		if err := peerB.Init(); err != nil {
			t.Error(err)
		}
		time.Sleep(1 * time.Second)
		if err := peerB.Send(id1, addr1.String(), &defines.Message{
			Desc: "ping to #1: " + addr1.String(),
		}); err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	wg.Wait()

	time.Sleep(10 * time.Millisecond)
}

func TestBNet_bUDP(t *testing.T) {
	testBNet(t, NetType_bUDP)
}

func TestBNet_bTCP(t *testing.T) {
	testBNet(t, NetType_bTCP)
}

func TestBNet_bTCP_Dual(t *testing.T) {
	testBNet(t, NetType_bTCP_Dual)
}
