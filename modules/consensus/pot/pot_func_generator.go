/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/18/20 9:47 AM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"github.com/azd1997/blockchain-consensus/defines"
)

// rangePitFunc 遍历Pit的函数
type rangePitFunc func(peer *defines.PeerInfo) error

// requestNeighborsFuncGenerator 根据当前pot状态机本地状态，生成用来请求邻居节点的函数
func (p *Pot) requestNeighborsFuncGenerator() (rangePitFunc, error) {

	selfPeerInfo, err := p.pit.Get(p.id)
	if err != nil {
		return nil, err
	}
	spib, err := selfPeerInfo.Encode()
	if err != nil {
		return nil, err
	}

	rnf := func(peer *defines.PeerInfo) error {
		if peer.Id == p.id {
			return nil
		}
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_ReqPeers,
			Epoch:   p.epoch, // 会变化
			From:    p.id,
			To:      peer.Id,
			Data: [][]byte{spib},
		}
		if err := msg.WriteDesc("type", "req-peers"); err != nil {
			return err
		}
		if err := p.signAndSendMsg(msg); err != nil {
			p.Errorf("requestNeighbors: to %s fail: %v", peer.Id, err)
			return err
		} else {
			p.Debugf("requestNeighbors: to %s", peer.Id)
		}
		return nil
	}
	return rnf, nil
}

// requestLatestBlockFuncGenerator 根据当前pot状态机本地状态，生成用来请求最新区块的函数
func (p *Pot) requestLatestBlockFuncGenerator() func(peer *defines.PeerInfo) error {
	return p.requestOneBlockFuncGenerator(-1)
}

// requestOneBlockFuncGenerator 根据当前pot状态机本地状态，生成用来请求某个区块的函数
// index为负数时反向索引; index为0不存在
func (p *Pot) requestOneBlockFuncGenerator(index int64) func(peer *defines.PeerInfo) error {
	rnf := func(peer *defines.PeerInfo) error {
		if peer.Id == p.id {
			return nil
		}
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_ReqBlockByIndex,
			Epoch:   p.epoch, // 会变化
			From:    p.id,
			To:      peer.Id,
			ReqBlockIndexStart:index,
			ReqBlockIndexCount:1,
		}
		if err := msg.WriteDesc("type", "req-block"); err != nil {
			return err
		}
		if err := p.signAndSendMsg(msg); err != nil {
			p.Errorf("requestBlock: to %s fail: %v", peer.Id, err)
			return err
		} else {
			p.Debugf("requestBlock: to %s", peer.Id)
		}
		return nil
	}
	return rnf
}
