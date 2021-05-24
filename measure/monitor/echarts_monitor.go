package monitor

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/measure/echarts"
	"github.com/azd1997/blockchain-consensus/modules/bnet/cs_net"
)

// EchartsMonitor 基于Echarts的监控器
// 各节点通过TCP将数据上报给EchartsMonitor，
// EchartsMonitor统计分析后得到性能指标，通过websocket更新至echarts图表
type EchartsMonitor struct {
	// TCPServer
	tcpServer *cs_net.Server
	msgchan chan *defines.Message
	mdataChan chan echarts.MeasureData

	// TODO: 数据库用于存储确定的区块数据

	// TODO: 数据库用于存储计算出来的MeasureData

	// 用于区块数据记录

	// 用于记录交易产生量
	txGenCounter int64
}

func (m *EchartsMonitor) Init() {

}

// 在MeasureLoop中对数据进行计算，并构建成MeasureData发给mdataChan
func (m *EchartsMonitor) MeasureLoop() {

}

func (m *EchartsMonitor) CollectLoop() {
	for msg := range m.msgchan {
		// 处理msg
		if msg.Type == defines.MessageType_NewBlock {
			m.handleMsgNewBlock(msg)
		} else if msg.Type == defines.MessageType_Txs {
			m.handleMsgTxs(msg)
		} else {	// 其他，不处理
			continue
		}
	}
}

func (m *EchartsMonitor) handleMsgNewBlock(msg *defines.Message) {

}

func (m *EchartsMonitor) handleMsgTxs(msg *defines.Message) {
	m.txGenCounter += int64(len(msg.Data))
}

func (m *EchartsMonitor) echartsLoop() {
	// 准备好websocket handler用于与html中websocket脚本通信
	// 准备好echarts render方法（需要指定chartId，因为自定义了脚本）
	// 准备好echarts图表
}
