/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 4:56 PM
* @Description: The file is for
***********************************************************************/

package budp

import (
	"github.com/azd1997/blockchain-consensus/log"
	"sync"
	"testing"
	"time"
)

func TestUDPNet(t *testing.T) {

	addr1 := "127.0.0.1:9981"
	addr2 := "127.0.0.1:9982"
	id1, id2 := "peerA", "peerB"
	log.InitGlobalLogger(id1, true, true)
	log.InitGlobalLogger(id2, true, true)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		logger := log.NewLogger("UDP", id1)
		peerA, _ := NewUDPNet(id1, addr1, logger, make(chan []byte, 100))
		if err := peerA.Init(); err != nil {
			t.Error(err)
		}
		time.Sleep(1 * time.Second)
		if err := peerA.Send(addr2, []byte("ping to #2: "+addr2)); err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	go func() {
		logger := log.NewLogger("UDP", id2)
		peerB, _ := NewUDPNet(id2, addr2, logger, make(chan []byte, 100))
		if err := peerB.Init(); err != nil {
			t.Error(err)
		}
		time.Sleep(1 * time.Second)
		if err := peerB.Send(addr1, []byte("ping to #1: "+addr1)); err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	wg.Wait()

	time.Sleep(10 * time.Millisecond)
}
