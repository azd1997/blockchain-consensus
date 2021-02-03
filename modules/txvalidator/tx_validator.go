package txvalidator

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/pitable"
	"github.com/azd1997/blockchain-consensus/modules/txvalidator/simpletv"
	"github.com/azd1997/blockchain-consensus/requires"
)

// TxValidator 交易验证器接口
// 该验证器指的是验证仅靠Ledger模块无法独立验证的内容。例如:
// 检查交易时间应早于区块时间，这不在TxValidator范围内；
// 检查交易数额是否满足余额条件或者一些自定义的检查条件，这个归TxValidator。
// 对于Blockchain它不需要理解验证器内逻辑，这只是业务侧的事情
// tv := NewTxValidator().SetLedger(ledger).SetPit(pit)
// if !tv.Ok() {return}
type TxValidator interface {

	/*配置依赖项*/

	// SetLedger 传入prepared ledger
	SetLedger(bc requires.BlockChain) TxValidator
	// SetPit 传入prepared pit
	SetPit(p pitable.Pit) TxValidator

	/*工具*/

	// Ok 检查Validator是否准备好
	Ok() bool

	/*职能*/

	// Validate 验证tx
	Validate(tx *defines.Transaction) error
}

func New(typ string, bc requires.BlockChain, pit pitable.Pit) TxValidator {
	switch typ {
	case "simpletv":
		return simpletv.New(bc, pit)
	default:
		return nil
	}
}
