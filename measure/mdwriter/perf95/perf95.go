package perf95

import (
	"github.com/azd1997/blockchain-consensus/measure/common"
)

// TxMakerController TxMaker的控制器
// 根据输入的MD，调整TxMaker测试参数
type TxMakerController struct {
	changeSpeed chan int
}

func NewTxMakerController(changeSpeed chan int) *TxMakerController {
	if changeSpeed == nil {
		panic("changeSpeed == nil")
	}
	return &TxMakerController{
		changeSpeed: changeSpeed,
	}
}

func (pt *TxMakerController) MDWrite(md common.MeasureData) error {
	if md.TotalBlockNum % 100 == 0 {	// 每100个md检查一次。这个数量要满足系统运行稳定，得到的参数反映系统的真是能力
		// 检查性能指标
		if md.RangeTxOutInRatio > 0.95 {
			// 触发参数调整（调高交易生成速度）(广播)
			pt.changeSpeed <- 1
		} else {
			// 发消息停止集群运行
			pt.changeSpeed <- 0
		}
	}
	return nil
}