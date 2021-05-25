package common

import "encoding/json"

const (
	DefaultShowBlockNum = 15 // 给echarts用的，只画出15个区块的数据
	DefaultRangeBlockNum = 100	// 给calculator用的，对100个区块取平均值
	DefaultDataChanSize = 5  // 因为数据产生速度很慢，所以其实这个chan 无size都可以，这里还是写个5，就是玩
)

// MeasureData 表征一轮新计算出来的各项测量数据
// 这里所有时间单位都是ns time.UnixNano
type MeasureData struct {
	BlockKey      string `json:"block_key"`
	BlockTime     int64  `json:"block_time"`      // 区块时间，横轴数据
	RangeBlockNum int64  `json:"range_block_num"` // 一段范围内的区块总数
	// 总数据
	TotalBlockNum int64 `json:"total_block_num"`  // 总区块数
	TotalTxGenNum int64 `json:"total_tx_gen_num"` // 总共生成的交易数
	TotalTxOutNum int64 `json:"total_tx_out_num"` // 总共在区块内
	// 本轮区块的数据
	CurTxOutNum int64 `json:"cur_tx_out_num"`
	CurTxGenNum int64 `json:"cur_tx_gen_num"`
	// 一段范围内的数据（common.DefaultShowBlockNum）
	RangeTxOutNum      int64 `json:"range_tx_out_num"`
	RangeTxGenNum      int64 `json:"range_tx_gen_num"`
	RangeBlockDuration int64 `json:"range_block_duration"`

	CurBlockDuration           int64   `json:"cur_block_duration"`            // ns
	RangeAverageBlockDuration  int64   `json:"range_average_block_duration"`  // ns
	CurAverageTxThroughput     float64 `json:"cur_average_tx_throughput"`     // 个/ns
	RangeAverageTxThroughput   float64 `json:"range_average_tx_throughput"`   // 个/ns
	CurAverageTxConfirmation   int64   `json:"cur_average_tx_confirmation"`   // ns
	RangeAverageTxConfirmation int64   `json:"range_average_tx_confirmation"` // ns
	CurTxOutInRatio            float64 `json:"cur_tx_out_in_ratio"`           // 0.x
	RangeTxOutInRatio          float64 `json:"range_tx_out_in_ratio"`         // 0.x
}

// JSON json序列化
func (md MeasureData) JSON() []byte {
	data, err := json.Marshal(md)
	if err != nil {
		return nil
	}
	return data
}
