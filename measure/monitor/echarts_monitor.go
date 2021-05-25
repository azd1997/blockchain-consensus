package monitor

import (
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/measure/calculator"
	"github.com/azd1997/blockchain-consensus/measure/common"
	"github.com/azd1997/blockchain-consensus/measure/echarts"
	"github.com/azd1997/blockchain-consensus/modules/bnet/cs_net"
)

func Run(tcpHost, httpHost string) {
	// 启动TCP服务器
	msgchan := make(chan *defines.Message)
	monitorId := "Monitor"
	log.InitGlobalLogger(monitorId, true, true)
	logger := log.NewLogger("MON", monitorId)
	tcpServer, err := cs_net.NewServer(monitorId, tcpHost, msgchan, logger)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	err = tcpServer.Init()
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	// 使能calculator
	mdChan := make(chan common.MeasureData, 10)
	calculator.SetMDChan(mdChan)

	// 启动msgChan消费通道
	go func() {
		for msg := range msgchan {

			fmt.Printf("Monitor recv message: %s\n", msg.String())

			if msg.Type == defines.MessageType_NewBlock {
				if len(msg.Data) == 0 {
					fmt.Println("errrr!!!!")
					continue
				}
				nb := new(defines.Block)
				if err := nb.Decode(msg.Data[0]); err != nil {
					fmt.Println("err: ", err)
					continue
				}
				calculator.AddBlock(nb)
			} else if msg.Type == defines.MessageType_Txs {
				if len(msg.Data) == 0 {
					fmt.Println("errrr!!!!")
					continue
				}
				for i:=0; i<len(msg.Data); i++ {
					tx := new(defines.Transaction)
					if err := tx.Decode(msg.Data[i]); err != nil {
						fmt.Println("err: ", err)
						continue
					}
					calculator.AddTx(tx)
				}
			}
		}
	}()

	// 启动echarts
	echarts.Run(httpHost, mdChan)
}

//// EchartsMonitor 基于Echarts的监控器
//// 各节点通过TCP将数据上报给EchartsMonitor，
//// EchartsMonitor统计分析后得到性能指标，通过websocket更新至echarts图表
//type EchartsMonitor struct {
//	// TCPServer
//	tcpServer *cs_net.Server
//	msgchan chan *defines.Message
//
//	// echarts读，calculator写
//	mdataChan chan common.MeasureData
//
//	// TODO: 数据库用于存储确定的区块数据
//
//	// TODO: 数据库用于存储计算出来的MeasureData
//
//	// 用于区块数据记录
//
//	// 用于记录交易产生量
//	txGenCounter int64
//}
//
//func (m *EchartsMonitor) Init() {
//
//}
//
//// 在MeasureLoop中对数据进行计算，并构建成MeasureData发给mdataChan
//func (m *EchartsMonitor) MeasureLoop() {
//
//}
//
//func (m *EchartsMonitor) CollectLoop() {
//	for msg := range m.msgchan {
//		// 处理msg
//		if msg.Type == defines.MessageType_NewBlock {
//			m.handleMsgNewBlock(msg)
//		} else if msg.Type == defines.MessageType_Txs {
//			m.handleMsgTxs(msg)
//		} else {	// 其他，不处理
//			continue
//		}
//	}
//}
//
//func (m *EchartsMonitor) handleMsgNewBlock(msg *defines.Message) {
//
//}
//
//func (m *EchartsMonitor) handleMsgTxs(msg *defines.Message) {
//	m.txGenCounter += int64(len(msg.Data))
//}
//
//func (m *EchartsMonitor) echartsLoop() {
//	// 准备好websocket handler用于与html中websocket脚本通信
//	// 准备好echarts render方法（需要指定chartId，因为自定义了脚本）
//	// 准备好echarts图表
//}
