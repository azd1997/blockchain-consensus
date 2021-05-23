package monitor

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/bnet/cs_net"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// EchartsMonitor 基于Echarts的监控器
// 各节点通过TCP将数据上报给EchartsMonitor，
// EchartsMonitor统计分析后得到性能指标，通过websocket更新至echarts图表
type EchartsMonitor struct {
	// TCPServer
	tcpServer *cs_net.Server
	msgchan chan *defines.Message

	// 用于区块数据记录

	// 用于记录交易产生量
	txGenCounter int64
}

func (m *EchartsMonitor) Init() {

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

func (m *EchartsMonitor) ServeWebsocket() {
	// create a new line instance
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeChalk}),
		charts.WithTitleOpts(opts.Title{Title: "Line example"}))

	// Put data into instance
	values1, values2 := generateRandValues(7, 300), generateRandValues(7, 3000)
	line.SetXAxis(days).
		AddSeries("Category A", generateLineItems(values1)).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	// Setup handler and obtains data that we're going to be using
	// to pass updates to be written over ws.
	wsHandler, dataC := ws.Handler()
	defer close(dataC)
	updateChart := func() {
		days = append(days[1:], days[0])
		values1 = append(values1[1:], rand.Intn(300))
		values2 = append(values2[1:], rand.Intn(300))
		line.MultiSeries = line.MultiSeries[:0]
		line.SetXAxis(days).
			AddSeries("Category A", generateLineItems(values1))
		line.Validate()
		dataC <- line.JSON()
	}

	go func() {
		for {
			select {
			case <-time.After(time.Second):
				updateChart()
			}
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { ws.Render(w, line, line.ChartID, r.Host) })
	http.HandleFunc("/ws", wsHandler)
	log.Println("Visit http://localhost:8080 to see the live chart")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func generateRandValues(len, max int) (values []int) {
	for i := 0; i < len; i++ {
		values = append(values, rand.Intn(max))
	}
	return
}



var days = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

