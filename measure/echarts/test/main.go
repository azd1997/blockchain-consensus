// 测试echarts包
package main

import (
	"math/rand"
	"time"

	"github.com/azd1997/blockchain-consensus/measure/echarts"
)

func main() {
	host := "localhost:9999"
	mdChan := make(chan echarts.MeasureData, echarts.DefaultDataChanSize)

	// 定时随机生成数据
	tick := time.Tick(3 * time.Second)
	go func ()  {
		for {
			select {
			case <-tick:
				mdChan <- echarts.MeasureData{
					BlockTime: time.Now().UnixNano()/1e6,
					InstantBlockDuration: rand.Intn(10),
					AverageBlockDuration: rand.Intn(10),
					InstantTxThroughput: rand.Intn(1000),
					AverageTxThroughput: rand.Intn(1000),
					InstantTxConfirmation: rand.Intn(500),
					AverageTxConfirmation: rand.Intn(500),
					TxOutInRatio: float64(rand.Intn(100)),
				}
			}
		}
	}()

	echarts.Run(host, mdChan)
}