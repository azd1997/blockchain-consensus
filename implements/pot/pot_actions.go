/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 5:17 PM
* @Description: The file is for
***********************************************************************/

package pot

import "github.com/azd1997/blockchain-consensus/defines"

// send指向网络中（或者说外部依赖的网络模块）发送消息。 注意本地消息不要通过该方法使用
// 这个发送消息可以是单播也可以是多播，具体看
func (p *Pot) send(msg *defines.Message) {
	p.msgout <- msg
}

//// broadcast
//func (n *Node) broadcast(msg *defines.Message) {
//
//}

// 广播getNeighbors请求
func (p *Pot) broadcastRequestNeighbors() error {
	nWait := 0
	req := &defines.Request{
		Type:       defines.RequestType_Neighbors,
	}
	f := func(peer *defines.PeerInfo) error {
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Req,
			From:    p.id,
			To:      peer.Id,
			Reqs:    []*defines.Request{req},
		}
		if err := msg.Sign(); err != nil {
			return err
		}
		nWait++
		p.send(msg)
		return nil
	}
	p.pit.RangePeers(f)
	p.pit.RangeSeeds(f)
	return nil
}

// 广播证明
func (p *Pot) broadcastProof() error {
	// 首先需要获取证明
	nb := p.txPool.GenBlock()
	p.maybeNewBlock = nb
	proof := &Proof{
		TxsNum:    uint64(len(nb.Txs)),
		TxsMerkle: nb.Merkle,
	}

	proofBytes, err := proof.Encode()
	if err != nil {
		return err
	}
	entry := &defines.Entry{
		Base:      nb.PrevHash,
		BaseIndex: nb.Index - 1,
		Type:      defines.EntryType_Proof,
		Data:      proofBytes,
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
