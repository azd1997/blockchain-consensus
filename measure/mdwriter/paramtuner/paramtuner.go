// TODO: PARAMTUNER实际实现并不是mdwriter，需要做成mdwriter的是perf95
// ParamTuner这个功能以后需要额外增加，这里将这部分代码先保留以作参考
package paramtuner

import (
	"fmt"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/measure/common"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
)

// ParamTuner 参数整定器
// 参数整定器的作用是，根据输入的MD，调整测试参数
type ParamTuner struct {
	net bnet.BNet
	//count int64	// 记录当前收到的MD的个数
}

func NewParamTuner(net bnet.BNet) *ParamTuner {
	if net == nil {
		panic("net == nil")
	}
	return &ParamTuner{
		net: net,
	}
}

func (pt *ParamTuner) MDWrite(md common.MeasureData) error {
	if md.TotalBlockNum % 100 == 0 {	// 每100个md检查一次。这个数量要满足系统运行稳定，得到的参数反映系统的真是能力
		// 检查性能指标
		if md.RangeTxOutInRatio > 0.95 {
			// 触发参数调整（调高交易生成速度）(广播)
			pt.increaseTxGenSpeed()
		} else {
			// 发消息停止集群运行
			pt.stopCluster()
		}
	}
	return nil
}

// 步进式增大交易生成速度
// 这里是不会控制具体增大多少的，这取决于集群的txmaker自己的做法
func (pt *ParamTuner) increaseTxGenSpeed() {
	// 构建msg模板
	msg, err := defines.NewMessageAndSign_GenTxFaster(pt.net.ID(), "")
	if err != nil {
		fmt.Println("ParamTuner increaseTxGenSpeed NewMessageAndSign_GenTxFaster fail. err=", err)
		return
	}
	// 广播
	err = pt.net.Broadcast(msg)
	if err != nil {
		fmt.Println("ParamTuner increaseTxGenSpeed Broadcast fail. err=", err)
		return
	}
}

func (pt *ParamTuner) stopCluster() {
	// 构建msg模板
	msg, err := defines.NewMessageAndSign_StopNode(pt.net.ID(), "")
	if err != nil {
		fmt.Println("ParamTuner stopCluster NewMessageAndSign_StopNode fail. err=", err)
		return
	}
	// 广播
	err = pt.net.Broadcast(msg)
	if err != nil {
		fmt.Println("ParamTuner stopCluster Broadcast fail. err=", err)
		return
	}
}