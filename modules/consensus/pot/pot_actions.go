/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 5:17 PM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"errors"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
)

// send 指向网络中（或者说外部依赖的网络模块）发送消息。 注意本地消息不要通过该方法使用
// 这个发送消息可以是单播也可以是多播，具体看
// 注意：send可能发生较长时间阻塞，调用时使用go send()
func (p *Pot) send(msg *defines.Message) error {
	to, err := p.pit.Get(msg.To)
	if err != nil {
		return err
	}

	// 由于本地Socket通信很快，所以为了模拟实际通信延时，这里随机睡眠一段时间
	randMs := rand.Intn(20) + 10 //   10~30ms
	time.Sleep(time.Duration(randMs) * time.Millisecond)

	return p.net.Send(msg.To, to.Addr, msg)
}

// signAndSendMsg 酌情使用 go signAndSendMsg()
func (p *Pot) signAndSendMsg(msg *defines.Message) error {
	if msg == nil {
		return errors.New("nil msg")
	}

	// 为所有msg填充
	msg.Version = defines.CodeVersion
	msg.Epoch = p.bc.GetMaxIndex()

	// 签名
	err := msg.Sign()
	if err != nil {
		return err
	}

	return p.send(msg)
}

// broadcastMsg 将msg广播给所有种子节点和共识节点
// 需要对传入的msg设置To
// 返回值total为需要发信的总数，succ为成功发信的数量，errs为返回的错误
// toseeds指定是否向seeds广播；topeers指定是否向peers广播
// TODO: 这里广播是并发的，但是广播的调用支出是否要go呢？如果同步操作的话还是需要一些时间的
func (p *Pot) broadcastMsg(msg *defines.Message, toseeds, topeers bool) (
	int, int, map[string]error) {

	// 由于total,succ,errs要在并发环境更改，所以需要进行保护
	total, succ := int32(0), int32(0)
	errs := make(map[string]error)
	errsLock := new(sync.Mutex)

	wg := new(sync.WaitGroup)

	f := func(msg *defines.Message, to string) {
		defer wg.Done()

		atomic.AddInt32(&total, 1)
		if err := p.signAndSendMsg(msg); err != nil {
			p.Errorf("broadcastMsg(%s): to %s fail: %v", msg.Type.String(), to, err)
			errsLock.Lock()
			errs[to] = err
			errsLock.Unlock()
		} else {
			p.Debugf("broadcastMsg(%s): to %s", msg.Type.String(), to)
			atomic.AddInt32(&succ, 1)
		}
	}

	if toseeds {
		seeds := p.pit.Seeds()
		for seed := range seeds {
			if seed != p.id {
				// 这里要注意，由于是并发的，所以msg必须复制避免data race
				// 但是由于只是更改了To，其他成员都是并发读，所以没必要深拷贝
				m := *msg
				m.To = seed
				wg.Add(1)
				go f(&m, seed)
			}
		}
	}

	if topeers {
		peers := p.pit.Peers()
		for peer := range peers {
			if peer != p.id {
				// 这里要注意，由于是并发的，所以msg必须复制避免data race
				// 但是由于只是更改了To，其他成员都是并发读，所以没必要深拷贝
				m := *msg
				m.To = peer
				wg.Add(1)
				go f(&m, peer)
			}
		}
	}

	wg.Wait()
	return int(atomic.LoadInt32(&total)), int(atomic.LoadInt32(&succ)), errs
}

// broadcastTx 将tx广播给所有种子节点和共识节点
func (p *Pot) broadcastTxs(txs ...*defines.Transaction) error {

	// 编码
	txBytes := make([][]byte, len(txs))
	for i := 0; i < len(txs); i++ {
		if txs[i] != nil {
			enced, err := txs[i].Encode()
			if err != nil {
				return err
			}
			txBytes[i] = enced
		}
	}

	// msg模板
	msg := &defines.Message{
		Version: defines.CodeVersion,
		Type:    defines.MessageType_Txs,
		From:    p.id,
		To:      "",
		Data:    txBytes,
	}
	if err := msg.WriteDesc("type", "txs"); err != nil {
		return err
	}

	// 广播
	if _, _, errs := p.broadcastMsg(msg, true, true); len(errs) > 0 {
		p.Errorf("broadcastMsg fail %d", len(errs))
	}

	return nil
}

// broadcastSelfProof 广播自己的证明
// cheat则意味着虚报很高的交易量
func (p *Pot) broadcastSelfProof(cheat bool) error {
	// 首先需要获取证明
	nb, err := p.bc.GenNextBlock()
	if err != nil {
		return err
	}
	p.maybeNewBlock = nb // 自己的区块设为可能的新区块

	proof := &Proof{
		Id:        p.id,
		TxsNum:    int64(len(nb.Txs)),
		BlockHash: nb.SelfHash,
		Base:      nb.PrevHash,
		BaseIndex: nb.Index - 1,
	}

	if cheat {
		proof.TxsNum = 99999999
		p.Infof("cheat when broadcast self proof(%s), txsnum set to 99999999", proof.Short())
	}

	// 添加到自己的proofs
	p.proofs.Add(proof)

	return p.broadcastProof(proof, true, true)
}

// broadcastProof 广播证明
// 目前的方案中，是seed节点用来广播winner的Proof的
func (p *Pot) broadcastProof(proof *Proof, toseeds, topeers bool) error {
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

	// 消息模板
	msg := &defines.Message{
		Version: defines.CodeVersion,
		Type:    msgtype,
		From:    p.id,
		To:      "",
		Data:    [][]byte{proofBytes},
	}
	if err := msg.WriteDesc("type", desc); err != nil {
		return err
	}

	// 广播
	if _, _, errs := p.broadcastMsg(msg, toseeds, topeers); len(errs) > 0 {
		p.Errorf("broadcastMsg fail %d", len(errs))
	}

	return nil
}

// broadcastPeers 广播节点信息
func (p *Pot) broadcastPeers(pis []*defines.PeerInfo, pisb [][]byte, toseeds, topeers bool) error {
	// 编码
	piBytes := make([][]byte, len(pis))
	for i := 0; i < len(pis); i++ {
		if pis[i] != nil {
			pib, err := pis[i].Encode()
			if err != nil {
				return err
			}
			piBytes[i] = pib
		}
	}

	// 消息模板
	msg := &defines.Message{
		Version: defines.CodeVersion,
		Type:    defines.MessageType_Peers,
		From:    p.id,
	}
	if len(piBytes) > 0 {
		msg.Data = piBytes
	} else if len(pisb) > 0 {
		msg.Data = pisb
	} else {
		return errors.New("no pitable to broadcast")
	}
	if err := msg.WriteDesc("type", "peers"); err != nil {
		return err
	}

	// 广播
	if _, _, errs := p.broadcastMsg(msg, toseeds, topeers); len(errs) > 0 {
		p.Errorf("broadcastMsg fail %d", len(errs))
	}

	return nil
}

// broadcastNewBlock 广播新区块
func (p *Pot) broadcastNewBlock(nb *defines.Block) error {

	// 编码
	blockBytes, err := nb.Encode()
	if err != nil {
		return err
	}

	// msg模板
	msg := &defines.Message{
		Version:   defines.CodeVersion,
		Type:      defines.MessageType_NewBlock,
		From:      p.id,
		To:        "",
		Base:      nb.PrevHash,
		BaseIndex: nb.Index - 1,
		Data:      [][]byte{blockBytes},
	}
	if err := msg.WriteDesc("type", "newblock"); err != nil {
		return err
	}

	// 广播
	if _, _, errs := p.broadcastMsg(msg, true, true); len(errs) > 0 {
		p.Errorf("broadcastMsg fail %d", len(errs))
	}

	return nil
}

// broadcastRequestNeighbors 广播getNeighbors请求
// toseeds true则向种子节点广播；否则向所有节点广播
func (p *Pot) broadcastRequestNeighbors(toseeds, topeers bool) (total int, succ int, err error) {
	// 查询自身节点信息
	self, err := p.pit.Get(p.id)
	if err != nil {
		return 0, 0, err
	}
	selfb, err := self.Encode()
	if err != nil {
		return 0, 0, err
	}

	// msg模板
	msg := &defines.Message{
		Version: defines.CodeVersion,
		Type:    defines.MessageType_ReqPeers,
		From:    p.id,
		To:      "",
		Data:    [][]byte{selfb}, // 自己的节点信息
	}
	if err := msg.WriteDesc("type", "req-peers"); err != nil {
		return 0, 0, err
	}

	// 广播
	total, succ, errs := p.broadcastMsg(msg, toseeds, topeers)
	if succ <= 0 {
		err = errors.New("broadcast msg all failed")
		p.Errorf("broadcastMsg all fail. total=%d, succ=%d, errs=%v", total, succ, errs)
		return total, succ, err
	}
	if len(errs) > 0 {
		p.Warnf("broadcastMsg par fail. total=%d, succ=%d, errs=%v", total, succ, errs)
	}
	return total, succ, nil
}

// broadcastRequestBlocksByIndex
func (p *Pot) broadcastRequestBlocksByIndex(start int64, count int64, toseeds, topeers bool) (total, succ int, err error) {

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
	if err := req.WriteDesc("type", "req-blockbyindex"); err != nil {
		return 0, 0, err
	}

	// 广播
	total, succ, errs := p.broadcastMsg(req, toseeds, topeers)
	if succ <= 0 {
		err = errors.New("broadcast msg all failed")
		p.Errorf("broadcastMsg all fail. total=%d, succ=%d, errs=%v", total, succ, errs)
		return total, succ, err
	}
	if len(errs) > 0 {
		p.Warnf("broadcastMsg par fail. total=%d, succ=%d, errs=%v", total, succ, errs)
	}
	return total, succ, nil
}

// broadcastHeartbeat 广播心跳消息
func (p *Pot) broadcastHeartbeat(t time.Time) error {
	// 查询自身节点信息
	self, err := p.pit.Get(p.id)
	if err != nil {
		return err
	}
	selfb, err := self.Encode()
	if err != nil {
		return err
	}

	// msg模板
	msg := &defines.Message{
		Version: defines.CodeVersion,
		Type:    defines.MessageType_ReqPeers,
		From:    p.id,
		To:      "",
		Data:    [][]byte{selfb}, // 自己的节点信息
	}
	if err := msg.WriteDesc("type", "heartbeat"); err != nil {
		return err
	}

	// 广播
	total, succ, errs := p.broadcastMsg(msg, true, true)
	if succ <= 0 {
		err = errors.New("broadcast msg all failed")
		p.Errorf("broadcastMsg all fail. total=%d, succ=%d, errs=%v", total, succ, errs)
		return err
	}
	if len(errs) > 0 {
		p.Warnf("broadcastMsg part fail. total=%d, succ=%d, errs=%v", total, succ, errs)
	}
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

// waitAndDecideOneBlock 等待某个区块，需要在wait阶段决定哪个才是正确的
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
