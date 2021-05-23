package cs_net

import (
	"fmt"
	"testing"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
)

func TestCS(t *testing.T) {
	serverId := "server"
	serverListenAddr := "127.0.0.1:9998"
	clientId := "client"
	clientListenAddr := "127.0.0.1:9999"
	serverMsgout := make(chan *defines.Message, 100)
	clientMsgout := make(chan *defines.Message, 100)

	// 先初始化全局日志器底层模块
	log.InitGlobalLogger(serverId, true, true)	// Server用

	// 启动服务器
	logger := log.NewLogger("SRV", serverId)
	server, err := NewServer(serverId, serverListenAddr, serverMsgout, logger)
	if err != nil {
		panic(err)
	}
	err = server.Init()
	if err != nil {
		panic(err)
	}

	// 启动客户端
	client, err := Dial(serverId, serverListenAddr, clientId, clientListenAddr, clientMsgout)
	if err != nil {
		panic(err)
	}

	times := 10

	// 循环写10次消息
	for i:=1;i<=times;i++ {
		err = client.Send(&defines.Message{
			Desc: fmt.Sprintf("#%d", i),
		})
	}

	// 循环读10次消息
	for i:=1;i<=times;i++ {
		msg := <- serverMsgout
		fmt.Printf("#%d msg.Desc: %v\n", i, msg.Desc)
	}
}
