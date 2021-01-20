/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 10:04 PM
* @Description: The file is for
***********************************************************************/

package simplechain

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"testing"

	"github.com/azd1997/blockchain-consensus/log"
)

func TestBlockChain(t *testing.T) {
	id := "peerA"
	log.InitGlobalLogger(id, true, true)

	// bc创建启动
	bc, err := NewBlockChain(id)
	if err != nil {
		t.Error(err)
	}
	err = bc.Init()
	if err != nil {
		t.Error(err)
	}

	// b1
	genesis, err := bc.CreateTheWorld()
	if err != nil {
		t.Error(err)
	}
	t.Log("b1: ", genesis.ShortName())
	t.Log(bc.Display())

	// b2
	b2, err := defines.NewBlockAndSign(2, id, genesis.SelfHash, nil, "2th block")
	if err != nil {
		t.Error(err)
	}
	t.Log("b2: ", b2.ShortName())
	err = bc.AddNewBlock(b2)
	if err != nil {
		t.Error(err)
	}
	t.Log(bc.Display())
	// b1 -> b2

	t.Log("continuous? ", !bc.Discontinuous())

	// 现在连续创建3个区块
	b3, err := defines.NewBlockAndSign(3, id, b2.SelfHash, nil, "3th block")
	if err != nil {
		t.Error(err)
	}
	t.Log("b3: ", b3.ShortName())
	b4, err := defines.NewBlockAndSign(4, id, b3.SelfHash, nil, "4th block")
	if err != nil {
		t.Error(err)
	}
	t.Log("b4: ", b4.ShortName())
	b5, err := defines.NewBlockAndSign(5, id, b4.SelfHash, nil, "5th block")
	if err != nil {
		t.Error(err)
	}
	t.Log("b5: ", b5.ShortName())

	// 假设从网络中刚得到b5
	err = bc.AddNewBlock(b5)
	if err != nil {
		t.Error(err)
	}
	t.Log(bc.Display())
	// b1 -> b2
	// b5
	t.Log("continuous? ", !bc.Discontinuous())

	// 接着获取b3,b4
	err = bc.AddBlock(b3)
	if err != nil {
		t.Error(err)
	}
	t.Log(bc.Display())
	// b1 -> b2
	// b5
	// (b3)
	t.Log("continuous? ", !bc.Discontinuous())
	err = bc.AddBlock(b4)
	if err != nil {
		t.Error(err)
	}
	t.Log(bc.Display())
	// b1 -> b2
	// b4 -> b5
	// (b3)
	t.Log("continuous? ", !bc.Discontinuous())

	// 接着收到网络新的最新区块b6
	b6, err := defines.NewBlockAndSign(6, id, b5.SelfHash, nil, "6th block")
	if err != nil {
		t.Error(err)
	}
	t.Log("b6: ", b6.ShortName())
	err = bc.AddNewBlock(b6)
	if err != nil {
		t.Error(err)
	}
	t.Log(bc.Display())
	// b1 -> b2 -> b3 -> b4 -> b5 -> b6		(添加b6之后会检查离散的区块集中是否可填进空缺，发现b3可用)
	t.Log()
	t.Log("continuous? ", !bc.Discontinuous())
}
