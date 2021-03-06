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

	if req.IndexStart >= 1 && req.IndexCount == 0 && len(req.Hashes) == 0 { // 特殊情况：响应方应该将IndexStart之后所有区块返回
		maxIndex := p.bc.GetMaxIndex()
		if maxIndex < req.IndexStart {
			return errors.New("maxIndex < req.IndexStart")
		}
		blocks, err = p.bc.GetBlocksByRange(req.IndexStart, maxIndex)
	} else if req.IndexStart == -1 { // start为-1，代表反向获取区块，例如start=-1,count=1表示把最新区块返回
		maxIndex := p.bc.GetMaxIndex()
		start := int64(0)
		if maxIndex+1 > req.IndexCount {
			start = maxIndex - req.IndexCount + 1
		}
		blocks, err = p.bc.GetBlocksByRange(start, maxIndex)
	} else if req.IndexCount > 0 { // 按Index请求
		blocks, err = p.bc.GetBlocksByRange(req.IndexStart, req.IndexCount) //TODO
	} else { // 按Hash请求
		blocks, err = p.bc.GetBlocksByHashes(req.Hashes)
	}

	if err != nil {
		p.Errorf("handleRequestBlocks: get blocks from blockchain(p.bc) for %s fail: %s", from, err)
		return err
	}
	p.Debugf("handleRequestBlocks: get blocks from blockchain(p.bc) for %s, blocks=%v", from, blocks)
	p.Debugf("current BC: %s", p.bc.Display())
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
			p.Errorf("handleRequestNeighbors: encode peerinfo(%v) fail: %s", *peer, err)
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
	if err := msg.WriteDesc("type", "rsp-neighbors"); err != nil {
		return err
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
		if err := msg.WriteDesc("type", "neighbor"); err != nil {
			return err
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
			p.Errorf("handleRequestNeighbors: encode peerinfo(%v) fail: %s", *peer, err)
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
	if err := msg.WriteDesc("type", "rsp-processes"); err != nil {
		return err
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
	if err := msg.WriteDesc("type", "rsp-blocks"); err != nil {
		return err
	}
	return p.signAndSendMsg(msg)
}
