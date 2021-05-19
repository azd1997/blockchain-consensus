/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 9:53 PM
* @Description: The file is for
***********************************************************************/

package ledger

import (
	"errors"

	"github.com/azd1997/blockchain-consensus/modules/ledger/simplechain"
	"github.com/azd1997/blockchain-consensus/requires"
)

type LedgerType uint8

const (
	LedgerType_SimpleChain LedgerType = iota
)

// New
func New(ledgertype LedgerType, id string) (requires.BlockChain, error) {
	switch ledgertype {
	case LedgerType_SimpleChain:
		return simplechain.New(id)
	default:
		return nil, errors.New("unsupport ledger type")
	}
}

// Ledger 账本。比如区块链账本
type Ledger interface {
	requires.BlockChain
}
