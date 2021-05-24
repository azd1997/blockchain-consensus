package calculator_test

import (
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/measure/calculator"
	"github.com/azd1997/blockchain-consensus/measure/common"
	"sync"
	"testing"
)

// 测试是否能判断出正确的区块
func TestCalculator_DecideBlock(t *testing.T) {
	mdChan := make(chan common.MeasureData, 10)
	calculator.SetMDChan(mdChan)

	// 读mdChan
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		md := <- mdChan
		fmt.Println("md=", md)
		wg.Done()
	}()

	b1, err := defines.NewBlockAndSign(1, "aaa", nil, nil, "b1")
	if err != nil {
		panic(err)
	}
	b1_, err := defines.NewBlockAndSign(1, "bbb", nil, nil, "b1_")
	if err != nil {
		panic(err)
	}

	// node1,node2支持b1，node3支持b1_
	calculator.AddBlock(b1)
	calculator.AddBlock(b1)
	calculator.AddBlock(b1_)

	// 使用一个index=2的区块触发calculator进行计算
	b2, err := defines.NewBlockAndSign(2, "ccc", nil, nil, "b2")
	if err != nil {
		panic(err)
	}
	calculator.AddBlock(b2)

	// 这里的测试逻辑是，如果decideBlock正确
	// 那么下面会输出md并退出
	wg.Wait()
}

// 测试是否能正常计算 TODO: 测试内容
func TestCalculator_Calculate(t *testing.T) {
	mdChan := make(chan common.MeasureData, 10)
	calculator.SetMDChan(mdChan)

	// 读mdChan
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		md := <- mdChan
		fmt.Println("md=", md)
		wg.Done()
	}()

	b1, err := defines.NewBlockAndSign(1, "aaa", nil, nil, "b1")
	if err != nil {
		panic(err)
	}
	b1_, err := defines.NewBlockAndSign(1, "bbb", nil, nil, "b1_")
	if err != nil {
		panic(err)
	}

	// node1,node2支持b1，node3支持b1_
	calculator.AddBlock(b1)
	calculator.AddBlock(b1)
	calculator.AddBlock(b1_)

	// 使用一个index=2的区块触发calculator进行计算
	b2, err := defines.NewBlockAndSign(2, "ccc", nil, nil, "b2")
	if err != nil {
		panic(err)
	}
	calculator.AddBlock(b2)

	// 这里的测试逻辑是，如果decideBlock正确
	// 那么下面会输出md并退出
	wg.Wait()
}

