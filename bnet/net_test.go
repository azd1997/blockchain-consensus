/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/14/20 10:38 AM
* @Description: 网络模块测试文件
***********************************************************************/

package bnet

import (
	"reflect"
	"testing"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
	_default "github.com/azd1997/blockchain-consensus/requires/default"
	"github.com/azd1997/blockchain-consensus/test"

	"github.com/agiledragon/gomonkey"
)

// 测试Net的创建与输入输出
func TestNet(t *testing.T) {

	// 打桩
	patches := gomonkey.ApplyMethod(
		reflect.TypeOf(),

		)

	id := "myid"
	addr := "127.0.0.1:8085"

	ln, err := _default.ListenTCP(id, addr)
	if err != nil {
		t.Error(err)
	}


	dialer, err := _default.NewDialer(id, addr, 0)
	if err != nil {
		t.Error(err)
	}

	// store
	store := getKvStoreWithPairs(&defines.PeerInfo{
		Id:   "seeda",
		Addr: "seeda-addr",
		Data: nil,
	})

	msgin := make(chan *defines.Message)
	msgout := make(chan *defines.Message)
	go func() {		// 一个简单的消息处理goroutine
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Data,
			From:    "from",
			To:      "to",
		}
		// 首先是向msgin写入一个消息, Net模块会将该消息通过相应的Conn发送出去
		// （所以Conn.Send会打桩，不向网络发送，而是直接编码后写回Conn，这样避免测试时总是要建多个节点）
		msgin <- msg
		// 接着尝试从msgout读消息
		amsg := <-msgout
		if !reflect.DeepEqual(msg, amsg) {
			t.Error("errorrrrr")
		}
	}()

	// 创建
	netmod, err := NewNet(ln, dialer, store, msgin, msgout)
	if err != nil {
		t.Error(err)
	}
	// 手动地创建一个连接出来


	// 处理输入
	// 这里我们只需要验证消息会通过msgin然后分发到对应的conn去（相信conn会做好接下来的网络发送工作）
	// 并且这里会向msgin写入一个消息，由Net处理输入，输出到消息处理模块处理消息（这里整个回显）

	// 处理输出
	// 验证循环接收
}

///////////////////////////////////

// 获取预先存有一些PeerInfo数据的test.Store
func getKvStoreWithPairs(infos ...*defines.PeerInfo) *test.Store {

	cf := requires.String2CF(PeerInfoKeyPrefix)

	n := len(infos)
	keys := make([]string, n)
	values := make([]string, n)
	for i:=0; i<n; i++ {
		keys[i] = string(cf[:]) + infos[i].Id
		b, _  := infos[i].Encode()
		values[i] = string(b)
	}

	store := &test.Store{
		Cfs: map[requires.CF]bool{
			cf:true,
		},
		Kvs: map[string]string{},
	}

	for i:=0; i<n; i++ {
		store.Kvs[keys[i]] = values[i]
	}

	return store
}