/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/15/20 9:27 AM
* @Description: 交易制造器
***********************************************************************/

package test

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
)

type TxMaker struct {
	id    string                    // 自己
	tos   []string                  // 可能的交易接收方
	txout chan *defines.Transaction // 传给pot状态机
}

func NewTxMaker(id string, tos []string, txout chan *defines.Transaction) *TxMaker {
	return &TxMaker{
		id:    id,
		tos:   tos,
		txout: txout,
	}
}

// 隔一段时间随机向某个to构造交易
// go tm.Start
func (tm *TxMaker) Start() {
	period := 5 * time.Millisecond
	tick := time.Tick(period)
	for {
		select {
		case <-tick:
			to := tm.tos[rand.Intn(len(tm.tos))]
			amount := rand.Intn(100)
			description := fmt.Sprintf("this is a tx from %s to %s", tm.id, to)
			tx, err := defines.NewTransactionAndSign(tm.id, to, int64(amount), nil, description)
			if err != nil {
				fmt.Printf("TxMaker make tx fail: %s\n", err)
			}

			// 随机拖延一段时间
			time.Sleep(time.Duration(rand.Intn(3)) * time.Millisecond)

			fmt.Printf("TxMaker make tx succ. from: %s, to: %s\n", tm.id, to)
			tm.txout <- tx
		}
	}
}
