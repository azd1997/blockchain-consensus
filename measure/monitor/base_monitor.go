package monitor

import (
	"fmt"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/measure/calculator"
	"github.com/azd1997/blockchain-consensus/measure/common"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
)

// BaseMonitor BaseMonitor只负责上报数据的收集与计算
//（逻辑上等价于 Collector -> Calculator），并将数据结果通过chan传出
// 不可独立运行使用
type BaseMonitor struct {
	net bnet.BNet
	msgChan chan *defines.Message
	mdChan chan common.MeasureData
}

func NewBaseMonitor(net bnet.BNet, msgChan chan *defines.Message, mdChan chan common.MeasureData) *BaseMonitor {
	if net == nil || !net.Inited() || msgChan == nil || mdChan == nil {return nil}
	return &BaseMonitor{
		net: net,
		msgChan: msgChan,
		mdChan: mdChan,
	}
}

func (m *BaseMonitor) Run() {
	if m.net == nil || !m.net.Inited() || m.msgChan == nil || m.mdChan == nil {
		return
	}

	// 使能calculator
	calculator.SetMDChan(m.mdChan)

	// 启动msgChan消费循环
	for msg := range m.msgChan {

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

}
