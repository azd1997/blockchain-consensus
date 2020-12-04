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
	case StateType_PreInited_RequestNeighbors:
		return p.handleMsgWhenPreInitedRN(msg)
	case StateType_PreInited_RequestFirstBlock:
		return p.handleMsgWhenPreInitedRFB(msg)
	case StateType_PreInited_RequestLatestBlock:
		return p.handleMsgWhenPreInitedRLB(msg)
	case StateType_NotReady:
		return p.handleMsgWhenNotReady(msg)
	case StateType_InPot:
		return p.handleMsgWhenInPot(msg)
	case StateType_PostPot:
		return p.handleMsgWhenPostPot(msg)
	default:
		p.Fatalf("unknown state(%d-%s)\n", state, state.String())
	}
	return nil
}

func (p *Pot) handleMsgWhenPreInitedRN(msg *defines.Message) error {
	duty := p.duty
	switch duty {
	case defines.PeerDuty_None:
		return p.handleMsgWhenPreInitedRNForDutyNone(msg)
	case defines.PeerDuty_Peer:
		return p.handleMsgWhenPreInitedRNForDutyPeer(msg)
	case defines.PeerDuty_Seed:
		return p.handleMsgWhenPreInitedRNForDutySeed(msg)
	default:
		p.Fatalf("unknown duty(%v)\n", duty)
	}
}

func (p *Pot) handleMsgWhenPreInitedRFB(msg *defines.Message) error {
	duty := p.duty
	switch duty {
	case defines.PeerDuty_None:
		return p.handleMsgWhenPreInitedRFBForDutyNone(msg)
	case defines.PeerDuty_Peer:
		return p.handleMsgWhenPreInitedRFBForDutyPeer(msg)
	case defines.PeerDuty_Seed:
		return p.handleMsgWhenPreInitedRFBForDutySeed(msg)
	default:
		p.Fatalf("unknown duty(%v)\n", duty)
	}
}

func (p *Pot) handleMsgWhenPreInitedRLB(msg *defines.Message) error {
	duty := p.duty
	switch duty {
	case defines.PeerDuty_None:
		return p.handleMsgWhenPreInitedRLBForDutyNone(msg)
	case defines.PeerDuty_Peer:
		return p.handleMsgWhenPreInitedRLBForDutyPeer(msg)
	case defines.PeerDuty_Seed:
		return p.handleMsgWhenPreInitedRLBForDutySeed(msg)
	default:
		p.Fatalf("unknown duty(%v)\n", duty)
	}
}

func (p *Pot) handleMsgWhenNotReady(msg *defines.Message) error {
	duty := p.duty
	switch duty {
	case defines.PeerDuty_None:
		return p.handleMsgWhenNotReadyForDutyNone(msg)
	case defines.PeerDuty_Peer:
		return p.handleMsgWhenNotReadyForDutyPeer(msg)
	case defines.PeerDuty_Seed:
		return p.handleMsgWhenNotReadyForDutySeed(msg)
	default:
		p.Fatalf("unknown duty(%v)\n", duty)
	}
}

func (p *Pot) handleMsgWhenInPot(msg *defines.Message) error {
	duty := p.duty
	switch duty {
	case defines.PeerDuty_None:
		return p.handleMsgWhenInPotForDutyNone(msg)
	case defines.PeerDuty_Peer:
		return p.handleMsgWhenInPotForDutyPeer(msg)
	case defines.PeerDuty_Seed:
		return p.handleMsgWhenInPotForDutySeed(msg)
	default:
		p.Fatalf("unknown duty(%v)\n", duty)
	}
}

func (p *Pot) handleMsgWhenPostPot(msg *defines.Message) error {
	duty := p.duty
	switch duty {
	case defines.PeerDuty_None:
		return p.handleMsgWhenPostPotForDutyNone(msg)
	case defines.PeerDuty_Peer:
		return p.handleMsgWhenPostPotForDutyPeer(msg)
	case defines.PeerDuty_Seed:
		return p.handleMsgWhenPostPotForDutySeed(msg)
	default:
		p.Fatalf("unknown duty(%v)\n", duty)
	}
}

//////////////////////////////////////////////////

// Message分Data和Req两大类

// PreInited_RN阶段
// 仅处理邻居消息

func (p *Pot) handleMsgWhenPreInitedRNForDutyNone(msg *defines.Message) error {
	return p.handleMsgWhenPreInitedRNForAllDuty(msg)
}

func (p *Pot) handleMsgWhenPreInitedRNForDutyPeer(msg *defines.Message) error {
	return p.handleMsgWhenPreInitedRNForAllDuty(msg)	
}

func (p *Pot) handleMsgWhenPreInitedRNForDutySeed(msg *defines.Message) error {
	return p.handleMsgWhenPreInitedRNForAllDuty(msg)
}

// 再PreInited阶段，所有类型的节点能处理的消息是相同的
func (p *Pot) handleMsgWhenPreInitedRNForAllDuty(msg *defines.Message) error {
	switch msg.Type {
	case defines.MessageType_None:
		p.Errorf("%s can only handle [EntryType_Neighbor]\n", p.DutyState())
	case defines.MessageType_Data:
		count := 0
		for _, ent := range msg.Entries {
			ent := ent
			if ent.Type == EntryType_Neighbor {
				count++
				// 通用的handleEntryNeighbor，添加就完事
				if err := p.handleEntryNeighbor(msg.From, ent); err != nil {
					p.Errorf("%s handle EntryType_Neighbor from (%s) fail: %s\n", p.DutyState(), msg.From, err)
				} else {
					p.Logf("%s handle EntryType_Neighbor from (%s) succ\n", p.DutyState(), msg.From)
				}
			}
		}
		if count > 0 && p.nWaitChan != nil { // 说明包含Neighbors
			p.nWaitChan <- 1 // 通知收到一个节点回传了节点信息表
		}
	case defines.MessageType_Req:
		p.Errorf("%s can only handle [EntryType_Neighbor]\n", p.DutyState())
	default:
		p.Errorf("%s met unknown msg type(%v)\n", p.DutyState(), msg.Type)
	}
	return nil
}

// PreInited_RFB阶段
// 仅处理区块消息，且是仅包含1个1号区块的区块消息

func (p *Pot) handleMsgWhenPreInitedRFBForDutyNone(msg *defines.Message) error {
	return p.handleMsgWhenPreInitedRFBForAllDuty(msg)
}

func (p *Pot) handleMsgWhenPreInitedRFBForDutyPeer(msg *defines.Message) error {
	return p.handleMsgWhenPreInitedRFBForAllDuty(msg)	
}

func (p *Pot) handleMsgWhenPreInitedRFBForDutySeed(msg *defines.Message) error {
	return p.handleMsgWhenPreInitedRFBForAllDuty(msg)
}

// 再PreInited阶段，所有类型的节点能处理的消息是相同的
func (p *Pot) handleMsgWhenPreInitedRFBForAllDuty(msg *defines.Message) error {
	switch msg.Type {
	case defines.MessageType_None:
		p.Errorf("%s can only handle [EntryType_Block]\n", p.DutyState())
	case defines.MessageType_Data:
		count := len(msg.Entries)
		// 等待的是1号区块，其baseindex=0
		if count == 0 || count != 1 || msg.Entries[0].Type != EntryType_Block || msg.Entries[0].BaseIndex > 0 {
			p.Errorf("%s received a unexpected msg from %s\n", p.DutyState(), msg.From)
		}
		firstBlock := new(defines.Block)
		err := firstBlock.Decode(msg.Entries[0].Data)
		if err != nil {
			return err
		}
		if p.nWaitBlockChan != nil {
			p.nWaitBlockChan <- firstBlock // 通知收到一个节点回传了1号区块
			p.Logf("%s handle EntryType_Block from (%s) succ\n", p.DutyState(), msg.From)
		}
		// 这个firstBlock由启动逻辑确定之后再写到本地

	case defines.MessageType_Req:
		p.Errorf("%s can only handle [EntryType_Block]\n", p.DutyState())
	default:
		p.Errorf("%s met unknown msg type(%v)\n", p.DutyState(), msg.Type)
	}
	return nil
}

// PreInited_RLB阶段
// 仅处理区块消息，且是仅包含1个最新区块(序号未知)的区块消息

func (p *Pot) handleMsgWhenPreInitedRLBForDutyNone(msg *defines.Message) error {
	return p.handleMsgWhenPreInitedRLBForAllDuty(msg)
}

func (p *Pot) handleMsgWhenPreInitedRLBForDutyPeer(msg *defines.Message) error {
	return p.handleMsgWhenPreInitedRLBForAllDuty(msg)	
}

func (p *Pot) handleMsgWhenPreInitedRLBForDutySeed(msg *defines.Message) error {
	return p.handleMsgWhenPreInitedRLBForAllDuty(msg)
}

// 再PreInited阶段，所有类型的节点能处理的消息是相同的
func (p *Pot) handleMsgWhenPreInitedRLBForAllDuty(msg *defines.Message) error {
	switch msg.Type {
	case defines.MessageType_None:
		p.Errorf("%s can only handle [EntryType_Block]\n", p.DutyState())
	case defines.MessageType_Data:
		count := len(msg.Entries)
		// 等待的是最新区块，其序号未知
		if count == 0 || count != 1 || msg.Entries[0].Type != EntryType_Block {
			p.Errorf("%s received a unexpected msg from %s\n", p.DutyState(), msg.From)
		}
		latestBlock := new(defines.Block)
		err := latestBlock.Decode(msg.Entries[0].Data)
		if err != nil {
			return err
		}
		if p.nWaitBlockChan != nil {
			p.nWaitBlockChan <- latestBlock // 通知收到一个节点回传了最新区块
			p.Logf("%s handle EntryType_Block from (%s) succ\n", p.DutyState(), msg.From)
		}
		// 这个latestBlock由启动逻辑确定之后再写到本地

	case defines.MessageType_Req:
		p.Errorf("%s can only handle [EntryType_Block]\n", p.DutyState())
	default:
		p.Errorf("%s met unknown msg type(%v)\n", p.DutyState(), msg.Type)
	}
	return nil
}

// NotReady阶段
// Req中只能处理邻居请求，能处理Data消息中的一部分
// Req最常见的就是请求邻居和请求区块
//
// NotReady 的条件：进度未补全。 当本机节点进度被补全，且再过一段时间到达邻近的PotStart时刻时，节点切换状态至InPot
// 显然是否能接受区块请求应该看是否IsSelfReady，而不是看当前是否处于NotReady状态.
// 或者不管是否Ready，如果本机有，就返回给请求方

func (p *Pot) handleMsgWhenNotReadyForDutyNone(msg *defines.Message) error {
	return nil
}

func (p *Pot) handleMsgWhenNotReadyForDutyPeer(msg *defines.Message) error {
	return nil
}

func (p *Pot) handleMsgWhenNotReadyForDutySeed(msg *defines.Message) error {
	switch msg.Type {
	case defines.MessageType_None:
		p.Errorf("%s can only handle [EntryType_Block]\n", p.DutyState())
	case defines.MessageType_Data:
		for _, ent := range msg.Entries {
			ent := ent
			switch ent.Type {
			case EntryType_Block:
				// 在NotReady状态下接收过往区块需要注意，所有接收到的区块临时存到一个哈希表
				// 且按LatestBlock倒序补漏
				if err := p.handleEntryBlock(msg.From, req); err != nil {
					p.Errorf("%s handle EntryType_Block from (%s) fail: %s\n", p.DutyState(), msg.From, err)
				} else {
					p.Logf("%s handle EntryType_Block from (%s) succ\n", p.DutyState(), msg.From)
				}
			case EntryType_Proof:	
				// 收集Proof. NotReady只是不竞选不校验，不代表不见证
				if err := p.handleEntryBlock(msg.From, req); err != nil {
					p.Errorf("%s handle EntryType_Block from (%s) fail: %s\n", p.DutyState(), msg.From, err)
				} else {
					p.Logf("%s handle EntryType_Block from (%s) succ\n", p.DutyState(), msg.From)
				}
			case EntryType_NewBlock:
				// 收集新区块

			case EntryType_Transaction:
			case EntryType_Neighbor:
			case EntryType_Process:

			default:
				p.Errorf("%s met unknown entry type(%v)\n", p.DutyState(), ent.Type)	
			}
		}
	case defines.MessageType_Req:
		for _, req := range msg.Reqs {
			req := req
			switch req.Type {
			case RequestType_Neighbors:
				if err := p.handleRequestNeighbors(msg.From, req); err != nil {
					p.Errorf("%s handle RequestType_Neighbors from (%s) fail: %s\n", p.DutyState(), msg.From, err)
				} else {
					p.Logf("%s handle RequestType_Neighbors from (%s) succ\n", p.DutyState(), msg.From)
				}
			case RequestType_Blocks:
				if err := p.handleRequestBlocks(msg.From, req); err != nil {
					p.Errorf("%s handle RequestType_Blocks from (%s) fail: %s\n", p.DutyState(), msg.From, err)
				} else {
					p.Logf("%s handle RequestType_Blocks from (%s) succ\n", p.DutyState(), msg.From)
				}
			case RequestType_Processes:
				if err := p.handleRequestProcesses(msg.From, req); err != nil {
					p.Errorf("%s handle RequestType_Processes from (%s) fail: %s\n", p.DutyState(), msg.From, err)
				} else {
					p.Logf("%s handle RequestType_Processes from (%s) succ\n", p.DutyState(), msg.From)
				}
			default:
				p.Errorf("%s met unknown req type(%v)\n", p.DutyState(), req.Type)	
			}
		}
	default:
		p.Errorf("%s met unknown msg type(%v)\n", p.DutyState(), msg.Type)
	}
	
	
	return nil
}

// InPot阶段
// 处理Req消息和Data消息

func (p *Pot) handleMsgWhenInPotForDutyNone(msg *defines.Message) error {
	return nil
}

func (p *Pot) handleMsgWhenInPotForDutyPeer(msg *defines.Message) error {
	return nil
}

func (p *Pot) handleMsgWhenInPotForDutySeed(msg *defines.Message) error {
	return nil
}

// PostPot阶段
// 处理Req消息和Data消息

func (p *Pot) handleMsgWhenPostPotForDutyNone(msg *defines.Message) error {
	return nil
}

func (p *Pot) handleMsgWhenPostPotForDutyPeer(msg *defines.Message) error {
	return nil
}

func (p *Pot) handleMsgWhenPostPotForDutySeed(msg *defines.Message) error {
	return nil
}


// /////////////////////////// 处理消息 /////////////////////////

// // 目前的消息只有类：区块消息、证明消息、
// // 快照消息（用于NotReady的节点快速获得区块链部分，目前直接用区块消息代表）。

// // 只处理Neighbors消息
// func (p *Pot) handleMsgWhenInitGetNeighbors(msg *defines.Message) error {
// 	// 验证消息格式与签名
// 	if err := msg.Verify(); err != nil {
// 		return err
// 	}

// 	switch msg.Type {
// 	case defines.MessageType_Data:
// 		count := len(msg.Entries)
// 		for _, ent := range msg.Entries {
// 			switch ent.Type {
// 			case defines.EntryType_Neighbor:
// 				count--
// 				err := p.handleEntryNeighbor(msg.From, ent)
// 				if err != nil {
// 					p.Errorf("%s handle EntryType_Neighbor from (%s) fail: %s\n", p.getState().String(), msg.From, err)
// 				} else {
// 					p.Logf("%s handle EntryType_Neighbor from (%s) succ\n", p.getState().String(), msg.From)
// 				}
// 			default: // 其他类型则忽略
// 				p.Errorf("%s can only handle EntryType_Neighbor\n", p.getState().String())
// 			}
// 		}
// 		if count == 0 { // 说明全是Neighbors
// 			p.nWaitChan <- 1 // 通知收到一个节点回传了节点信息表
// 		}
// 	case defines.MessageType_Req:
// 		// 该阶段不能处理Req消息
// 		p.Errorf("%s can only handle EntryType_Neighbor\n", p.getState().String())
// 	default:
// 		p.Errorf("unknown msg type")
// 	}
// 	return nil
// }

// // 收集进度
// func (p *Pot) handleMsgWhenInitGetProcesses(msg *defines.Message) error {
// 	// 验证消息格式与签名
// 	if err := msg.Verify(); err != nil {
// 		return err
// 	}

// 	switch msg.Type {
// 	case defines.MessageType_Data:
// 		count := len(msg.Entries)
// 		for _, ent := range msg.Entries {
// 			switch ent.Type {
// 			case defines.EntryType_Neighbor:
// 				err := p.handleEntryNeighbor(msg.From, ent)
// 				if err != nil {
// 					p.Errorf("%s handle EntryType_Neighbor from (%s) fail: %s\n", p.getState().String(), msg.From, err)
// 				} else {
// 					p.Logf("%s handle EntryType_Neighbor from (%s) succ\n", p.getState().String(), msg.From)
// 				}
// 			case defines.EntryType_Process:
// 				count--
// 				err := p.handleEntryProcess(msg.From, ent)
// 				if err != nil {
// 					p.Errorf("%s handle EntryType_Process from (%s) fail: %s\n", p.getState().String(), msg.From, err)
// 				} else {
// 					p.Logf("%s handle EntryType_Process from (%s) succ\n", p.getState().String(), msg.From)
// 				}
// 			default: // 其他类型则忽略
// 				p.Errorf("%s can only handle EntryType_Neighbor or EntryType_Process\n", p.getState().String())
// 			}
// 		}
// 		if count == 0 {
// 			p.nWaitChan <- 1
// 		}
// 	case defines.MessageType_Req:
// 		// 该阶段不能处理Req消息
// 		p.Errorf("%s can only handle EntryType_Neighbor\n", p.getState().String())
// 	default:
// 		p.Errorf("unknown msg type")
// 	}
// 	return nil
// }

// // 收集区块追赶进度
// func (p *Pot) handleMsgWhenInitGetBlocks(msg *defines.Message) error {
// 	// 验证消息格式与签名
// 	if err := msg.Verify(); err != nil {
// 		return err
// 	}

// 	switch msg.Type {
// 	case defines.MessageType_Data:
// 		for _, ent := range msg.Entries {
// 			switch ent.Type {
// 			case defines.EntryType_Neighbor:
// 				err := p.handleEntryNeighbor(msg.From, ent)
// 				if err != nil {
// 					p.Errorf("%s handle EntryType_Neighbor fail: %s\n", p.getState().String(), err)
// 				}
// 			case defines.EntryType_Process:
// 				err := p.handleEntryProcess(msg.From, ent)
// 				if err != nil {
// 					p.Errorf("%s handle EntryType_Progress fail: %s\n", p.getState().String(), err)
// 				}
// 			default: // 其他类型则忽略
// 				p.Errorf("%s can only handle EntryType_Neighbor or EntryType_Process\n", p.getState().String())
// 			}
// 		}
// 	case defines.MessageType_Req:
// 		// 该阶段不能处理Req消息
// 		p.Errorf("%s can only handle EntryType_Neighbor\n", p.getState().String())
// 	default:
// 		p.Errorf("unknown msg type")
// 	}
// 	return nil
// }

// func (p *Pot) handleMsgWhenNotReady(msg *defines.Message) error {
// 	// 验证消息格式与签名
// 	if err := msg.Verify(); err != nil {
// 		return err
// 	}

// 	switch msg.Type {
// 	case defines.MessageType_Data:
// 		// 该阶段只能处理block消息
// 		// 根据EntryType来处理
// 		for _, ent := range msg.Entries {
// 			switch ent.Type {
// 			case defines.EntryType_Block:
// 				return p.handleEntryBlock(msg.From, ent)
// 			default: // 其他类型则忽略
// 			}
// 		}
// 	case defines.MessageType_Req:
// 		// 该阶段不能处理Req消息
// 	default:
// 		// 报错打日志
// 	}
// 	return nil
// }

// // ReadyCompete状态下可以收集交易、新区块，以及区块和邻居请求
// func (p *Pot) handleMsgWhenReadyCompete(msg *defines.Message) error {
// 	// 检查自身是否追上最新进度
// 	if !p.latest() {
// 		return errors.New("not latest process")
// 	}
// 	// 验证消息格式与签名
// 	if err := msg.Verify(); err != nil {
// 		return err
// 	}

// 	switch msg.Type {
// 	case defines.MessageType_Data:
// 		// ReadyCompete状态下不允许接收区块和证明
// 		for i := 0; i < len(msg.Entries); i++ {
// 			ent := msg.Entries[i]
// 			switch ent.Type {
// 			case defines.EntryType_Block:
// 				return p.handleEntryBlock(msg.From, ent)
// 			case defines.EntryType_Proof:
// 			case defines.EntryType_NewBlock:
// 				return p.handleEntryNewBlock(msg.From, ent)
// 			case defines.EntryType_Transaction:
// 				return p.handleEntryTransaction(msg.From, ent)
// 			default:
// 				return errors.New("unknown entry type")
// 			}
// 		}
// 	case defines.MessageType_Req:
// 		// 检查请求内容，查看
// 		if len(msg.Entries) > 0 || len(msg.Reqs) == 0 {
// 			return errors.New("not a req msg")
// 		}
// 		for i := 0; i < len(msg.Reqs); i++ {
// 			req := msg.Reqs[i]
// 			switch req.Type {
// 			case defines.RequestType_Blocks:
// 				p.handleRequestBlocks(msg.From, req)
// 			case defines.RequestType_Neighbors:
// 				p.handleRequestNeighbors(msg.From, req)
// 			default: // 其他类型则忽略
// 				//return errors.New("unknown req type")
// 			}
// 		}
// 	default:
// 		return errors.New("unknown msg type")
// 	}

// 	return nil
// }

// // Competing状态下msg处理
// func (p *Pot) handleMsgWhenCompeting(msg *defines.Message) error {
// 	// 检查自身是否追上最新进度
// 	if !p.latest() {
// 		return errors.New("not latest process")
// 	}
// 	// 验证消息格式与签名
// 	if err := msg.Verify(); err != nil {
// 		return err
// 	}

// 	switch msg.Type {
// 	case defines.MessageType_Data:
// 		// ReadyCompete状态下不允许接收区块，允许接收证明
// 		if len(msg.Entries) != 1 {
// 			return errors.New("recv proof only, but len(msg.Entries) != 1")
// 		}

// 		ent := msg.Entries[0]
// 		switch ent.Type {
// 		case defines.EntryType_Proof:
// 			return p.handleEntryProof(msg.From, ent)
// 		default: // 其他类型则忽略
// 		}

// 	case defines.MessageType_Req:
// 		// 检查请求内容，查看
// 		if len(msg.Entries) > 0 || len(msg.Reqs) == 0 {
// 			return errors.New("not a req msg")
// 		}
// 		// 根据RequestType来处理
// 		for _, req := range msg.Reqs {
// 			switch req.Type {
// 			case defines.RequestType_Blocks:
// 				return p.handleRequestBlocks(msg.From, req)
// 			case defines.RequestType_Neighbors:
// 				return p.handleRequestNeighbors(msg.From, req)
// 			default: // 其他类型则忽略
// 			}
// 		}

// 	default:
// 		return errors.New("unknown msg type")
// 	}
// 	return nil
// }

// // CompeteOver 状态
// // 此状态极为短暂，该期间只能照常处理交易、邻居消息，其他消息不处理
// func (p *Pot) handleMsgWhenCompeteOver(msg *defines.Message) error {
// 	// 验证消息格式与签名
// 	if err := msg.Verify(); err != nil {
// 		return err
// 	}

// 	switch msg.Type {
// 	case defines.MessageType_None:
// 		// 啥也不干
// 	case defines.MessageType_Data:
// 		for i := 0; i < len(msg.Entries); i++ {
// 			ent := msg.Entries[i]
// 			switch ent.Type {
// 			case defines.EntryType_Block:
// 			case defines.EntryType_Proof:
// 			case defines.EntryType_NewBlock:
// 			case defines.EntryType_Transaction:
// 			default:
// 				return errors.New("unknown entry type")
// 			}
// 		}
// 	case defines.MessageType_Req:
// 		for i := 0; i < len(msg.Reqs); i++ {
// 			req := msg.Reqs[i]
// 			switch req.Type {
// 			case defines.RequestType_Blocks:
// 			case defines.RequestType_Neighbors:
// 			default:
// 				return errors.New("unknown entry type")
// 			}
// 		}
// 	default:
// 		return errors.New("unknown msg type")
// 	}
// 	return nil
// }

// func (p *Pot) handleMsgWhenCompeteWinner(msg *defines.Message) error {
// 	// 验证消息格式与签名
// 	if err := msg.Verify(); err != nil {
// 		return err
// 	}

// 	switch msg.Type {
// 	case defines.MessageType_None:
// 		// 啥也不干
// 	case defines.MessageType_Data:
// 		for i := 0; i < len(msg.Entries); i++ {
// 			ent := msg.Entries[i]
// 			switch ent.Type {
// 			case defines.EntryType_Block:
// 			case defines.EntryType_Proof:
// 			case defines.EntryType_NewBlock:
// 			case defines.EntryType_Transaction:
// 			default:
// 				return errors.New("unknown entry type")
// 			}
// 		}
// 	case defines.MessageType_Req:
// 		for i := 0; i < len(msg.Reqs); i++ {
// 			req := msg.Reqs[i]
// 			switch req.Type {
// 			case defines.RequestType_Blocks:
// 			case defines.RequestType_Neighbors:
// 			default:
// 				return errors.New("unknown entry type")
// 			}
// 		}
// 	default:
// 		return errors.New("unknown msg type")
// 	}
// 	return nil
// }

// // CompeteLoser 状态
// //
// func (p *Pot) handleMsgWhenCompeteLoser(msg *defines.Message) error {
// 	// CompeteLoser状态下可能短暂落后其他节点
// 	//

// 	// 处理新区块、交易两种数据，处理Neighbors
// 	switch msg.Type {
// 	case defines.MessageType_None:
// 	case defines.MessageType_Data:
// 		// 处理所有Entry
// 		for i := 0; i < len(msg.Entries); i++ {
// 			ent := msg.Entries[i]
// 			switch ent.Type {
// 			case defines.EntryType_NewBlock:

// 			case defines.EntryType_Block:

// 			case defines.EntryType_Proof:
// 			case defines.EntryType_Transaction:

// 			default:
// 				return errors.New("unknown entry type")
// 			}
// 		}
// 	case defines.MessageType_Req:
// 		for i := 0; i < len(msg.Reqs); i++ {
// 			req := msg.Reqs[i]
// 			switch req.Type {
// 			case defines.RequestType_Blocks:
// 			case defines.RequestType_Neighbors:
// 			default:
// 				return errors.New("unknown entry type")
// 			}
// 		}
// 	default:
// 		return errors.New("unknown msg type")
// 	}

// 	return nil
// }
