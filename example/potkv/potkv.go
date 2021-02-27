package main

import (
	bcc "github.com/azd1997/blockchain-consensus"
	"github.com/azd1997/blockchain-consensus/modules/txvalidator"
)

// 最简单的情况下就是存字符串

// PotKV 基于PoT协议的kv存储
// 那么其实就是在Tx内包含对数据kv的增删改查操作
// 这里考虑所有节点区块链暂时维护在内存中（后续再考虑持久化），将
type PotKV struct {
	// p2p节点
	node *bcc.Node

	// 自定义的组件
	tv txvalidator.TxValidator

	// 自定义的TxInterpreter，其实就是该结构自身。（抽象化过于困难）
}

// NewPotKV 新建
func NewPotKV() (*PotKV, error) {

}

////////////////////// Usage /////////////////////

type Op string

const (
	OpSet Op = "SET"	// 添加或修改
	OpDel Op = "DEL" 	// 删除
)

// Command 用来描述对数据的更改
type Command struct {
	Op Op
	Key string
	Value string
}

// WriteCommand 添加键值对
func (pkv *PotKV) WriteCommand(c Command) error {
	pkv.node.
}



