/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/18/20 9:47 AM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"github.com/azd1997/blockchain-consensus/defines"
)

// requestNeighborsFuncGenerator 根据当前pot状态机本地状态，生成用来请求邻居节点的函数
func (p *Pot) requestNeighborsFuncGenerator() func(peer *defines.PeerInfo) error {
	selfPeerInfo, err := p.pit.Get(p.id)
	if err != nil {
		return nil
	}
	spib, err := selfPeerInfo.Encode()
	if err != nil {
		return nil
	}

	rnf := func(peer *defines.PeerInfo) error {
		if peer.Id == p.id {
			return nil
		}
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Req,
			Epoch:   p.epoch, // 会变化
			From:    p.id,
			To:      peer.Id,
			Reqs: []*defines.Request{&defines.Request{
				Type: defines.RequestType_Neighbors,
				Data: spib,
			}},
		}
		return p.signAndSendMsg(msg)
	}
	return rnf
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
			Type:    defines.MessageType_Req,
			Epoch:   p.epoch, // 会变化
			From:    p.id,
			To:      peer.Id,
			Reqs: []*defines.Request{&defines.Request{
				Type:       defines.RequestType_Blocks,
				IndexStart: index,
				IndexCount: 1,
			}},
		}
		return p.signAndSendMsg(msg)
	}
	return rnf
}
