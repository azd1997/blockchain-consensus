// 测试echarts包
package main

import (
	"math/rand"
	"time"

	"github.com/azd1997/blockchain-consensus/measure/common"
	"github.com/azd1997/blockchain-consensus/measure/echarts_old"
)

func main() {
	host := "localhost:9999"
	mdChan := make(chan common.MeasureData, common.DefaultDataChanSize)

	// 定时随机生成数据
	tick := time.Tick(3 * time.Second)
	go func ()  {
		for {
			select {
			case <-tick:
				mdChan <- common.MeasureData{
					BlockTime: time.Now().UnixNano()/1e6,
					CurBlockDuration: int64(rand.Intn(10)),
					RangeAverageBlockDuration: int64(rand.Intn(10)),
					CurAverageTxThroughput: float64(rand.Intn(1000)),
					RangeAverageTxThroughput: float64(rand.Intn(1000)),
					CurAverageTxConfirmation: int64(rand.Intn(500)),
					RangeAverageTxConfirmation: int64(rand.Intn(500)),
					CurTxOutInRatio: float64(rand.Intn(100)),
				}
			}
		}
	}()

	echarts_old.Run(host, mdChan)
}