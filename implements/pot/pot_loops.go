/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 5:18 PM
* @Description: Pot的工作循环函数
***********************************************************************/

package pot

import "sync"

// 状态切换循环
// 这里要注意这个循环启动的前提是同步到了最新进度，拿到了世界时钟之后才进入状态切换循环
func (p *Pot) stateMachineLoop() {

	once := new(sync.Once)

	for {
		select {
		case <-p.done:
			p.Infof("stateMachineLoop: return ...\n")
			return
		case moment := <-p.clock.Tick:
			p.Infof("stateMachineLoop: clock tick: %s\n", moment.String())
			p.Info(p.bc.Display())

			// 根据当前状态来处理此滴答消息
			state := p.getState()
			switch state {

			case StateType_PreInited_RequestNeighbors: // nothing
			case StateType_PreInited_RequestFirstBlock:
				// 通过该chan向启动逻辑传递时刻信号
				if moment.Type == MomentType_PotStart {
					once.Do(func() {
						p.potStartBeforeReady <- moment
					})
				}
			case StateType_PreInited_RequestLatestBlock: // nothing
			case StateType_NotReady:
				// 能够处理邻居消息,区块消息,最新区块消息

				p.Debugf("current is ready? %v", !p.bc.Discontinuous())

				// 检查是否满足切换Ready的条件
				if !p.bc.Discontinuous() && moment.Type == MomentType_PotStart {
					p.setState(StateType_InPot)
					p.Infof("switch state from %s to %s\n", StateType_NotReady, StateType_InPot)
					p.startPot(moment)
				} else {
					// 还没满足切换Ready状态的条件，暂时不能收集证明消息(有一定但没有全部判别能力)，
					// 只能在pot竞赛结束后听从seed和winner
					if moment.Type == MomentType_PotStart {
						// 啥也不干，收集证明消息.   或者不收集？
						p.proofs.Reset()
					} else if moment.Type == MomentType_PotOver {
						// 开始接收新区块和seed广播的winner
					}
				}
			case StateType_InPot:
				if moment.Type == MomentType_PotOver { // 正常情况应该是PotOver时刻到来
					p.setState(StateType_PostPot)
					p.Infof("switch state from %s to %s\n", StateType_InPot, StateType_PostPot)
					// 汇总已收集的证明消息，决出胜者，判断自己是否出块，接下去等待胜者区块和seed广播的胜者证明
					p.endPot(moment)
				} else if moment.Type == MomentType_PotStart { // 不可能出现的错误
					p.Errorf("stateMachineLoop: Moment(%s) comes at StateInPot", moment)
				}

			case StateType_PostPot:
				if moment.Type == MomentType_PotStart { // 正常情况应该是PotOver时刻到来
					p.setState(StateType_InPot)
					p.Infof("switch state from %s to %s", StateType_PostPot, StateType_InPot)
					// 决定出新区块
					p.decide(moment)
					// 汇总已收集的证明消息，决出胜者，判断自己是否出块，接下去等待胜者区块和seed广播的胜者证明
					p.startPot(moment)
				} else if moment.Type == MomentType_PotOver { // 不可能出现的错误
					p.Errorf("stateMachineLoop: Moment(%s) comes at StatePostPot", moment)
				}
			default:
				p.Fatalf("stateMachineLoop: Moment(%s) comes at UnknownState(%s)", moment, state.String())

				//case StateType_NotReady:
				//	// 没准备好，啥也不干，等区块链同步
				//
				//	// 如果追上进度了则切换状态为ReadyCompete
				//	if p.latest() {
				//		p.setState(StateType_ReadyCompete)
				//	} else {
				//		// 否则请求快照数据
				//		p.broadcastRequestBlocks(true)
				//	}
				//
				//case StateType_ReadyCompete:
				//	// 当前是ReadyCompete，则状态切换为Competing
				//
				//	// 状态切换
				//	p.setState(StateType_Competing)
				//	// 发起竞争（广播证明消息）
				//	p.broadcastSelfProof()
				//
				//case StateType_Competing:
				//	// 当前是Competing，则状态切换为CompetingEnd，并判断竞赛结果，将状态迅速切换为Winner或Loser
				//
				//	// 状态切换
				//	p.setState(StateType_CompeteOver)
				//	// 判断竞赛结果，状态切换
				//	if p.winner == p.id { // 自己胜出
				//		p.setState(StateType_CompeteWinner)
				//		// 广播新区块
				//		p.broadcastNewBlock(p.maybeNewBlock)
				//	} else { // 别人胜出
				//		p.setState(StateType_CompeteLoser)
				//		// 等待新区块(“逻辑上”的等待，代码中并不需要wait)
				//	}
				//
				//case StateType_CompeteOver:
				//	// 正常来说，tick时不会是恰好CompeteOver而又没确定是Winner/Loser
				//	// 所以暂时无视
				//case StateType_CompeteWinner:
				//	// Winner来说的话，立马广播新区块，广播结束后即切换为Ready
				//	// 所以不太可能tick时状态为Winner
				//	// 暂时无视
				//case StateType_CompeteLoser:
				//	// Loser等待新区块，接收到tick说明还没得到新区块
				//	// 状态切换为Ready
				//
				//	p.setState(StateType_ReadyCompete)
				//	// 发起竞争（广播证明消息）
				//	p.broadcastSelfProof()
			}
		}
	}
}

// 消息处理循环
func (p *Pot) msgHandleLoop() {
	var err error
	for {
		select {
		case <-p.done:
			p.Infof("msgHandleLoop: return ...")
			return
		case msg := <-p.msgin:
			err = p.handleMsg(msg)
			if err != nil {
				p.Errorf("msgHandleLoop: handle msg(%s) fail: msg=%s,err=%s", msg.Desc, msg, err)
			}
		case tx := <-p.localTxIn:
			// 存到本地
			p.bc.TxInChan() <- tx
			// 广播
			err = p.broadcastTx(tx)
			if err != nil {
				p.Errorf("msgHandleLoop: broadcast localtx(%s) fail: tx=%v, err=%s", tx.ShortName(), tx, err)
			}
		}
	}
}

//
//// 启动到Ready之间的工作循环
////
//// 理解这里的的Ready很重要。
//// 由于每个时间新区块产生之后的时间被划分为：
//// 			a.t(n) -> b.pot竞争开始 -> c.pot竞争结束 -> d.出块/等待出块 -> e. t(n+1)或者仍是t(n)，pot竞争又开始
//// 在a-c的过程中，可以相信在网络正常情况下，所有正常节点都拿到了相同进度的最新区块。那么对于最新启动的节点，
//// ****** 在loopBeforeReady()环节的结束条件就是在c时刻之前取得第n个区块 ******
//// 另一种情况是：如果响应都到达新启动节点，新启动节点整理完毕发现自身当前处于c-e阶段，那么仍是NotReady的，需要再接收
//// 但是前面两种情况，都可以归到一起讨论：
//// 既然seed被设定为诚实可靠的节点，请求区块的话，直接从seed获取
//// 每获得一次完整的区块响应之后，检查最新区块的创建时间，判断当前时刻处于a-e哪一个阶段
//// 并且相应地切换到对应的状态去处理
//// 这里还隐含了一个问题：
////		TODO
////
//func (p *Pot) loopBeforeReady() {
//	//// 初始为GetNeighbors
//	//if p.getState() != StateType_Init_GetNeighbors {
//	//	p.Fatalf("error state(%s), should be %s\n",
//	//		p.getState().String(), StateType_Init_GetNeighbors.String())
//	//}
//
//	// 向种子节点请求节点表信息
//	p.setState(StateType_Init_GetNeighbors)
//	p.Logf("loopBeforeReady: start collect neighbors\n")
//	err := p.loopCollectNeighbors(p.duty != defines.PeerDuty_Seed, 0)
//	// 种子节点则退出而后进入正常的消息处理循环即可，这里无需额外的逻辑
//	if err == ErrCannotConnectToSeedsWhenInit {
//		if p.duty == defines.PeerDuty_Seed {	// 种子节点
//			// 创建0号区块，初始化世界时钟
//			// TODO
//
//			//genesis, err := p.bc.CreateTheWorld()
//			//if err != nil {
//			//	p.Fatalf("create the world fail: %s\n", err)
//			//}
//			//// 构建时钟
//			//if p.
//
//			return
//		} else {	// 非种子节点
//			p.Fatal("cannot find any seed or peer to conn, exit...")
//		}
//	}
//
//
//	//if p.getState() != StateType_Init_GetProcesses {
//	//	p.Fatalf("error state(%s), should be %s\n",
//	//		p.getState().String(), StateType_Init_GetProcesses.String())
//	//}
//
//	// 向种子节点请求所有共识节点的进度信息
//	//p.setState(StateType_Init_GetProcesses)
//	//p.Logf("loopBeforeReady: start collect processes\n")
//	//p.loopCollectProcesses(p.duty != defines.PeerDuty_Seed, 0)
//
//	//if p.getState() != StateType_Init_GetBlocks {
//	//	p.Fatalf("error state(%s), should be %s\n",
//	//		p.getState().String(), StateType_Init_GetBlocks.String())
//	//}
//
//	// 向种子节点(或者共识节点的某一些)请求区块数据，直至追赶上最新进度
//	p.setState(StateType_Init_GetLatestBlock)
//	p.Logf("loopBeforeReady: start collect blocks until catch up with the latest progress\n")
//	p.loopCollectLatestBlock()
//
//
//}
//
//// 等待邻居消息的循环
//// toseeds true则只向seeds请求；否则向seeds和peers请求
////
//// 本机节点若为seed，则向所有已在线的seed和peer请求信息（涵盖了第一个seed启动的情况），toseeds=true
//// 若为peer，则只向在线的seed请求信息。，toseeds=false
////
//// try 为已尝试次数， 从0开始，若try>=2则直接退出
////
//// 调用：p.loopCollectNeighbors(p.duty != defines.PeerDuty_Seed, 0)
//func (p *Pot) loopCollectNeighbors(toseeds bool, try int) error {
//	if try >= 2 {
//		p.Error("loopCollectNeighbors: disconnect and retry over 2 times")
//
//		// 由于getNeighbors是节点启动后的第一次进行网络交互，所以可以通过此处来判断当前节点是否存在网络问题
//		// 重试两次，都得不到响应，可能的原因有：
//		// 1. 节点表种子节点都没在线（或者说对于本机节点来说，连不上相当于掉线）
//		// 2. 自己网络有问题
//		// 以一个最简单的例子：三节点网络 seed1, peer1-3
//		// 启动顺序： seed1 -> peer1-3
//		// 对于seed1，
//		// 		进入消息处理循环。
//		//		发现网络中没有其他节点，seeds其他seed不存在或未能连通，则构建0号区块，初始化世界时钟
//		// 		当peer1接入时，seed1给它0号区块
//		// 对于随后启动的peer1,
//		// 		通过seed1得到节点表/进度/获取区块
//		//		收到初始区块和世界纪元，根据初始区块和当前纪元，计算相对时间，确定当前处于何种阶段
//		// 其他节点与peer1类似
//
//		return ErrCannotConnectToSeedsWhenInit
//	}
//
//	timeoutD := 4 * time.Second
//	timeout := time.NewTimer(timeoutD)
//	if p.nWaitChan == nil {
//		p.nWaitChan = make(chan int)
//	}
//
//	// 1. 广播getNeighbors消息
//	if err := p.broadcastRequestNeighbors(toseeds); err != nil {
//		p.Errorf("loopCollectNeighbors: broadcast request fail: %s\n", err)
//		return err
//	} else {
//		p.Logf("loopCollectNeighbors: broadcast request succ, nWait: %d\n", p.nWait)
//	}
//
//
//	// 2. 持续收集邻居节点，直至邻居们都有所回应，或等待超时
//	cnt := 0
//	for {
//		select {
//		case <-p.done:
//			p.Logf("loopCollectNeighbors: done and return\n")
//			return nil
//		case <-p.nWaitChan:
//			p.nWait--
//			cnt++
//			p.Logf("loopCollectNeighbors: nWait--\n")
//			if p.nWait == 0 {
//				p.Logf("loopCollectNeighbors: wait finish and return\n")
//				return nil
//			}
//		case <-timeout.C:
//			// 超时需要判断两种情况：
//			if cnt == 0 {	// 一个回复都没收到
//				p.Logf("loopCollectNeighbors: wait timeout, no response received, retry %d\n", try+1)
//				return p.loopCollectNeighbors(false, try + 1)
//			} else {
//				//p.nWait = 0 // 重置，接下来nWait会在broadcastRequestProcesses()中赋一个值
//				p.Logf("loopCollectNeighbors: wait timeout, %d responses received, return\n", cnt)
//				return nil
//			}
//		}
//	}
//}
//
//// TODO: 不需要了
//// 等待进度消息的循环
//// toseeds true则只向seeds请求；否则向seeds和peers请求
////
//// 本机节点若为seed，则向所有已在线的seed和peer请求信息（涵盖了第一个seed启动的情况），toseeds=true
//// 若为peer，则只向在线的seed请求信息。，toseeds=false
////
//// try 为已尝试次数， 从0开始，若try>=2则直接退出
////
//// 调用：p.loopCollectProcesses(p.duty != defines.PeerDuty_Seed, 0)
//func (p *Pot) loopCollectProcesses(toseeds bool, try int) {
//
//	if try >= 2 {
//		p.Error("loopCollectProcesses: disconnect and retry over 2 times")
//		return
//	}
//
//	timeoutD := 4 * time.Second
//	timeout := time.NewTimer(timeoutD)
//	if p.nWaitChan == nil {
//		p.nWaitChan = make(chan int)
//	}
//
//	// 3. 广播getProcess消息
//	if err := p.broadcastRequestProcesses(toseeds); err != nil {
//		p.Errorf("loopCollectProcesses: broadcast request fail: %s\n", err)
//		return
//	} else {
//		p.Logf("loopCollectProcesses: broadcast request succ, nWait: %d\n", p.nWait)
//	}
//
//
//	// 4. 持续收集Process消息
//	cnt := 0
//	for {
//		select {
//		case <-p.done:
//			p.Logf("loopCollectProcesses: done and return\n")
//			return
//		case <-p.nWaitChan:
//			p.nWait--
//			cnt++
//			p.Logf("loopCollectProcesses: nWait--\n")
//			// 等待结束
//			if p.nWait == 0 {
//				p.Logf("loopCollectProcesses: wait finish and return\n")
//				return
//			}
//		case <-timeout.C:
//			// 超时需要判断两种情况：
//			if cnt == 0 {	// 一个回复都没收到
//				p.Logf("loopCollectProcesses: wait timeout, no response received, retry %d\n", try+1)
//				p.loopCollectProcesses(false, try + 1)
//				return
//			} else {
//				p.Logf("loopCollectProcesses: wait timeout, %d responses received, return\n", cnt)
//				return
//			}
//		}
//	}
//}
//
////// 循环收集区块，判断自身
////func (p *Pot) loopCollectBlocks() {
////	// 5. 广播请求区块消息
////
////
////	if err := p.broadcastRequestBlocks(true); err != nil {
////		p.Errorf("loopCollectProcesses: broadcast request fail: %s\n", err)
////		return
////	} else {
////		p.Logf("loopCollectProcesses: broadcast request succ, nWait: %d\n", p.nWait)
////	}
////	// 6. 循环直至追上最新进度
////	for !p.latest() {
////		time.Sleep(5 * time.Millisecond)
////	}
////}
//
//// 请求最新区块
//func (p *Pot) loopCollectLatestBlock() {
//	// 广播请求
//	if err := p.broadcastRequestLatestBlock(); err != nil {
//		p.Errorf("broadcastRequestLatestBlock: broadcast request fail: %s\n", err)
//		return
//	} else {
//		p.Logf("broadcastRequestLatestBlock: broadcast request succ, nWait: %d\n", p.nWait)
//	}
//	// 6. 循环直至追上最新进度
//	for !p.latest() {
//		time.Sleep(5 * time.Millisecond)
//	}
//}
//
//// 节点启动动作：收集邻居信息、收集最新区块信息。 不需要收集process
//
//
//
//
/////////////////////////////////////////////////////////////
//
//// TODO: 假设种子节点只有1台，后续再考虑种子节点的集群容灾
//
//// 种子节点的启动
//// 包括初始化启动与重启动两种
//// 1. 初始化启动： 启动、创造0号区块
//func (p *Pot) loopBeforeReady_Seed() {
//
//
//	// 向种子节点请求节点表信息
//	p.setState(StateType_Init_GetNeighbors)
//	p.Logf("loopBeforeReady: start collect neighbors\n")
//	err := p.loopCollectNeighbors(p.duty != defines.PeerDuty_Seed, 0)
//	// 种子节点则退出而后进入正常的消息处理循环即可，这里无需额外的逻辑
//	if err == ErrCannotConnectToSeedsWhenInit {
//		if p.duty == defines.PeerDuty_Seed {	// 种子节点
//			// 创建0号区块，初始化世界时钟
//			// TODO
//
//			//genesis, err := p.bc.CreateTheWorld()
//			//if err != nil {
//			//	p.Fatalf("create the world fail: %s\n", err)
//			//}
//			//// 构建时钟
//			//if p.
//
//			return
//		} else {	// 非种子节点
//			p.Fatal("cannot find any seed or peer to conn, exit...")
//		}
//	}
//
//
//	//if p.getState() != StateType_Init_GetProcesses {
//	//	p.Fatalf("error state(%s), should be %s\n",
//	//		p.getState().String(), StateType_Init_GetProcesses.String())
//	//}
//
//	// 向种子节点请求所有共识节点的进度信息
//	//p.setState(StateType_Init_GetProcesses)
//	//p.Logf("loopBeforeReady: start collect processes\n")
//	//p.loopCollectProcesses(p.duty != defines.PeerDuty_Seed, 0)
//
//	//if p.getState() != StateType_Init_GetBlocks {
//	//	p.Fatalf("error state(%s), should be %s\n",
//	//		p.getState().String(), StateType_Init_GetBlocks.String())
//	//}
//
//	// 向种子节点(或者共识节点的某一些)请求区块数据，直至追赶上最新进度
//	p.setState(StateType_Init_GetLatestBlock)
//	p.Logf("loopBeforeReady: start collect blocks until catch up with the latest progress\n")
//	p.loopCollectLatestBlock()
//
//
//}
