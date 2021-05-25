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

// handleMsgBlocks 处理MsgBlocks
func (p *Pot) handleMsgBlocks(msg *defines.Message) error {
	stage := p.getStage()

	// 1. RN阶段忽略
	if stage == StageType_PreInited_RequestNeighbors {
		return nil
	}

	// 2. RFB阶段只处理b1
	if stage == StageType_PreInited_RequestFirstBlock {
		if len(msg.Data) == 0 {
			return errors.New("empty ")
		}
		fb := new(defines.Block)
		if err := fb.Decode(msg.Data[0]); err != nil {
			return err
		}
		if fb.Index != 1 {
			return errors.New("err")
		}

		// 对b作检查
		// TODO

		if p.nWaitBlockChan != nil {
			p.nWaitBlockChan <- fb // 通知收到一个节点回传了1号区块
			p.Debugf("%s handle MsgBlock from (%s) succ", p.DutyStageState(), msg.From)
		}
		// 这个firstBlock由启动逻辑确定之后再写到本地

		//if err := p.bc.AddBlock(fb); err != nil {
		//	p.Errorf("p.bc.AddBlock(%s) fail. err=%s,b=%v", fb.ShortName(), err, fb)
		//}
		return nil
	}

	// 3. 其他阶段正常接收所有区块
	for _, bb := range msg.Data {
		b := new(defines.Block)
		if err := b.Decode(bb); err != nil {
			return err
		}
		// 对b作检查
		// TODO

		if err := p.bc.AddBlock(b); err != nil {
			p.Errorf("p.bc.AddBlock(%s) fail. err=%s,b=%v", b.ShortName(), err, b)
		}
	}

	return nil
}

// handleMsgNewBlock
func (p *Pot) handleMsgNewBlock(msg *defines.Message) error {
	stage := p.getStage()

	// 1. RN/RFB/RLB阶段忽略
	if stage == StageType_PreInited_RequestNeighbors ||
		stage == StageType_PreInited_RequestFirstBlock {
		return nil
	}

	// 2. 其他阶段，则进行接收,添加到udbt中
	if len(msg.Data) == 0 {
		return errors.New("empty ")
	}
	nb := new(defines.Block)
	if err := nb.Decode(msg.Data[0]); err != nil {
		return err
	}
	p.udbt.Add(nb)
	return nil
}

// handleMsgTxs
func (p *Pot) handleMsgTxs(msg *defines.Message) error {
	stage := p.getStage()

	// 1. 非peer忽略，RN/RFB/RLB阶段忽略
	if p.duty != defines.PeerDuty_Peer ||
		stage == StageType_PreInited_RequestNeighbors ||
		stage == StageType_PreInited_RequestFirstBlock {
		return nil
	}

	// 2. 其他阶段，则进行接收
	for _, tb := range msg.Data {
		tx := new(defines.Transaction)
		if err := tx.Decode(tb); err != nil {
			return err
		}
		// 检查Tx有无问题

		// 加入到BC中
		if err := p.bc.AddTx(tx); err != nil {
			p.Errorf("AddTx %s fail. err=%s", tx.Key(), err)
		}
	}

	return nil
}

// handleMsgLocalTxs 本机制造的交易要发送出去
func (p *Pot) handleMsgLocalTxs(msg *defines.Message) error {

	// 解码存到本地
	for _, tb := range msg.Data {
		tx := new(defines.Transaction)
		if err := tx.Decode(tb); err != nil {
			return err
		}
		// 检查Tx有无问题

		// 加入到本地BC中
		if err := p.bc.AddTx(tx); err != nil {
			p.Errorf("AddLocalTx %s fail. err=%s", tx.Key(), err)
		}

		// 上报
		if err := p.ReportNewTx(tx); err != nil {
			p.Errorf("ReportNewTx fail: %s", err)
		}
	}

	// 广播给其他节点
	msg.Type = defines.MessageType_Txs
	if _, _, errs := p.broadcastMsg(msg, true, true); len(errs) > 0 {
		p.Errorf("broadcastMsg fail %d", len(errs))
	}
	return nil
}

// handleMsgPeers
func (p *Pot) handleMsgPeers(msg *defines.Message) error {

	if p.getStage() == StageType_PreInited_RequestNeighbors {
		count := 0
		for _, pb := range msg.Data {
			count++
			pi := new(defines.PeerInfo)
			if err := pi.Decode(pb); err != nil {
				return err
			}
			p.pit.Set(pi)
		}
		if count > 0 && p.nWaitChan != nil { // 说明包含Neighbors
			p.nWaitChan <- 1 // 通知收到一个节点回传了节点信息表
		}
		return nil
	}

	// 任何情况下都是直接加入本地
	// 除非已在本地黑名单
	for _, pb := range msg.Data {
		pi := new(defines.PeerInfo)
		if err := pi.Decode(pb); err != nil {
			return err
		}
		p.pit.Set(pi)
	}

	return nil
}

// handleMsgProof
func (p *Pot) handleMsgProof(msg *defines.Message) error {
	stage := p.getStage()
	if stage != StageType_InPot && stage != StageType_PostPot {
		return nil
	}

	if len(msg.Data) == 0 {
		return errors.New("err")
	}

	proof := new(Proof)
	if err := proof.Decode(msg.Data[0]); err != nil {
		return err
	}

	// 1. 只接受来自peer本人的proof
	if stage == StageType_InPot &&
		proof.Id == msg.From &&
		p.pit.IsPeer(msg.From) {
		p.proofs.Add(proof)
		return nil
	}
	// 2. 只接受来自种子节点转发的proof
	if stage == StageType_PostPot &&
		proof.Id != msg.From &&
		p.pit.IsPeer(proof.Id) &&
		p.pit.IsSeed(msg.From) {
		p.proofs.AddProofRelayedBySeed(proof)
		return nil
	}

	return nil
}

// handleMsgReqBlockByIndex 必须本地是ready状态才可以回复区块请求
func (p *Pot) handleMsgReqBlockByIndex(msg *defines.Message) error {
	stage := p.getStage()
	if stage != StageType_InPot && stage != StageType_PostPot {
		return nil
	}

	// TODO
	p.Debugf("handleMsgReqBlockByIndex: msg=%s", msg)

	// 查询
	bs, err := p.bc.GetBlocksByRange(msg.ReqBlockIndexStart, msg.ReqBlockIndexCount)
	if err != nil {
		return err
	}

	blockBytes := make([][]byte, len(bs))
	for i := 0; i < len(bs); i++ {
		blockByte, err := bs[i].Encode()
		if err != nil {
			continue
		}
		blockBytes[i] = blockByte
	}

	// 回复
	reply := &defines.Message{
		Version: defines.CodeVersion,
		Type:    defines.MessageType_Blocks,
		Epoch:   p.epoch,
		From:    p.id,
		To:      msg.From,
		Data:    blockBytes,
		Desc:    "",
	}
	if err := msg.WriteDesc("type", "rsp-blocks"); err != nil {
		return err
	}
	if err := p.signAndSendMsg(reply); err != nil {
		return err
	}

	return nil
}

// handleMsgReqBlockByHash 必须本地是ready状态才可以回复区块请求
func (p *Pot) handleMsgReqBlockByHash(msg *defines.Message) error {
	stage := p.getStage()
	if stage != StageType_InPot && stage != StageType_PostPot {
		return nil
	}

	// TODO

	p.Debugf("handleMsgReqBlockByHash: msg=%s", msg)

	// 查询
	bs, err := p.bc.GetBlocksByHashes(msg.Hashes)
	if err != nil {
		return err
	}

	blockBytes := make([][]byte, len(bs))
	for i := 0; i < len(bs); i++ {
		blockByte, err := bs[i].Encode()
		if err != nil {
			continue
		}
		blockBytes[i] = blockByte
	}

	// 回复
	reply := &defines.Message{
		Version: defines.CodeVersion,
		Type:    defines.MessageType_Blocks,
		Epoch:   p.epoch,
		From:    p.id,
		To:      msg.From,
		Data:    blockBytes,
		Desc:    "",
	}
	if err := msg.WriteDesc("type", "rsp-blocks"); err != nil {
		return err
	}
	if err := p.signAndSendMsg(reply); err != nil {
		return err
	}

	return nil
}

// handleMsgReqPeers
func (p *Pot) handleMsgReqPeers(msg *defines.Message) error {

	// TODO 条件限制

	// 验证msg.From与msg.Data[0]携带的peerinfo是否匹配

	p.Debugf("handleMsgReqPeers: msg=%s", msg)

	// 查询
	pis := p.pit.Peers()

	piBytes := make([][]byte, 0, len(pis))
	for _, pi := range pis {
		piByte, err := pi.Encode()
		if err != nil {
			continue
		}
		piBytes = append(piBytes, piByte)
	}

	p.Debugf("handleMsgReqPeers: 1111, msg=%s", msg)

	// 回复
	reply := &defines.Message{
		Version: defines.CodeVersion,
		Type:    defines.MessageType_Peers,
		Epoch:   p.epoch,
		From:    p.id,
		To:      msg.From,
		Data:    piBytes,
		Desc:    "",
	}
	if err := msg.WriteDesc("type", "rsp-peers"); err != nil {
		return err
	}
	if err := p.signAndSendMsg(reply); err != nil {
		return err
	}

	p.Debugf("handleMsgReqPeers: 2222, msg=%s", msg)

	// 将msg.Data[0]携带的peerinfo加入到本地，并转发给其他节点
	if _, err := p.pit.Get(msg.From); err == nil { // 如果已经有该节点的地址信息，则直接返回，无须继续广播
		return nil
	}
	if err := p.broadcastPeers(nil, msg.Data, true, true); err != nil {
		return err
	}

	p.Debugf("handleMsgReqPeers: 3333, msg=%s", msg)

	return nil
}
