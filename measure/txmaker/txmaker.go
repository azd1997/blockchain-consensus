package txmaker

import (
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/bnet"
	"math/rand"
	"time"
)

//type TxMaker struct {
//	net bnet.BNet
//	changeSpeed chan int	// +1 加快 -1 减慢
//}
//
//func NewTxMaker(net bnet.BNet, changeSpeed chan int) *TxMaker {
//
//}
//
//func (tm *TxMaker) Run() {
//
//}

var (
	genTxNumPerTime = 1
)

func Run(net bnet.BNet, changeSpeed chan int) {
	unit := time.Millisecond
	period := 50 * unit
	ticker := time.NewTicker(period)

	for {
		select {
		case <-ticker.C:	// 制造交易
			genTxRandomly(net)
		case cs := <- changeSpeed:	// 调整交易生成速度（调整ticker间隔没有调整单次生成交易数量有效，选择后者）
			if cs > 0 {	// 加快
				genTxNumPerTime++
			} else if cs < 0 {	// 减慢
				genTxNumPerTime--
			} else {	// cs == 0	// 停止
				genTxNumPerTime = 0
				return
			}
		}
	}
}

func genTxRandomly(net bnet.BNet) error {

	// 随机交易金额
	amount := rand.Intn(100)
	description := fmt.Sprintf("this is a tx from %s to %s", tm.id, to)
	tx, err := defines.NewTransactionAndSign(tm.id, to, int64(amount), nil, description)
	if err != nil {
		fmt.Printf("TxMaker(%s) make tx fail: %s\n", tm.id, err)
	}

	// 随机拖延一段时间
	time.Sleep(time.Duration(rand.Intn(3)) * time.Millisecond)

	txb, err := tx.Encode()
	if err != nil {
		fmt.Printf("TxMaker(%s) make tx fail: %s\n", tm.id, err)
	}
	fmt.Printf("TxMaker(%s) make tx succ. from: %s, to: %s\n", tm.id, tm.id, to)

	msg, err := defines.NewMessageAndSign_Txs(net.ID(), "", 0, txs)
	if err != nil {
		return err
	}
	net.Broadcast()
	tm.txout <-
}

func genRandomTxs(net bnet.BNet, txNum int) []*defines.Transaction {
	// 随机选择一部分对象
	tos := net.IDAddrs()
	tos = shuffle(tos)
	to := tos[rand.Intn(len(tos))]
}

func genRandomDesc() string {

}

func shuffle(strs [][2]string) [][2]string {
	// 洗牌算法

	var index int
	l := len(strs)
	for i := 0; i < l; i++ {
		index = rand.Intn(l-i) + i
		strs[i], strs[index] = strs[index], strs[i]
	}

	return strs
}
