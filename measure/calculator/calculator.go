package calculator

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/measure/common"
)


var defaultCalculator Calculator

func AddBlock(b *defines.Block) {
	if defaultCalculator == nil {return}
	defaultCalculator.AddBlock(b)
}

func AddTx(tx *defines.Transaction) {
	if defaultCalculator == nil {return}
	defaultCalculator.AddTx(tx)
}

func SetMDChan(mdChan chan<- common.MeasureData) {
	defaultCalculator = NewBasicCalculator(mdChan)
	defaultCalculator.SetMDChan(mdChan)
}

// Calculator 接口描述接收到区块/交易之后如何计算得到MeasureData
type Calculator interface {
	AddBlock(b *defines.Block)
	AddTx(tx *defines.Transaction)
	SetMDChan(mdChan chan<- common.MeasureData)
}

type BlockVote struct {
	B *defines.Block
	Votes int
}

type BlockInfo struct {
	B *defines.Block
	BlockDuration int64	// 时间戳与上一区块时间戳之差
	TxOutNum int64	// 区块内包含的总交易数
	TotalTxConfirmation int64	// 区块内所有交易的确认时间总和
	TxGenNum int64	// 在上一区块至该区块中间生成的交易总数
}
