/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 5:17 PM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"errors"
	"math/rand"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
)

// send指向网络中（或者说外部依赖的网络模块）发送消息。 注意本地消息不要通过该方法使用
// 这个发送消息可以是单播也可以是多播，具体看
// 注意：send可能发生较长时间阻塞，调用时使用go send()
func (p *Pot) send(msg *defines.Message) error {
	//merr := &defines.MessageWithError{
	//	Msg: msg,
	//	Err: make(chan error),
	//}
	//p.msgout <- merr
	//return <-merr.Err

	m, err := msg.Encode()
	if err != nil {
		return err
	}
	to, err := p.pit.Get(msg.To)
	if err != nil {
		return err
	}

	// 由于本地Socket通信很快，所以为了模拟实际通信延时，这里随机睡眠一段时间
	randMs := rand.Intn(190) + 10	// 10~200ms
	time.Sleep(time.Duration(randMs) * time.Millisecond)

	return p.net.Send(to.Addr, m)
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

// 将tx广播给所有种子节点和共识节点
func (p *Pot) broadcastTx(tx *defines.Transaction) error {
	txBytes, err := tx.Encode()
	if err != nil {
		return err
	}

	// 广播
	f := func(peer *defines.PeerInfo) error {
		if peer.Id != p.id {
			msg := &defines.Message{
				Version: defines.CodeVersion,
				Type:    defines.MessageType_Txs,
				From:    p.id,
				To:      peer.Id,
				Data: [][]byte{txBytes},
			}
			if err := msg.WriteDesc("type", "onetx"); err != nil {
				return err
			}
			if err := p.signAndSendMsg(msg); err != nil {
				p.Errorf("broadcastTx: to %s fail: %v", peer.Id, err)
				return err
			} else {
				p.Debugf("broadcastTx: to %s", peer.Id)
			}
		}
		return nil
	}

	//p.pit.RangeSeeds(f)
	p.pit.RangePeers(f)

	return nil
}

// 广播自己的证明
func (p *Pot) broadcastSelfProof() error {
	// 首先需要获取证明
	nb, err := p.bc.GenNextBlock()
	if err != nil {
		return err
	}
	p.maybeNewBlock = nb // 自己的区块设为可能的新区块

	//process := p.processes.get(p.id)
	//if process == nil {
	//	return errors.New("nil self process")
	//}

	proof := &Proof{
		Id:        p.id,
		TxsNum:    int64(len(nb.Txs)),
		BlockHash: nb.SelfHash,
		//Base:      process.Hash,
		//BaseIndex: process.Index,
		Base:      nb.PrevHash,
		BaseIndex: nb.Index - 1,
	}

	// 添加到自己的proofs
	p.proofs.Add(proof)

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

	msgtype := defines.MessageType_Proof
	desc := "proof"
	if p.id != proof.Id {
		msgtype = defines.MessageType_Proof
		desc = "relayedproof"
	}

	// 广播
	f := func(peer *defines.PeerInfo) error {
		if peer.Id != p.id {
			msg := &defines.Message{
				Version: defines.CodeVersion,
				Type:    msgtype,
				From:    p.id,
				To:      peer.Id,
				Data: [][]byte{proofBytes},
			}
			if err := msg.WriteDesc("type", desc); err != nil {
				return err
			}
			if err := p.signAndSendMsg(msg); err != nil {
				p.Errorf("broadcastProof: to %s fail: %v", peer.Id, err)
				return err
			} else {
				p.Debugf("broadcastProof: to %s", peer.Id)
			}
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
		return err
	}

	// 广播
	f := func(peer *defines.PeerInfo) error {
		if peer.Id != p.id {
			msg := &defines.Message{
				Version: defines.CodeVersion,
				Type:    defines.MessageType_NewBlock,
				From:    p.id,
				To:      peer.Id,
				Base:nb.PrevHash,
				BaseIndex:nb.Index-1,
				Data: [][]byte{blockBytes},
			}
			if err := msg.WriteDesc("type", "newblock"); err != nil {
				return err
			}
			if err := p.signAndSendMsg(msg); err != nil {
				p.Errorf("broadcastNewBlock: to %s fail: %v", peer.Id, err)
				return err
			} else {
				p.Debugf("broadcastNewBlock: to %s", peer.Id)
			}
		}
		return nil
	}
	p.pit.RangePeers(f)
	p.pit.RangeSeeds(f)
	return nil
}

// 广播getNeighbors请求
// toseeds true则向种子节点广播；否则向所有节点广播
func (p *Pot) broadcastRequestNeighborsInStartup(toseeds bool) error {
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

	f := func(peer *defines.PeerInfo) error {
		if peer.Id == p.id {
			return nil
		}

		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_ReqPeers,
			From:    p.id,
			To:      peer.Id,
			Data: [][]byte{selfb},	// 自己的节点信息
		}
		if err := msg.WriteDesc("type", "req-peers"); err != nil {
			return err
		}
		if err := p.signAndSendMsg(msg); err != nil {
			p.Errorf("requestNeighbors: to %s fail: %v", peer.Id, err)
			return err
		} else {
			p.Debugf("requestNeighbors: to %s", peer.Id)
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

// 广播请求所有节点最新进度
// toseeds true表示向seeds请求所有共识节点进度
// false 表示seeds/peers都问
//func (p *Pot) broadcastRequestProcesses(toseeds bool) error {
//	p.nWait = 0 // 重置
//	req := &defines.Request{
//		Type: defines.RequestType_Processes,
//	}
//	f := func(peer *defines.PeerInfo) error {
//		if peer.Id == p.id {
//			return nil
//		}
//
//		msg := &defines.Message{
//			Version: defines.CodeVersion,
//			Type:    defines.MessageType_Req,
//			From:    p.id,
//			To:      peer.Id,
//			Reqs:    []*defines.Request{req},
//		}
//		if err := msg.WriteDesc("type", "process"); err != nil {
//			return err
//		}
//		if err := p.signAndSendMsg(msg); err != nil {
//			p.Errorf("broadcastRequestProcesses: to %s fail: %v", peer.Id, err)
//			return err
//		} else {
//			p.Debugf("broadcastRequestProcesses: to %s", peer.Id)
//			p.nWait++
//		}
//
//		return nil
//	}
//	p.pit.RangeSeeds(f)
//	if !toseeds {
//		p.pit.RangePeers(f)
//	}
//
//	return nil
//}

// 广播请求区块，追上最新进度
// random3 true则指向最新进度的3个节点发送请求区块消息（这里需要注意的是响应端会将所有确定的区块返回）
// false 则向全部最新进度节点发送请求区块消息
//func (p *Pot) broadcastRequestBlocksInStartup(random3 bool) error {
//
//	p.nWait = 0 // 重置
//
//	process := p.processes.get(p.id)
//	req := &defines.Request{
//		Type:       defines.RequestType_Blocks,
//		IndexStart: process.Index + 1, // 自己的进度+1
//		IndexCount: 0,                 // 0表示响应端要回复所有
//	}
//
//	// 收集要广播的节点：seeds + 若干peer
//	var peers []string
//	seeds := p.pit.Seeds()
//	for id := range seeds {
//		peers = append(peers, id)
//	}
//
//	// 暂时不用random3这样的设定
//	//if random3 {
//	//	peers = append(peers, p.processes.nLatestPeers(3)...)
//	//} else {
//	//	peers = append(peers, p.processes.nLatestPeers(0)...)
//	//}
//
//	for _, peer := range peers {
//		if peer == p.id {
//			continue
//		}
//		msg := &defines.Message{
//			Version: defines.CodeVersion,
//			Type:    defines.MessageType_Req,
//			From:    p.id,
//			To:      peer,
//			Reqs:    []*defines.Request{req},
//		}
//		if err := msg.WriteDesc("type", "req-blocks"); err != nil {
//			return err
//		}
//		if err := p.signAndSendMsg(msg); err != nil {
//			p.Errorf("broadcastRequestBlocks: to %s fail: %s", peer, err)
//			return err
//		} else {
//			p.Debugf("broadcastRequestBlocks: to %s", peer)
//			p.nWait++
//		}
//	}
//
//	return nil
//}

// 广播请求区块
// start, end 为请求的区块的index区间。 start为0则不使用index索引
// hashes 请求的区块的哈希。 hashes为nil或长度为0，则不使用哈希查找
// nPeers, nSeeds 对请求的peer和seed数量做限制. <0则不生效
// ids 手动指定向哪些节点请求
//func (p *Pot) requestBlocks(start, end int64, hashes [][]byte, nPeers, nSeeds int, ids ...string) error {
//
//	process := p.processes.get(p.id)
//	req := &defines.Request{
//		Type:       defines.RequestType_Blocks,
//		IndexStart: process.Index + 1, // 自己的进度+1
//		IndexCount: 0,                 // 0表示响应端要回复所有
//	}
//
//	// 构造请求
//	req := &defines.Request{
//		Type:       defines.RequestType_Blocks,
//		IndexStart: process.Index + 1, // 自己的进度+1
//		IndexCount: 0,                 // 0表示响应端要回复所有
//	}
//
//	// 收集要广播的节点：seeds + 若干peer
//	tos := make(map[string]struct{})
//	seeds := p.pit.Seeds()
//	if nSeeds < 0 || nSeeds > len(seeds) {
//		nSeeds = len(seeds)
//	}
//	for id := range seeds {
//		if nSeeds > 0 {
//			tos[id] = struct{}{}
//			nSeeds--
//		} else {
//			break
//		}
//	}
//	peers := p.pit.Peers()
//	if nPeers < 0 || nPeers > len(peers) {
//		nPeers = len(peers)
//	}
//	for id := range peers {
//		if nPeers > 0 {
//			tos[id] = struct{}{}
//			nSeeds--
//		} else {
//			break
//		}
//	}
//	//
//	if n := len(ids); n > 0 {
//		for i := 0; i < n; i++ {
//			tos[ids[i]] = struct{}{}
//		}
//	}
//
//	for _, peer := range peers {
//		if peer == p.id {
//			continue
//		}
//		msg := &defines.Message{
//			Version: defines.CodeVersion,
//			Type:    defines.MessageType_Req,
//			From:    p.id,
//			To:      peer,
//			Reqs:    []*defines.Request{req},
//		}
//		if err := msg.WriteDesc("type", "req-blocks"); err != nil {
//			return err
//		}
//		if err := p.signAndSendMsg(msg); err != nil {
//			p.Errorf("broadcastRequestBlocks: to %s fail: %s", peer, err)
//			return err
//		} else {
//			p.Debugf("broadcastRequestBlocks: to %s", peer)
//			p.nWait++
//		}
//	}
//
//	return nil
//}

func (p *Pot) broadcastRequestBlocksByIndex(start int64, count int64) error {

	req := &defines.Message{
		Version:            defines.CodeVersion,
		Type:               defines.MessageType_ReqBlockByIndex,
		Epoch:              p.epoch,
		From:               p.id,
		To:                 "",
		ReqBlockIndexStart: start,
		ReqBlockIndexCount: count,
		Desc:               "",
	}
	seeds := p.pit.Seeds()
	for _, seed := range seeds {
		req.To = seed.Id
		p.signAndSendMsg(req)
	}
	return nil
}

// 向种子节点广播获取最新区块
func (p *Pot) broadcastRequestLatestBlock() error {

	p.nWait = 0 // 重置

	//req := &defines.Request{
	//	Type:       defines.RequestType_Blocks,
	//	IndexStart: -1, // 请求对方的最新区块
	//	IndexCount: 1,  // 最新的那个区块
	//}

	p.pit.RangeSeeds(func(peer *defines.PeerInfo) error {
		if peer.Id == p.id {
			return nil
		}
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_ReqBlockByIndex,
			From:    p.id,
			To:      peer.Id,
			ReqBlockIndexStart:-1,
			ReqBlockIndexCount:1,
		}
		if err := msg.WriteDesc("type", "req-latestblock"); err != nil {
			return err
		}
		if err := p.signAndSendMsg(msg); err != nil {
			p.Errorf("broadcastRequestBlocks: to %s fail: %s", peer.Id, err)
			return err
		} else {
			p.Debugf("broadcastRequestBlocks: to %s", peer.Id)
			p.nWait++
		}
		return nil
	})

	return nil
}

///////////////////////////////////////////////////////////////////

// wait 函数用于等待邻居们的某一类消息回应
func (p *Pot) wait(nWait int) error {
	timeoutD := time.Duration(2*TickMs) * time.Millisecond
	timeout := time.NewTimer(timeoutD)
	if p.nWaitChan == nil {
		p.nWaitChan = make(chan int)
	}

	cnt := 0
	for {
		select {
		case <-p.done:
			p.Debugf("wait: done and return")
			return nil
		case <-p.nWaitChan:
			nWait--
			cnt++
			p.Debugf("wait: nWait--")
			// 等待结束
			if nWait == 0 {
				p.Debugf("wait: wait finish and return")
				return nil
			}
		case <-timeout.C:
			// 超时需要判断两种情况：
			if cnt == 0 { // 一个回复都没收到
				p.Errorf("wait: timeout, no response received")
				return errors.New("wait timeout and no response received")
			}
			p.Debugf("wait: timeout, %d responses received, return", cnt)
			return nil
		}
	}
}

// 等待某个区块，需要在wait阶段决定哪个才是正确的
// blockIndex=-1时表示等最新区块; nWait表示等待的数量
func (p *Pot) waitAndDecideOneBlock(blockIndex int64, nWait int) (*defines.Block, error) {
	timeoutD := time.Duration(2*TickMs) * time.Millisecond
	timeout := time.NewTimer(timeoutD)
	if p.nWaitBlockChan == nil {
		p.nWaitBlockChan = make(chan *defines.Block)
	}

	p.udbt.Reset(blockIndex) // 重置未决区块表

	cnt := 0
	for {
		select {
		case <-p.done: // 程序被关闭
			p.Debugf("wait: done and return")
			return nil, nil
		case b := <-p.nWaitBlockChan:
			nWait--
			cnt++
			p.Debugf("wait: nWait--")
			p.udbt.Add(b) // 添加到未决区块表
			// 等待结束
			if nWait == 0 {
				p.Debugf("wait: wait finish and return")
				return p.udbt.Major(), nil
			}
		case <-timeout.C:
			// 超时需要判断两种情况：
			if cnt == 0 { // 一个回复都没收到
				p.Errorf("wait: timeout, no response received")
				return nil, errors.New("wait timeout and no response received")
			}
			p.Debugf("wait: timeout, %d responses received, return", cnt)
			return p.udbt.Major(), nil
		}
	}
}
