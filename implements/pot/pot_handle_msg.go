/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 5:16 PM
* @Description: Pot的handle相关方法
***********************************************************************/

package pot

import (
	"errors"

	"github.com/azd1997/blockchain-consensus/defines"
)

// 处理外界消息输入和内部消息
func (p *Pot) handleMsg(msg *defines.Message) error {
	// 检查消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	// 根据当前状态不同，执行不同的消息处理
	state := p.getState()
	switch state {
	case StateType_Init_GetNeighbors:
		return p.handleMsgWhenInitGetNeighbors(msg)
	case StateType_Init_GetProcesses:
		return p.handleMsgWhenInitGetProcesses(msg)
	case StateType_Init_GetBlocks:
		return p.handleMsgWhenInitGetBlocks(msg)
	case StateType_NotReady:
		return p.handleMsgWhenNotReady(msg)
	case StateType_ReadyCompete:
		return p.handleMsgWhenReadyCompete(msg)
	case StateType_Competing:
		return p.handleMsgWhenCompeting(msg)
	case StateType_CompeteOver:
		return p.handleMsgWhenCompeteOver(msg)
	case StateType_CompeteWinner:
		return p.handleMsgWhenCompeteWinner(msg)
	case StateType_CompeteLoser:
		return p.handleMsgWhenCompeteLoser(msg)
	}
	return nil
}

/////////////////////////// 处理消息 /////////////////////////

// 目前的消息只有类：区块消息、证明消息、
// 快照消息（用于NotReady的节点快速获得区块链部分，目前直接用区块消息代表）。

// 只处理Neighbors消息
func (p *Pot) handleMsgWhenInitGetNeighbors(msg *defines.Message) error {
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_Data:
		count := len(msg.Entries)
		for _, ent := range msg.Entries {
			switch ent.Type {
			case defines.EntryType_Neighbor:
				count--
				err := p.handleEntryNeighbor(ent)
				if err != nil {
					p.Errorf("%s handle EntryType_Neighbor fail: %s\n", p.getState().String(), err)
				}
			default: // 其他类型则忽略
				p.Errorf("%s can only handle EntryType_Neighbor\n", p.getState().String())
			}
		}
		if count == 0 {		// 说明全是Neighbors
			p.nWaitChan <- 1	// 通知收到一个节点回传了节点信息表
		}
	case defines.MessageType_Req:
		// 该阶段不能处理Req消息
		p.Errorf("%s can only handle EntryType_Neighbor\n", p.getState().String())
	default:
		p.Errorf("unknown msg type")
	}
	return nil
}

// 收集进度
func (p *Pot) handleMsgWhenInitGetProcesses(msg *defines.Message) error {
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_Data:
		for _, ent := range msg.Entries {
			switch ent.Type {
			case defines.EntryType_Neighbor:
				err := p.handleEntryNeighbor(ent)
				if err != nil {
					p.Errorf("%s handle EntryType_Neighbor fail: %s\n", p.getState().String(), err)
				}
			case defines.EntryType_Process:
				err := p.handleEntryProcess(ent)
				if err != nil {
					p.Errorf("%s handle EntryType_Progress fail: %s\n", p.getState().String(), err)
				}
			default: // 其他类型则忽略
				p.Errorf("%s can only handle EntryType_Neighbor or EntryType_Process\n", p.getState().String())
			}
		}
	case defines.MessageType_Req:
		// 该阶段不能处理Req消息
		p.Errorf("%s can only handle EntryType_Neighbor\n", p.getState().String())
	default:
		p.Errorf("unknown msg type")
	}
	return nil
}

// 收集区块追赶进度
func (p *Pot) handleMsgWhenInitGetBlocks(msg *defines.Message) error {
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_Data:
		for _, ent := range msg.Entries {
			switch ent.Type {
			case defines.EntryType_Neighbor:
				err := p.handleEntryNeighbor(ent)
				if err != nil {
					p.Errorf("%s handle EntryType_Neighbor fail: %s\n", p.getState().String(), err)
				}
			case defines.EntryType_Process:
				err := p.handleEntryProcess(ent)
				if err != nil {
					p.Errorf("%s handle EntryType_Progress fail: %s\n", p.getState().String(), err)
				}
			default: // 其他类型则忽略
				p.Errorf("%s can only handle EntryType_Neighbor or EntryType_Process\n", p.getState().String())
			}
		}
	case defines.MessageType_Req:
		// 该阶段不能处理Req消息
		p.Errorf("%s can only handle EntryType_Neighbor\n", p.getState().String())
	default:
		p.Errorf("unknown msg type")
	}
	return nil
}

func (p *Pot) handleMsgWhenNotReady(msg *defines.Message) error {
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_Data:
		// 该阶段只能处理block消息
		// 根据EntryType来处理
		for _, ent := range msg.Entries {
			switch ent.Type {
			case defines.EntryType_Block:
				return p.handleEntryBlock(ent)
			default: // 其他类型则忽略
			}
		}
	case defines.MessageType_Req:
		// 该阶段不能处理Req消息
	default:
		// 报错打日志
	}
	return nil
}

// ReadyCompete状态下可以收集交易、新区块，以及区块和邻居请求
func (p *Pot) handleMsgWhenReadyCompete(msg *defines.Message) error {
	// 检查自身是否追上最新进度
	if !p.latest() {
		return errors.New("not latest process")
	}
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_Data:
		// ReadyCompete状态下不允许接收区块和证明
		for i:=0; i<len(msg.Entries); i++ {
			ent := msg.Entries[i]
			switch ent.Type {
			case defines.EntryType_Block:
				return p.handleEntryBlock(ent)
			case defines.EntryType_Proof:
			case defines.EntryType_NewBlock:
				return p.handleEntryNewBlock(ent)
			case defines.EntryType_Transaction:
				return p.handleEntryTransaction(ent)
			default:
				return errors.New("unknown entry type")
			}
		}
	case defines.MessageType_Req:
		// 检查请求内容，查看
		if len(msg.Entries) > 0 || len(msg.Reqs) == 0 {
			return errors.New("not a req msg")
		}
		for i:=0; i<len(msg.Reqs); i++ {
			req := msg.Reqs[i]
			switch req.Type {
			case defines.RequestType_Blocks:
				return p.handleRequestBlocks(msg.From, req)
			case defines.RequestType_Neighbors:
				return p.handleRequestNeighbors(msg.From, req)
			default: // 其他类型则忽略
				return errors.New("unknown req type")
			}
		}
	default:
		return errors.New("unknown msg type")
	}

	return nil
}

// Competing状态下msg处理
func (p *Pot) handleMsgWhenCompeting(msg *defines.Message) error {
	// 检查自身是否追上最新进度
	if !p.latest() {
		return errors.New("not latest process")
	}
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_Data:
		// ReadyCompete状态下不允许接收区块，允许接收证明
		if len(msg.Entries) != 1 {
			return errors.New("recv proof only, but len(msg.Entries) != 1")
		}

		ent := msg.Entries[0]
		switch ent.Type {
		case defines.EntryType_Proof:
			return p.handleEntryProof(ent, msg.From)
		default: // 其他类型则忽略
		}

	case defines.MessageType_Req:
		// 检查请求内容，查看
		if len(msg.Entries) > 0 || len(msg.Reqs) == 0 {
			return errors.New("not a req msg")
		}
		// 根据RequestType来处理
		for _, req := range msg.Reqs {
			switch req.Type {
			case defines.RequestType_Blocks:
				return p.handleRequestBlocks(msg.From, req)
			case defines.RequestType_Neighbors:
				return p.handleRequestNeighbors(msg.From, req)
			default: // 其他类型则忽略
			}
		}

	default:
		return errors.New("unknown msg type")
	}
	return nil
}

// CompeteOver 状态
// 此状态极为短暂，该期间只能照常处理交易、邻居消息，其他消息不处理
func (p *Pot) handleMsgWhenCompeteOver(msg *defines.Message) error {
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_None:
		// 啥也不干
	case defines.MessageType_Data:
		for i:=0; i<len(msg.Entries); i++ {
			ent := msg.Entries[i]
			switch ent.Type {
			case defines.EntryType_Block:
			case defines.EntryType_Proof:
			case defines.EntryType_NewBlock:
			case defines.EntryType_Transaction:
			default:
				return errors.New("unknown entry type")
			}
		}
	case defines.MessageType_Req:
		for i:=0; i<len(msg.Reqs); i++ {
			req := msg.Reqs[i]
			switch req.Type {
			case defines.RequestType_Blocks:
			case defines.RequestType_Neighbors:
			default:
				return errors.New("unknown entry type")
			}
		}
	default:
		return errors.New("unknown msg type")
	}
	return nil
}

func (p *Pot) handleMsgWhenCompeteWinner(msg *defines.Message) error {
	// 验证消息格式与签名
	if err := msg.Verify(); err != nil {
		return err
	}

	switch msg.Type {
	case defines.MessageType_None:
		// 啥也不干
	case defines.MessageType_Data:
		for i:=0; i<len(msg.Entries); i++ {
			ent := msg.Entries[i]
			switch ent.Type {
			case defines.EntryType_Block:
			case defines.EntryType_Proof:
			case defines.EntryType_NewBlock:
			case defines.EntryType_Transaction:
			default:
				return errors.New("unknown entry type")
			}
		}
	case defines.MessageType_Req:
		for i:=0; i<len(msg.Reqs); i++ {
			req := msg.Reqs[i]
			switch req.Type {
			case defines.RequestType_Blocks:
			case defines.RequestType_Neighbors:
			default:
				return errors.New("unknown entry type")
			}
		}
	default:
		return errors.New("unknown msg type")
	}
	return nil
}

// CompeteLoser 状态
//
func (p *Pot) handleMsgWhenCompeteLoser(msg *defines.Message) error {
	// CompeteLoser状态下可能短暂落后其他节点
	//

	// 处理新区块、交易两种数据，处理Neighbors
	switch msg.Type {
	case defines.MessageType_None:
	case defines.MessageType_Data:
		// 处理所有Entry
		for i:=0; i<len(msg.Entries); i++ {
			ent := msg.Entries[i]
			switch ent.Type {
			case defines.EntryType_NewBlock:

			case defines.EntryType_Block:

			case defines.EntryType_Proof:
			case defines.EntryType_Transaction:

			default:
				return errors.New("unknown entry type")
			}
		}
	case defines.MessageType_Req:
		for i:=0; i<len(msg.Reqs); i++ {
			req := msg.Reqs[i]
			switch req.Type {
			case defines.RequestType_Blocks:
			case defines.RequestType_Neighbors:
			default:
				return errors.New("unknown entry type")
			}
		}
	default:
		return errors.New("unknown msg type")
	}

	return nil
}
