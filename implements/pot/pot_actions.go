/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 5:17 PM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"errors"

	"github.com/azd1997/blockchain-consensus/defines"
)

// send指向网络中（或者说外部依赖的网络模块）发送消息。 注意本地消息不要通过该方法使用
// 这个发送消息可以是单播也可以是多播，具体看
func (p *Pot) send(msg *defines.Message) {
	p.msgout <- msg
}

func (p *Pot) signAndSendMsg(msg *defines.Message) error {
	if msg == nil || (msg.Check() != nil) {
		return errors.New("invalid msg")
	}
	err := msg.Sign()
	if err != nil {
		return err
	}

	p.send(msg)
	return nil
}

//// broadcast
//func (n *Node) broadcast(msg *defines.Message) {
//
//}

// 向seeds广播getNeighbors请求
func (p *Pot) broadcastRequestNeighbors() error {
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
		Type:       defines.RequestType_Neighbors,
		Data:selfb,
	}
	f := func(peer *defines.PeerInfo) error {
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Req,
			From:    p.id,
			To:      peer.Id,
			Reqs:    []*defines.Request{req},
		}
		p.nWait++
		return p.signAndSendMsg(msg)
	}
	//p.pit.RangePeers(f)
	p.pit.RangeSeeds(f)
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
		Id:p.id,
		TxsNum:    uint64(len(nb.Txs)),
		TxsMerkle: nb.Merkle,
		Base:process.Hash,
		BaseIndex:process.Index,
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
		Type:      defines.EntryType_Proof,
		Data:      proofBytes,
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
func (p *Pot) broadcastRequestProcesses() error {
	req := &defines.Request{
		Type:       defines.RequestType_Processes,
	}
	p.pit.RangePeers(func(peer *defines.PeerInfo) error {
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Req,
			From:    p.id,
			To:      peer.Id,
			Sig:     nil,
			Reqs:    []*defines.Request{req},
		}

		err := msg.Sign()
		if err != nil {
			return err
		}

		p.send(msg)
		return nil
	})
	return nil
}

// 广播请求区块，追上最新进度
func (p *Pot) broadcastRequestBlocks() error {
	req := &defines.Request{
		Type:       defines.RequestType_Blocks,
		IndexStart: 0,
		IndexCount: 0,
		Hashes:     nil,
	}
	p.pit.RangePeers(func(peer *defines.PeerInfo) error {
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Req,
			From:    p.id,
			To:      peer.Id,
			Sig:     nil,
			Reqs:    []*defines.Request{req},
		}

		err := msg.Sign()
		if err != nil {
			return err
		}

		p.send(msg)
		return nil
	})
	return nil
}

// 暂时是直接向一个节点发送区块请求，但最好后面改成随机选举三个去请求
func (p *Pot) requestBlocks() {
	// TODO
}
