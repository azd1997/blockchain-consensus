package monitor

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/measure/common"
	"github.com/azd1997/blockchain-consensus/measure/mdwriter"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
)

const (
	DefaultMonitorId     = "Monitor"
	defaultMonitorLogMod = "MON"
)

// Monitor 监控器
type Monitor struct {
	BaseMonitor                 // 负责数据收集与计算
	writers []mdwriter.MDWriter // 负责对BaseMonitor的mdChan读取到的数据进行处理
	*log.Logger
}

func Run(net bnet.BNet, msgChan chan *defines.Message, writers ...mdwriter.MDWriter) {
	//msgChan := make(chan *defines.Message, 10)
	mdChan := make(chan common.MeasureData, 10)

	// 初始化日志
	//log.InitGlobalLogger(monitorId, true, true)
	//logger := log.NewLogger("MON", defaultMonitorId)

	// 构建net模块
	//net, err := bnet.NewBNet(defaultMonitorId, bnet.NetType_bTCP_Dual, netAddr, msgChan)	// 这个协议类型需要与共识集群使用的一致
	//PanicErr(err)
	//err = net.Init()
	//PanicErr(err)

	// 构建BaseMonitor
	bm := NewBaseMonitor(net, msgChan, mdChan)
	go bm.Run()

	// mdChan消费循环
	for {
		select {
		case md := <-mdChan:
			for _, w := range writers {
				err := w.MDWrite(md)
				PanicErr(err)
			}
		}
	}
}

func RunWithTxMaker(net bnet.BNet, msgChan chan *defines.Message, writers ...mdwriter.MDWriter) {
	//msgChan := make(chan *defines.Message, 10)
	mdChan := make(chan common.MeasureData, 10)

	// 构建BaseMonitor
	bm := NewBaseMonitor(net, msgChan, mdChan)
	go bm.Run()

	// mdChan消费循环
	for {
		select {
		case md := <-mdChan:
			for _, w := range writers {
				err := w.MDWrite(md)
				PanicErr(err)
			}
		}
	}
}

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}
