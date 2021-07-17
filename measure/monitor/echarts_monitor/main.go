package main

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/measure/mdwriter/echarts"
	"github.com/azd1997/blockchain-consensus/measure/monitor"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
)

func main() {
	Run("127.0.0.1:9998", "127.0.0.1:9999")
}

func Run(tcpHost, httpHost string) {

	// 启动EchartsPageRunner
	go echarts.Run(httpHost)

	// 创建net模块
	msgChan := make(chan *defines.Message, 10)
	net, err := bnet.NewBNet(monitor.DefaultMonitorId, bnet.NetType_bTCP_Dual, tcpHost, msgChan) // 这个协议类型需要与共识集群使用的一致
	monitor.PanicErr(err)
	err = net.Init()
	monitor.PanicErr(err)

	// 启动Monitor
	monitor.Run(net, msgChan, echarts.DefaultPageRunner)
}
