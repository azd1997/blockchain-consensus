/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/14/20 10:38 AM
* @Description: 网络模块测试文件
***********************************************************************/

package btcp

import (
	"fmt"
	"github.com/azd1997/blockchain-consensus/modules/peerinfo/memorypit"
	"reflect"
	"testing"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/test"
)

// 运行节点 已经存在的
// 在内部根据配置启动节点(这里指Net)，而后返回其结构体
func runPeer(t *testing.T, id string, addr string,
	initF func(n *Net) error, msgHandleF func(n *Net, msg *defines.Message) error,
	neighbors ...*defines.PeerInfo) *Net {

	store := getKvStoreWithPairs(neighbors...)
	pit, err := memorypit.NewPeerInfoTable(id, store)
	if err != nil {
		t.Error(err)
	}
	err = pit.Init()
	if err != nil {
		t.Error(err)
	}

	opt := &Option{
		Id:                  id,
		Addr:                addr,
		Listener:            nil,
		Dialer:              nil,
		Pit:                 pit,
		MsgIn:               make(chan *defines.MessageWithError, 10),
		MsgOut:              make(chan *defines.Message, 10),
		CustomInitFunc:      initF,
		CustomMsgHandleFunc: msgHandleF,
	}

	peer, err := NewNet(opt)
	if err != nil {
		t.Error(err)
	}

	err = peer.Init()
	if err != nil {
		t.Error(err)
	}

	return peer
}

// 测试Net的创建与输入输出
func TestNet(t *testing.T) {

	// 初始化日志

	// peerA作为接受请求端
	peerA := runPeer(t, "peerA", "127.0.0.1:8091",
		nil,
		func(n *Net, msg *defines.Message) error {
			// peerA单纯地将msg回显
			fmt.Printf("22222222222222222\n")
			msg.From, msg.To = msg.To, msg.From
			merr := &defines.MessageWithError{
				Msg: msg,
				Err: make(chan error),
			}
			n.msgin <- merr
			fmt.Printf("merr: %v\n", <-merr.Err)

			//n.Close()
			return nil
		}, &defines.PeerInfo{
			Id:   "peerA",
			Addr: "127.0.0.1:8091",
			Data: nil,
		})

	fmt.Printf("0000000000000000\n")

	// peerB作为主动发信端
	peerB := runPeer(t, "peerB", "127.0.0.1:8092",
		func(n *Net) error { // 节点B作为主动的一方，需要主动与peerA发送消息
			fmt.Printf("0000111100001111\n")
			merr := &defines.MessageWithError{
				Msg: &defines.Message{
					Version: defines.CodeVersion,
					Type:    defines.MessageType_OneBlock,
					From:    "peerB",
					To:      "peerA",
					Sig:     []byte("signature"),
					Desc:    "test message",
				},
				Err: make(chan error),
			}
			n.msgin <- merr
			fmt.Printf("merr: %v\n", <-merr.Err)
			fmt.Printf("111111111111111\n")
			return nil
		},
		func(n *Net, msg *defines.Message) error {
			// 收到的消息进行比较
			if !reflect.DeepEqual(msg, &defines.Message{
				Version: defines.CodeVersion,
				Type:    defines.MessageType_OneBlock,
				From:    "peerA",
				To:      "peerB",
				Sig:     []byte("signature"),
				Desc:    "test message",
			}) {
				t.Error("error!!!!")
			}
			t.Log("success!!!!")

			//n.Close()
			return nil
		})

	// 等待1s，避免直接退出
	time.Sleep(1 * time.Second)
	peerA.Close()
	peerB.Close()
}

///////////////////////////////////

// 获取预先存有一些PeerInfo数据的test.Store
func getKvStoreWithPairs(infos ...*defines.PeerInfo) *test.Store {

	cf := requires.String2CF(memorypit.PeerInfoKeyPrefix)

	n := len(infos)
	keys := make([]string, n)
	values := make([]string, n)
	for i := 0; i < n; i++ {
		keys[i] = string(cf[:]) + infos[i].Id
		b, _ := infos[i].Encode()
		values[i] = string(b)
	}

	store := &test.Store{
		Cfs: map[requires.CF]bool{
			cf: true,
		},
		Kvs: map[string]string{},
	}

	for i := 0; i < n; i++ {
		store.Kvs[keys[i]] = values[i]
	}

	return store
}
