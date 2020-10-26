/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 7:10 PM
* @Description: 处理请求
***********************************************************************/

package pot

import (
	"errors"

	"github.com/azd1997/blockchain-consensus/defines"
)

// 将请求的区块返回回去
func (p *Pot) handleRequestBlocks(from string, req *defines.Request) error {
	var block *defines.Block
	var err error

	entries := make([]*defines.Entry, 0)

	if req.IndexStart >= 0 && req.IndexCount == 0 && len(req.Hashes) == 0 {		// 特殊情况：响应方应该将IndexStart之后所有区块返回
		maxIndex := p.bc.GetMaxIndex()
		if maxIndex < req.IndexStart {
			return errors.New("maxIndex < req.IndexStart")
		}

		var blocks []*defines.Block
		blocks, err = p.bc.GetBlocksByRange(req.IndexStart, req.IndexCount)
		if err != nil {
			p.Errorf("handleRequestBlocks: get block(%d-%d) fail: %s\n",
				req.IndexStart, req.IndexStart+req.IndexCount-1, err)
			return err
		}
		for _, block := range blocks {
			b, err := block.Encode()
			if err != nil {
				p.Errorf("handleRequestBlocks: encode block(%x) fail: %s\n", block.SelfHash, err)
				continue
			}
			entries = append(entries, &defines.Entry{
				BaseIndex: block.Index-1,
				Base:      block.PrevHash,
				Type:      defines.EntryType_Block,
				Data:      b,
			})
		}
	} else if req.IndexCount > 0 {		// 按Index请求

		var blocks []*defines.Block
		blocks, err = p.bc.GetBlocksByRange(req.IndexStart, req.IndexCount)
		if err != nil {
			p.Errorf("handleRequestBlocks: get block(%d-%d) fail: %s\n",
				req.IndexStart, req.IndexStart+req.IndexCount-1, err)
			return err
		}
		for _, block := range blocks {
			b, err := block.Encode()
			if err != nil {
				p.Errorf("handleRequestBlocks: encode block(%x) fail: %s\n", block.SelfHash, err)
				continue
			}
			entries = append(entries, &defines.Entry{
				BaseIndex: block.Index-1,
				Base:      block.PrevHash,
				Type:      defines.EntryType_Block,
				Data:      b,
			})
		}

	} else {	// 按Hash请求

		for _, h := range req.Hashes {
			block, err = p.bc.GetBlockByHash(h)
			if err != nil {
				p.Errorf("handleRequestBlocks: get block(%x) fail: %s\n", h, err)
				continue
			}
			b, err := block.Encode()
			if err != nil {
				p.Errorf("handleRequestBlocks: encode block(%x) fail: %s\n", h, err)
				continue
			}
			entries = append(entries, &defines.Entry{
				BaseIndex: block.Index-1,
				Base:      block.PrevHash,
				Type:      defines.EntryType_Block,
				Data:      b,
			})
		}
	}

	msg := &defines.Message{
		Version: defines.CodeVersion,
		Type:    defines.MessageType_Data,
		From:    p.id,
		To:      from,
		Sig:     nil,
		Entries: entries,
	}

	// 签名
	if err := msg.Sign(); err != nil {
		return err
	}

	// 发送
	p.send(msg)

	return nil
}

// 将自身的邻居表整理回发
//
func (p *Pot) handleRequestNeighbors(from string, req *defines.Request) error {

	// TODO: seed考虑From的广播

	// 回发节点表
	entries := make([]*defines.Entry, 0)
	p.pit.RangePeers(func(peer *defines.PeerInfo) error {
		b, err := peer.Encode()
		if err != nil {
			p.Errorf("handleRequestNeighbors: encode peerinfo(%v) fail: %s\n", *peer, err)
			return err
		}
		entries = append(entries, &defines.Entry{
			Type:      defines.EntryType_Neighbor,
			Data:      b,
		})
		return nil
	})

	msg := &defines.Message{
		Version: defines.CodeVersion,
		Type:    defines.MessageType_Data,
		From:    p.id,
		To:      from,
		Entries: entries,
	}

	// 签名
	if err := msg.Sign(); err != nil {
		return err
	}

	// 发送
	p.send(msg)

	return nil
}


/////////////////////////////////////////////////

// 向节点回应其请求的区块
func (p *Pot) responseBlocks(to string, blocks ...*defines.Block) error {
	l := len(blocks)
	if l == 0 {
		return nil
	}

	blockBytes := make([][]byte, l)
	for i:=0; i<l; i++ {
		b, err := blocks[i].Encode()
		if err != nil {
			//p.Errorf("response blocks fail: %s\n", err)
			return err
		}
		blockBytes[i] = b
	}

	entries := make([]*defines.Entry, l)
	for i := 0; i < l; i++ {
		entries[i] = &defines.Entry{
			BaseIndex: blocks[i].Index - 1,
			Base:      blocks[i].PrevHash,
			Type:      defines.EntryType_Block,
			Data:      blockBytes[i],
		}
	}

	msg := &defines.Message{
		Version: defines.CodeVersion,
		Type:    defines.MessageType_Data,
		From:    p.id,
		To:      to,
		Entries: entries,
	}
	err := msg.Sign()
	if err != nil {
		// p.Errorf("response blocks fail: %s\n", err)
		return err
	}

	p.send(msg)
	return nil
}


