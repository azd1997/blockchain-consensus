/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 6:19 PM
* @Description: The file is for
***********************************************************************/

package bnet

import (
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// 通常建议一个UDP包不要超过576B

const (
	longStringLength = 10 * 1024 	// 1M
)

func generateLongString(length int) string {
	str := "this is a long string: "
	length -= len(str)
	str += strings.Repeat("x", length)
	return str
}

func testBNet(t *testing.T, network NetType, longContent bool) {
	addr1 := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9981}
	addr2 := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 9982}
	id1, id2 := "peerA", "peerB"
	log.InitGlobalLogger(id1, true, true)
	log.InitGlobalLogger(id2, true, true)

	var wg sync.WaitGroup
	wg.Add(4)

	chanA := make(chan *defines.Message, 100)
	chanB := make(chan *defines.Message, 100)

	go func() {
		peerA, _ := NewBNet(id1, network, addr1.String(), chanA)
		if err := peerA.Init(); err != nil {
			t.Error(err)
		}
		time.Sleep(1 * time.Second)		// peerA先等一会，让peerB先发
		desc := "ping to #2: " + addr2.String()
		if longContent {
			desc += generateLongString(longStringLength)
		}
		if err := peerA.Send(id2, addr2.String(), &defines.Message{
			Desc: desc,
		}); err != nil {
			t.Error(err)
		}
		//fmt.Print("yyyy\n")
		//// 展示本机节点所有连接
		fmt.Print(peerA.DisplayAllConns(false))
		//fmt.Print("xxxx\n")

		wg.Done()
	}()

	go func() {
		peerB, _ := NewBNet(id2, network, addr2.String(), chanB)
		if err := peerB.Init(); err != nil {
			t.Error(err)
		}
		//time.Sleep(1 * time.Second)
		desc := "ping to #1: " + addr2.String()
		if longContent {
			desc += generateLongString(longStringLength)
		}
		if err := peerB.Send(id1, addr1.String(), &defines.Message{
			Desc: desc,
		}); err != nil {
			t.Error(err)
		}

		//fmt.Print("zzzz\n")
		//
		//// 展示本机节点所有连接
		fmt.Print(peerB.DisplayAllConns(false))
		//
		//fmt.Print("sssss\n")

		wg.Done()
	}()

	// 两个goroutine读接收到的数据
	go func() {
		data := <- chanA
		t.Logf("chanA recv: %v\n", data)
		wg.Done()
	}()
	go func() {
		data := <- chanB
		t.Logf("chanB recv: %v\n", data)
		wg.Done()
	}()

	wg.Wait()

	time.Sleep(10 * time.Millisecond)
}

func TestBNet_bUDP(t *testing.T) {
	testBNet(t, NetType_bUDP, false)
}

func TestBNet_bUDP_LongContent(t *testing.T) {
	testBNet(t, NetType_bUDP, true)
}
// longStringLength = 1MB 时
// write udp 127.0.0.1:9981->127.0.0.1:9982: wsasendto: A message sent on a datagram socket was larger than the internal message buffer or some other network limit, or the buffer used to receive a datagram into was smaller than the datagram itself.
// longStringLength = 1KB 时
// 可行，抓包得到UDP包大小1270，没有超过1500
// longStringLength = 2KB 时
// 可行，抓包得到UDP包大小2294，没有超过1500

func TestBNet_bTCP(t *testing.T) {
	testBNet(t, NetType_bTCP, false)
}

func TestBNet_bTCP_Dual(t *testing.T) {
	testBNet(t, NetType_bTCP_Dual, false)
}
