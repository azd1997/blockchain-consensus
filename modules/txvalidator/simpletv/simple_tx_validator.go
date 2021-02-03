package simpletv

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/pitable"
	"github.com/azd1997/blockchain-consensus/modules/txvalidator"
	"github.com/azd1997/blockchain-consensus/requires"
)

// SimpleTxValidator 简单的交易验证器
type SimpleTxValidator struct {
	bc  requires.BlockChain
	pit pitable.Pit
}

func New(bc requires.BlockChain, pit pitable.Pit) *SimpleTxValidator {
	return &SimpleTxValidator{
		bc:  bc,
		pit: pit,
	}
}

func (s *SimpleTxValidator) SetLedger(bc requires.BlockChain) txvalidator.TxValidator {
	s.bc = bc
	return s
}

func (s *SimpleTxValidator) SetPit(p pitable.Pit) txvalidator.TxValidator {
	s.pit = p
	return s
}

func (s *SimpleTxValidator) Ok() bool {
	return s.bc != nil && s.pit != nil
}

func (s *SimpleTxValidator) Validate(tx *defines.Transaction) error {
	// 检查tx的签名
	if err := tx.Verify(); err != nil {
		return err
	}

	// 检查其他条件，例如id是否有效合法等等，这里不做

	return nil
}
