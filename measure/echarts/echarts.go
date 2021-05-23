package echarts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
	"net/http"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/gorilla/websocket"
)

// 区块间隔
// 区块瞬时间隔使用梯形折线图
// 区块平均间隔使用曲线折线图
// 这两者放在一张图Line中

// 交易吞吐量
// 瞬时吞吐量直接使用区块内交易数除以区块瞬时间隔，使用曲线图
// 平均吞吐量除以当前总区块内交易数除以总区块耗时，使用曲线图
// 放在一张Line上

// 交易确认时间
// 类似前面，也是一张Line图两条线

// 交易吞吐能力输入输出比
// Gauge

const (
	DefaultShowBlockNum = 20 // 只取最近的一百个区块统计数据
	DefaultDataChanSize = 5   // 因为数据产生速度很慢，所以其实这个chan 无size都可以，这里还是写个5，就是玩

	// writeWait is the time allowed to write the file to the client.
	writeWait = 10 * time.Second
	// pongWait is the time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second
	// pingPeriod is the interval between pings sent to client. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// MeasureData 表征一轮新计算出来的各项测量数据
type MeasureData struct {
	BlockTime             int64  `json:"block_time"`              // 区块时间，横轴数据
	InstantBlockDuration  int     `json:"instant_block_duration"`  // s
	AverageBlockDuration  int     `json:"average_block_duration"`  // s
	InstantTxThroughput   int     `json:"instant_tx_throughput"`   // 个
	AverageTxThroughput   int     `json:"average_tx_throughput"`   // 个
	InstantTxConfirmation int     `json:"instant_tx_confirmation"` // ms
	AverageTxConfirmation int     `json:"average_tx_confirmation"` // ms
	TxOutInRatio          float64 `json:"tx_out_in_ratio"`         // 0.x
}

// JSON json序列化
func (md MeasureData) JSON() []byte {
	data, err := json.Marshal(md)
	if err != nil {
		return nil
	}
	return data
}

// DefaultEchartsPageRunner 默认的EchartsPageRunner
var DefaultEchartsPageRunner = &EchartsPageRunner{}
var Run = DefaultEchartsPageRunner.Run
// func Run(host string) {
// 	DefaultEchartsPageRunner.Run(host)
// }

// EchartsPageRunner 利用ws更新EchartsPage
// 监控数据流向：
// 各个P2P节点 --TCP--> EchartsMonitor (TCP Server) --对区块/交易进行分析--> 得到MeasureData
// --channel--> EchartsPageRunner(HTTP Server) --WebSocket--> Client(Browser)
type EchartsPageRunner struct {
	page *EchartsPage
	Host string // HTTP/Websocket主机地址

	mdChan <-chan MeasureData // 数据chan。 由外部（EchartsMonitor）传入
	
	dispatches map[net.Addr]chan<- []byte	// <remoteAddr, dataC> 分发给各个连接
	dispatchesLock sync.RWMutex	
}

// Run 阻塞式运行
func (epr *EchartsPageRunner) Run(host string, mdChan <-chan MeasureData) {
	epr.Host = host
	epr.mdChan = mdChan
	epr.dispatches = make(map[net.Addr]chan<- []byte)

	// 构建EchartsPage页面
	page := NewEchartsPage(epr.Host)
	epr.page = page

	// 启动数据分发协程
	go epr.dispatchLoop()

	// 启动HTTP(Websocket)服务器
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { epr.page.Render(w) }) // 浏览器访问 host/ 即可返回初次渲染的页面
	http.HandleFunc("/ws", epr.handler())                                                     // 用于与客户端进行websocket通信
	log.Printf("Visit http://%s to see the live chart\n", epr.Host)
	if err := http.ListenAndServe(epr.Host, nil); err != nil {
		log.Fatal(err)
	}
}

func (epr *EchartsPageRunner) handler() http.HandlerFunc {
	if !epr.ok() {return nil}

	var dataC = make(chan []byte)
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	return func(w http.ResponseWriter, r *http.Request) {
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			if _, ok := err.(websocket.HandshakeError); !ok {
				log.Println(err)
			}
			return
		}

		// 将dataC注册到epr中
		epr.registerDispatch(ws, dataC)

		log.Printf("websocket conn with remote(%s) established...\n", ws.RemoteAddr().String())

		go epr.wsWriteLoop(ws, dataC) 
		epr.wsReadLoop(ws)	// 阻塞
	}
}

func (epr *EchartsPageRunner) registerDispatch(ws *websocket.Conn, dataC chan<- []byte) {
	if !epr.ok() {return}
	epr.dispatchesLock.Lock()
	epr.dispatches[ws.RemoteAddr()] = dataC
	epr.dispatchesLock.Unlock()
}

func (epr *EchartsPageRunner) unRegisterDispatch(ws *websocket.Conn) {
	if !epr.ok() {return}
	epr.dispatchesLock.Lock()
	delete(epr.dispatches, ws.RemoteAddr())
	epr.dispatchesLock.Unlock()
}

func (epr *EchartsPageRunner) wsWriteLoop(ws *websocket.Conn, dataC <-chan []byte) {
	var (
		err error
		data []byte
		pingTicker = time.NewTicker(pingPeriod)
	) 
	defer func() {
		pingTicker.Stop()
	}()
	for {
		select {
		case data = <-dataC:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if data == nil {
				err = errors.New("data == nil")
				goto ERR
			}
			if err = ws.WriteMessage(websocket.TextMessage, data); err != nil {
				goto ERR
			}
			log.Printf("websocket conn with remote(%s) writeLoop send: %s\n", ws.RemoteAddr().String(), string(data))
		case <-pingTicker.C:	// 隔一段时间ping一下客户端，看还在不，不在的话，考虑把对应的dataC移除
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err = ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				epr.unRegisterDispatch(ws)// 移除自己的dataC
				goto ERR
			}
			log.Printf("websocket conn with remote(%s) writeLoop send: PING\n", ws.RemoteAddr().String())
		}
	}
ERR:
	log.Printf("websocket conn with remote(%s) writeLoop return. err=%s\n", ws.RemoteAddr().String(), err)
	return
}

func (epr *EchartsPageRunner) wsReadLoop(ws *websocket.Conn) {
	ws.SetReadLimit(512)
	ws.SetReadDeadline(time.Now().Add(pongWait))
	ws.SetPongHandler(func(string) error { ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("websocket conn with remote(%s) readLoop recv: PONG\n", ws.RemoteAddr().String())
	}
}

// 数据分发循环
func (epr *EchartsPageRunner) dispatchLoop() {
	if !epr.ok() {return}

	for mdata := range epr.mdChan {
		chs := make([]chan<- []byte, 0)

		// 将dispatched表中的ch弄到一个数组中
		epr.dispatchesLock.RLock()
		for _, ch := range epr.dispatches {
			chs = append(chs, ch)
		}
		epr.dispatchesLock.RUnlock()

		// 分发
		for _, ch := range chs {
			ch <- mdata.JSON()
		}
	}
}

// 检查内部成员是否可用。部分API需要调用ok()检查
func (epr *EchartsPageRunner) ok() bool {
	dispatchesOk := false
	epr.dispatchesLock.RLock()
	if epr.dispatches != nil {
		dispatchesOk = true
	}
	epr.dispatchesLock.RUnlock() 

	return dispatchesOk && epr.page != nil && epr.mdChan != nil
}

// EchartsPage 把要用到的echarts封装在一起
type EchartsPage struct {
	page              *components.Page
	host              string
	dataChan          chan MeasureData

	// page中所含的chart
	// 如果要修改page中所含的chart，一定也要跟着修改ScriptFmt以及genScript()的方法
	blockDurationDL   *DoubleLine
	txThroughputDL    *DoubleLine
	txConfirmationDL  *DoubleLine
	txOutInRatioGauge *Gauge	
}

func NewEchartsPage(wsHost string) *EchartsPage {
	blockDuration := NewDoubleLine("区块间隔", "瞬时区块间隔", "平均区块间隔", "时间", "间隔/s")
	txThroughput := NewDoubleLine("交易吞吐量", "瞬时交易吞吐量", "平均交易吞吐量", "时间", "数量/笔")
	txConfirmation := NewDoubleLine("交易确认时间", "瞬时交易确认时间", "平均交易确认时间", "时间", "确认时间/s")
	txRatio := NewGauge("交易 输出量/输入量 —— 表征系统交易处理能力是否达到上限（降至100%以下）", "输出/输入", "交易O/I")

	page := components.NewPage()
	page.SetLayout(components.PageFlexLayout)
	page.AddCharts(blockDuration.line, txThroughput.line, txConfirmation.line, txRatio.gauge)
	page.PageTitle = "Eiger's Monitor"
	
	page.Validate()

	return &EchartsPage{
		page:              page,
		blockDurationDL:   blockDuration,
		txThroughputDL:    txThroughput,
		txConfirmationDL:  txConfirmation,
		txOutInRatioGauge: txRatio,
		host:              wsHost,
		dataChan:          make(chan MeasureData, DefaultDataChanSize),
	}
}

func (ep *EchartsPage) Render(w ...io.Writer) error {
	var buf bytes.Buffer
	var err error
	if err = ep.page.Render(&buf); err != nil {
		return fmt.Errorf("while pre-rendering: %w", err)
	}
	script := ep.genScript()
	for _, writer := range w {
		_, err = writer.Write(bytes.Replace(buf.Bytes(), []byte("</body>"), []byte(script), -1))
		if err != nil {
			return err
		}
	}

	return nil
}

// ScriptFmt is the template used for rendering ws-enabled charts
var ScriptFmt = `
<script type="text/javascript">
	// 为chart对象起别名（尽管有，但是因为要针对不同chart更新不同值，必须我们给它硬编码进去）
	var blockDurationDL = goecharts_%s;
	var txThroughputDL = goecharts_%s;
	var txConfirmationDL = goecharts_%s;
	var txOutInRatioGauge = goecharts_%s;
	// 为opt对象起别名（引用）
	var blockDurationDLOpt = option_%s;
	var txThroughputDLOpt = option_%s;
	var txConfirmationDLOpt = option_%s;
	var txOutInRatioGaugeOpt = option_%s;

	// 建立websocket连接
    let conn = new WebSocket("ws://%s/ws");
	// 连接关闭时的动作
    conn.onclose = function(evt) {
        console.log("connection closed");
    }
	// 通过连接收到消息时的动作，evt.data就是收到的数据（evt的类型是MessageEvent，是默认的参数）
    conn.onmessage = function(evt) {
        let data = JSON.parse(evt.data);	// 将json内容转为json对象

		// 更新charts的选项 data.xxx名称要和go代码里的tag对应起来
		updateDoubleLine(blockDurationDLOpt, data.block_time, data.instant_block_duration, data.average_block_duration);
		updateDoubleLine(txThroughputDLOpt, data.block_time, data.instant_tx_throughput, data.average_tx_throughput);
		updateDoubleLine(txConfirmationDLOpt, data.block_time, data.instant_tx_confirmation, data.average_tx_confirmation);
		updateGauge(txOutInRatioGaugeOpt, data.tx_out_in_ratio);	
		// 重新渲染charts
        blockDurationDL.setOption(blockDurationDLOpt);
		txThroughputDL.setOption(txThroughputDLOpt);
		txConfirmationDL.setOption(txConfirmationDLOpt);
		txOutInRatioGauge.setOption(txOutInRatioGaugeOpt);

		// 打印接受的数据
        console.log('Received data:', data);
    }
	
	function updateDoubleLine(lineOpt, xAxisValue, series1Value, series2Value) {
		// 更新x轴数据
		var xAxis = lineOpt.xAxis[0].data;
		//console.log('lineOpt:', lineOpt);
		//console.log('XAxis:', xAxis);
		xAxis.shift(); // 头部移除一个元素并返回
		xAxis.push(xAxisValue);

		// 更新series1
		var series1 = lineOpt.series[0].data;
		var item1 = series1.shift();
		item1.value = series1Value;
		series1.push(item1);

		// 更新series2
		var series2 = lineOpt.series[1].data;
		var item2 = series2.shift();
		item2.value = series2Value;
		series2.push(item2);
	}
	function updateGauge(gaugeOpt, value) {
		gaugeOpt.series[0].data[0].value = value;
	}
</script>

</body>
`

func (ep *EchartsPage) genScript() string {
	script := fmt.Sprintf(ScriptFmt, 
		ep.blockDurationDL.line.ChartID, 	// 
		ep.txThroughputDL.line.ChartID,
		ep.txConfirmationDL.line.ChartID,
		ep.txOutInRatioGauge.gauge.ChartID,
		ep.blockDurationDL.line.ChartID, 	// 
		ep.txThroughputDL.line.ChartID,
		ep.txConfirmationDL.line.ChartID,
		ep.txOutInRatioGauge.gauge.ChartID,
		ep.host)
	return script
}

type DoubleLine struct {
	line             *charts.Line
	// xAixs_blockTime  []string // 横轴坐标（区块时间）
	// series1_instance []int    // 线条1数据
	// series2_average  []int    // 线条2数据
}

func NewDoubleLine(title, series1Name, series2Name, xAxisName, yAxisName string) *DoubleLine {
	xAixs := make([]string, DefaultShowBlockNum)
	series1 := make([]int, DefaultShowBlockNum)
	series2 := make([]int, DefaultShowBlockNum)

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeMacarons,
			PageTitle: "Eiger's Monitor",
			}), // 设置主题
		charts.WithTitleOpts(opts.Title{
			Title: title,
			}),  // 设置标题
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true, 
			Orient: "horizontal", 
			Left:"right", 
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{Show: true},
				DataZoom: &opts.ToolBoxFeatureDataZoom{Show: true},
				DataView: &opts.ToolBoxFeatureDataView{Show: true},
				Restore: &opts.ToolBoxFeatureRestore{Show: true},}}),	// 使能工具箱
		charts.WithTooltipOpts(opts.Tooltip{
			Show: true,
			Trigger: "axis",
			TriggerOn: "mousemove|click",
		}),	// 鼠标移动或点击时的提示框
		charts.WithLegendOpts(opts.Legend{
			Data: []string{series1Name, series2Name},
			Top: "midlle",
			Orient: "horizontal",
		}),	// 图例
		charts.WithDataZoomOpts(opts.DataZoom{

		}),	// 区域缩放
		charts.WithXAxisOpts(opts.XAxis{
			Name: xAxisName, 
			Show: true,
			Type: "category", // 注意这里不能使用time或value之类的，因为我们想要区块时间而非实际数据绘图时时间，所以是离散数据
			GridIndex: 0,
			SplitArea: &opts.SplitArea{Show: true},
			AxisLabel: &opts.AxisLabel{
				Show: true,
				Interval: "0",	// 强制显示所有标签
				Rotate: 40,		// 逆时针旋转角度
				FontSize: "9",
			},	// 设置x轴标签展示问题（斜放展示）
			}),	// 设置x轴	
		charts.WithYAxisOpts(opts.YAxis{
			Name: yAxisName,
			Show: true,
		}), // y轴
			
		)                             
	line.SetXAxis(xAixs). // 设置横轴数据
				AddSeries(series1Name, generateLineItems(series1)). // 设置折线数据1
				AddSeries(series2Name, generateLineItems(series2)).
				SetSeriesOptions(
					charts.WithLineChartOpts(opts.LineChart{
						Smooth: false,
						}),// 设置线条属性（平滑）
					charts.WithLabelOpts(opts.Label{
						Show: true,
						Position: "top",
						}),	// 设置标签
					charts.WithRippleEffectOpts(opts.RippleEffect{}),	
					
						) 					
	line.Validate()			
	return &DoubleLine{
		line:             line,
		// xAixs_blockTime:  xAixs,
		// series1_instance: series1,
		// series2_average:  series2,
	}
}

// generateLineItems 生成Line上数据
func generateLineItems(values []int) []opts.LineData {
	items := make([]opts.LineData, len(values))
	for i := 0; i < len(values); i++ {
		items[i] = opts.LineData{Value: values[i]}
	}
	return items
}

type Gauge struct {
	gauge   *charts.Gauge
	// series1 int
}

func NewGauge(title, seriesName, dataName string) *Gauge {
	gauge := charts.NewGauge()
	gauge.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeMacarons,
			}), // 设置主题
		charts.WithTitleOpts(opts.Title{
			Title: title,
			}),  // 设置标题
		charts.WithToolboxOpts(opts.Toolbox{
			Show: true, 
			Orient: "horizontal", 
			Left:"right", 
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{Show: true},
				DataZoom: &opts.ToolBoxFeatureDataZoom{Show: true},
				DataView: &opts.ToolBoxFeatureDataView{Show: true},
				Restore: &opts.ToolBoxFeatureRestore{Show: true},}}),	// 使能工具箱
		charts.WithTooltipOpts(opts.Tooltip{
			Show: true,
			Trigger: "axis",
			TriggerOn: "mousemove|click",
		}),	// 鼠标移动或点击时的提示框
		charts.WithLegendOpts(opts.Legend{
			Left: "right",
		}),	// 图例
	
	)
	gauge.AddSeries(seriesName, []opts.GaugeData{{Name: dataName, Value: 0}})
	gauge.Validate()
	return &Gauge{
		gauge:   gauge,
		//series1: 0, // 初始值
	}
}
