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

	var err error
	var blocks []*defines.Block

	if req.IndexStart >= 0 && req.IndexCount == 0 && len(req.Hashes) == 0 { // 特殊情况：响应方应该将IndexStart之后所有区块返回
		maxIndex := p.bc.GetMaxIndex()
		if maxIndex < uint64(req.IndexStart) {
			return errors.New("maxIndex < req.IndexStart")
		}
		blocks, err = p.bc.GetBlocksByRange(uint64(req.IndexStart), maxIndex)
	} else if req.IndexCount > 0 { // 按Index请求
		blocks, err = p.bc.GetBlocksByRange(uint64(req.IndexStart), uint64(req.IndexCount)) //TODO
	} else { // 按Hash请求
		blocks, err = p.bc.GetBlocksByHashes(req.Hashes)
	}

	if err != nil {
		p.Errorf("handleRequestBlocks: get blocks from blockchain(p.bc) for %s fail: %s\n", from, err)
		return err
	}
	p.Errorf("handleRequestBlocks: get blocks from blockchain(p.bc) for %s \n", from)
	return p.responseBlocks(from, blocks...)
}

// 将自身的邻居表整理回发
// 只有seed节点才处理该消息
func (p *Pot) handleRequestNeighbors(from string, req *defines.Request) error {

	if p.duty != defines.PeerDuty_Seed {
		return nil
	}

	// 回发节点表
	entries := make([]*defines.Entry, 0)
	p.pit.RangePeers(func(peer *defines.PeerInfo) error {
		b, err := peer.Encode()
		if err != nil {
			p.Errorf("handleRequestNeighbors: encode peerinfo(%v) fail: %s\n", *peer, err)
			return err
		}
		entries = append(entries, &defines.Entry{
			Type: defines.EntryType_Neighbor,
			Data: b,
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

	err := p.signAndSendMsg(msg)
	if err != nil {
		return err
	}

	// seed节点将该节点信息广播给所有peer节点
	if _, err := p.pit.Get(from); err == nil { // 如果已经有该节点的地址信息，则直接返回，无须继续广播
		return nil
	}
	bEntry := &defines.Entry{
		Type: defines.EntryType_Neighbor,
		Data: req.Data, // 请求方的节点信息
	}
	p.pit.RangePeers(func(peer *defines.PeerInfo) error {
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Data,
			From:    p.id,
			To:      peer.Id,
			Entries: []*defines.Entry{bEntry},
		}
		return p.signAndSendMsg(msg)
	})

	return nil
}

// 将自身的
func (p *Pot) handleRequestProcesses(from string, req *defines.Request) error {

	// 回发节点表
	entries := make([]*defines.Entry, 0)
	p.pit.RangePeers(func(peer *defines.PeerInfo) error {
		b, err := peer.Encode()
		if err != nil {
			p.Errorf("handleRequestNeighbors: encode peerinfo(%v) fail: %s\n", *peer, err)
			return err
		}
		entries = append(entries, &defines.Entry{
			Type: defines.EntryType_Neighbor,
			Data: b,
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

	return p.signAndSendMsg(msg)
}

/////////////////////////////////////////////////

// 向节点回应其请求的区块
func (p *Pot) responseBlocks(to string, blocks ...*defines.Block) error {
	l := len(blocks)
	if l == 0 {
		return nil
	}

	blockBytes := make([][]byte, l)
	for i := 0; i < l; i++ {
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
	return p.signAndSendMsg(msg)
}
