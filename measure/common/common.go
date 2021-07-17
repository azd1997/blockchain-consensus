package common

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

const (
	DefaultShowBlockNum  = 15  // 给echarts用的，只画出15个区块的数据
	DefaultRangeBlockNum = 100 // 给calculator用的，对100个区块取平均值
	DefaultDataChanSize  = 5   // 因为数据产生速度很慢，所以其实这个chan 无size都可以，这里还是写个5，就是玩
)

// MeasureData 表征一轮新计算出来的各项测量数据
// 这里所有时间单位都是ns time.UnixNano
type MeasureData struct {
	BlockKey      string `json:"block_key" csv:"block_key"`
	BlockTime     int64  `json:"block_time" csv:"block_time"`           // 区块时间，横轴数据
	RangeBlockNum int64  `json:"range_block_num" csv:"range_block_num"` // 一段范围内的区块总数
	// 总数据
	TotalBlockNum int64 `json:"total_block_num" csv:"total_block_num"`   // 总区块数
	TotalTxGenNum int64 `json:"total_tx_gen_num" csv:"total_tx_gen_num"` // 总共生成的交易数
	TotalTxOutNum int64 `json:"total_tx_out_num" csv:"total_tx_out_num"` // 总共在区块内
	// 本轮区块的数据
	CurTxOutNum int64 `json:"cur_tx_out_num" csv:"cur_tx_out_num"`
	CurTxGenNum int64 `json:"cur_tx_gen_num" csv:"cur_tx_gen_num"`
	// 一段范围内的数据（common.DefaultShowBlockNum）
	RangeTxOutNum      int64 `json:"range_tx_out_num" csv:"range_tx_out_num"`
	RangeTxGenNum      int64 `json:"range_tx_gen_num" csv:"range_tx_gen_num"`
	RangeBlockDuration int64 `json:"range_block_duration" csv:"range_block_duration"`

	CurBlockDuration           int64   `json:"cur_block_duration" csv:"cur_block_duration"`                       // ns
	RangeAverageBlockDuration  int64   `json:"range_average_block_duration" csv:"range_average_block_duration"`   // ns
	CurAverageTxThroughput     float64 `json:"cur_average_tx_throughput" csv:"cur_average_tx_throughput"`         // 个/ns
	RangeAverageTxThroughput   float64 `json:"range_average_tx_throughput" csv:"range_average_tx_throughput"`     // 个/ns
	CurAverageTxConfirmation   int64   `json:"cur_average_tx_confirmation" csv:"cur_average_tx_confirmation"`     // ns
	RangeAverageTxConfirmation int64   `json:"range_average_tx_confirmation" csv:"range_average_tx_confirmation"` // ns
	CurTxOutInRatio            float64 `json:"cur_tx_out_in_ratio" csv:"cur_tx_out_in_ratio"`                     // 0.x
	RangeTxOutInRatio          float64 `json:"range_tx_out_in_ratio" csv:"range_tx_out_in_ratio"`                 // 0.x
}

// JSON json序列化
func (md MeasureData) JSON() []byte {
	data, err := json.Marshal(md)
	if err != nil {
		return nil
	}
	return data
}

// CSV csv序列化
func (md MeasureData) CSV() []string {

	sv := reflect.ValueOf(md)
	numField := sv.NumField()

	row := make([]string, numField)

	for i := 0; i < numField; i++ {
		field := sv.Field(i)
		row[i] = fmt.Sprintf("%v", field.Interface()) // 直接使用field.String()得到的字符串格式不是我想要的
	}

	return row
}

// CSVHeader csv序列化的表头
func (md MeasureData) CSVHeader() []string {

	st := reflect.TypeOf(md)
	numField := st.NumField()

	row := make([]string, numField)

	for i := 0; i < numField; i++ {
		field := st.Field(i)
		csv, ok := field.Tag.Lookup("csv")
		if ok && csv != "" {
			row[i] = csv
		}
	}

	return row
}

// GenerateRandomMD 生成随机MD
func GenerateRandomMD() MeasureData {
	now := time.Now()
	return MeasureData{
		BlockKey:                   fmt.Sprintf("blockkey_%d", now.UnixNano()),
		BlockTime:                  now.UnixNano(),
		RangeBlockNum:              int64(rand.Intn(20)),
		TotalBlockNum:              int64(rand.Intn(100)),
		TotalTxGenNum:              int64(rand.Intn(10000)),
		TotalTxOutNum:              int64(rand.Intn(10000)),
		CurTxOutNum:                int64(rand.Intn(200)),
		CurTxGenNum:                int64(rand.Intn(200)),
		RangeTxOutNum:              int64(rand.Intn(2000)),
		RangeTxGenNum:              int64(rand.Intn(2000)),
		RangeBlockDuration:         int64(rand.Intn(20)),
		CurBlockDuration:           int64(rand.Intn(10)),
		RangeAverageBlockDuration:  int64(rand.Intn(20)),
		CurAverageTxThroughput:     float64(rand.Intn(500)),
		RangeAverageTxThroughput:   float64(rand.Intn(500)),
		CurAverageTxConfirmation:   int64(rand.Intn(80000)),
		RangeAverageTxConfirmation: int64(rand.Intn(80000)),
		CurTxOutInRatio:            float64(rand.Intn(100)) / 100,
		RangeTxOutInRatio:          float64(rand.Intn(100)) / 100,
	}
}
