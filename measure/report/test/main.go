// Package main 配合monitor/main 进行测试
package main

import (
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/modules/bnet/cs_net"
	"math/rand"
	"sync"
	"time"
)

func main() {
	serverId := "Monitor"
	serverListenAddr := "127.0.0.1:9998"
	clientId := "Client"
	clientListenAddr := "127.0.0.1:9997"	// 不需要
	msgout := make(chan *defines.Message)	// 不会使用
	client, err := cs_net.Dial(serverId, serverListenAddr, clientId, clientListenAddr, msgout)
	if err != nil {
		panic(err)
	}

	txs := make([]*defines.Transaction, 0)
	txsLock := sync.Mutex{}
	// 伪造消息发送
	blockTick := time.Tick(10 * time.Second)
	txTick := time.Tick(1 * time.Second)
	index := int64(1)
	nodes := []string{"node1", "node2", "node3", "node4", "node5"}
	lastBlock := &defines.Block{}
	for {
		select {
		case t := <-blockTick:
			// 随机延迟一点时间
			time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
			// 构建区块
			var atxs []*defines.Transaction
			txsLock.Lock()
			atxs  = append([]*defines.Transaction{}, txs...)
			txs = txs[:0]	// 清空txs
			txsLock.Unlock()
			b, err := defines.NewBlockAndSign(
				index, nodes[rand.Intn(len(nodes))],
				lastBlock.SelfHash, atxs, fmt.Sprintf("%s this is a block", time.Now().String()))
			if err != nil {
				panic(err)
			}

			// 构建区块消息
			msg, err := defines.NewMessage_NewBlock("", "", 0, b)
			if err != nil {
				panic(err)
			}
			// 发送区块消息
			err = client.Send(msg)
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s send block(%s)\n", t, b.Key())
			// 索引自增
			index++
		case t := <- txTick:
			// 随机等待一小段时间
			time.Sleep(time.Duration(rand.Intn(20)) * time.Millisecond)
			// 随机数量
			num := rand.Intn(3) + 1
			// 构建交易
			atxs := make([]*defines.Transaction, num)
			for i:=0;i<num;i++ {
				tx, err := defines.NewTransactionAndSign(
					"xxx", "yyy", int64(rand.Intn(1000)), nil,
					fmt.Sprintf("%s this is a tx", time.Now().String()))
				if err != nil {
					panic(err)
				}
				atxs[i] = tx
			}
			// 存于txs
			txsLock.Lock()
			txs = append(txs, atxs...)
			txsLock.Unlock()
			// 构建交易消息
			msg, err := defines.NewMessage_Txs("", "", 0, atxs)
			if err != nil {
				panic(err)
			}
			// 发送交易消息
			err = client.Send(msg)
			if err != nil {
				panic(err)
			}

			l := 0
			txsLock.Lock()
			l = len(txs)
			txsLock.Unlock()
			fmt.Printf("%s send %d tx, now len(txs)=%d\n", t, num, l)
		}
	}
}