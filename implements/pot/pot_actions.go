/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 5:17 PM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"errors"
	"github.com/azd1997/blockchain-consensus/defines"
	"time"
)

// send指向网络中（或者说外部依赖的网络模块）发送消息。 注意本地消息不要通过该方法使用
// 这个发送消息可以是单播也可以是多播，具体看
// 注意：send可能发生较长时间阻塞，调用时使用go send()
func (p *Pot) send(msg *defines.Message) error {
	merr := &defines.MessageWithError{
		Msg: msg,
		Err: make(chan error),
	}
	p.msgout <- merr
	return <- merr.Err
}

// 酌情使用 go signAndSendMsg()
func (p *Pot) signAndSendMsg(msg *defines.Message) error {
	if msg == nil {
		return errors.New("nil msg")
	}
	err := msg.Sign()
	if err != nil {
		return err
	}

	return p.send(msg)
}

//// broadcast
//func (n *Node) broadcast(msg *defines.Message) {
//
//}

// 广播getNeighbors请求
// toseeds true则向种子节点广播；否则向所有节点广播
func (p *Pot) broadcastRequestNeighbors(toseeds bool) error {
	p.nWait = 0

	// 查询自身节点信息
	self, err := p.pit.Get(p.id)
	if err != nil {
		return err
	}
	selfb, err := self.Encode()
	if err != nil {
		return err
	}

	// 构造请求
	req := &defines.Request{
		Type: defines.RequestType_Neighbors,
		Data: selfb,
	}
	f := func(peer *defines.PeerInfo) error {
		if peer.Id == p.id {
			return nil
		}

		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Req,
			From:    p.id,
			To:      peer.Id,
			Reqs:    []*defines.Request{req},
		}
		err = p.signAndSendMsg(msg)
		if err != nil {
			p.Errorf("broadcastRequestNeighbors: to %s fail: %v\n", peer.Id, err)
		} else {
			p.Logf("broadcastRequestNeighbors: to %s\n", peer.Id)
			p.nWait++
		}

		return nil
	}

	p.pit.RangeSeeds(f)
	if !toseeds {
		p.pit.RangePeers(f)
	}
	return nil
}

// 广播自己的证明
func (p *Pot) broadcastSelfProof() error {
	// 首先需要获取证明
	nb := p.txPool.GenBlock()
	p.maybeNewBlock = nb

	process := p.processes.get(p.id)
	if process == nil {
		return errors.New("nil self process")
	}

	proof := &Proof{
		Id:        p.id,
		TxsNum:    uint64(len(nb.Txs)),
		TxsMerkle: nb.Merkle,
		Base:      process.Hash,
		BaseIndex: process.Index,
	}

	return p.broadcastProof(proof, false)
}

// 广播证明
// 目前的方案中，是seed节点用来广播winner的Proof的
//
func (p *Pot) broadcastProof(proof *Proof, onlypeers bool) error {
	proofBytes, err := proof.Encode()
	if err != nil {
		return err
	}
	entry := &defines.Entry{
		Type: defines.EntryType_Proof,
		Data: proofBytes,
	}

	// 广播
	f := func(peer *defines.PeerInfo) error {
		if peer.Id != p.id {
			msg := &defines.Message{
				Version: defines.CodeVersion,
				Type:    defines.MessageType_Data,
				From:    p.id,
				To:      peer.Id,
				Entries: []*defines.Entry{entry},
			}
			return p.signAndSendMsg(msg)
		}
		return nil
	}

	if !onlypeers {
		p.pit.RangeSeeds(f)
	}
	p.pit.RangePeers(f)

	return nil
}

// 广播新区块
func (p *Pot) broadcastNewBlock(nb *defines.Block) error {

	blockBytes, err := nb.Encode()
	if err != nil {
		//p.Errorf("broadcast new block(%v) fail: %s\n", nb, err)
		return err
	}

	entry := &defines.Entry{
		Base:      nb.PrevHash,
		BaseIndex: nb.Index - 1,
		Type:      defines.EntryType_NewBlock,
		Data:      blockBytes,
	}

	// 广播
	p.pit.RangePeers(func(peer *defines.PeerInfo) error {
		if peer.Id != p.id {
			msg := &defines.Message{
				Version: defines.CodeVersion,
				Type:    defines.MessageType_Data,
				From:    p.id,
				To:      peer.Id,
				Entries: []*defines.Entry{entry},
			}
			err := msg.Sign()
			if err != nil {
				return err
			}
			p.send(msg)
		}
		return nil
	})
	return nil
}

// 广播请求所有节点最新进度
// toseeds true表示向seeds请求所有共识节点进度
// false 表示seeds/peers都问
func (p *Pot) broadcastRequestProcesses(toseeds bool) error {
	p.nWait = 0	// 重置
	req := &defines.Request{
		Type: defines.RequestType_Processes,
	}
	f := func(peer *defines.PeerInfo) error {
		if peer.Id == p.id {return nil}

		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Req,
			From:    p.id,
			To:      peer.Id,
			Reqs:    []*defines.Request{req},
		}
		err := p.signAndSendMsg(msg)
		if err != nil {
			p.Errorf("broadcastRequestProcesses: to %s fail: %v\n", peer.Id, err)
		} else {
			p.Logf("broadcastRequestProcesses: to %s\n", peer.Id)
			p.nWait++
		}

		return nil
	}
	p.pit.RangeSeeds(f)
	if !toseeds {
		p.pit.RangePeers(f)
	}

	return nil
}

// 广播请求区块，追上最新进度
// random3 true则指向最新进度的3个节点发送请求区块消息（这里需要注意的是响应端会将所有确定的区块返回）
// false 则向全部最新进度节点发送请求区块消息
func (p *Pot) broadcastRequestBlocks(random3 bool) error {

	p.nWait = 0		// 重置

	process := p.processes.get(p.id)
	req := &defines.Request{
		Type:       defines.RequestType_Blocks,
		IndexStart: int64(process.Index+1),	// 自己的进度+1
		IndexCount: 0,	// 0表示响应端要回复所有
	}

	// 收集要广播的节点：seeds + 若干peer
	var peers []string
	seeds := p.pit.Seeds()
	for id := range seeds {
		peers = append(peers, id)
	}

	if random3 {
		peers = append(peers, p.processes.nLatestPeers(3)...)
	} else {
		peers = append(peers, p.processes.nLatestPeers(0)...)
	}

	for _, peer := range peers {
		if peer == p.id {continue}
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Req,
			From:    p.id,
			To:      peer,
			Reqs:    []*defines.Request{req},
		}
		if err := p.signAndSendMsg(msg); err != nil {
			p.Errorf("broadcastRequestBlocks: to %s fail: %s\n", peer, err)
		} else {
			p.Logf("broadcastRequestBlocks: to %s\n", peer)
			p.nWait++
		}
	}

	return nil
}

// 向种子节点广播获取最新区块
func (p *Pot) broadcastRequestLatestBlock() error {

	p.nWait = 0		// 重置

	req := &defines.Request{
		Type:       defines.RequestType_Blocks,
		IndexStart: -1,	// 请求对方的最新区块
		IndexCount: 1,	// 最新的那个区块
	}

	p.pit.RangeSeeds(func(peer *defines.PeerInfo) error {
		if peer.Id == p.id {return nil}
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Req,
			From:    p.id,
			To:      peer.Id,
			Reqs:    []*defines.Request{req},
		}
		if err := p.signAndSendMsg(msg); err != nil {
			p.Errorf("broadcastRequestBlocks: to %s fail: %s\n", peer.Id, err)
		} else {
			p.Logf("broadcastRequestBlocks: to %s\n", peer.Id)
			p.nWait++
		}
		return nil
	})

	return nil
}

// wait() 函数用于等待邻居们的某一类消息回应
func (p *Pot) wait() error {
	timeoutD := 1 * time.Second
	timeout := time.NewTimer(timeoutD)
	if p.nWaitChan == nil {
		p.nWaitChan = make(chan int)
	}

	cnt := 0
	for {
		select {
		case <-p.done:
			p.Logf("wait: done and return\n")
			return nil
		case <-p.nWaitChan:
			p.nWait--
			cnt++
			p.Logf("wait: nWait--\n")
			// 等待结束
			if p.nWait == 0 {
				p.Logf("wait: wait finish and return\n")
				return nil
			}
		case <-timeout.C:
			// 超时需要判断两种情况：
			if cnt == 0 {	// 一个回复都没收到
				p.Logf("wait: timeout, no response received\n")
				return errors.New("wait timeout and no response received")
			} else {
				p.Logf("wait: timeout, %d responses received, return\n", cnt)
				return nil
			}
		}
	}
}