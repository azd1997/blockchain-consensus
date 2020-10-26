/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/26/20 5:18 PM
* @Description: Pot的工作循环函数
***********************************************************************/

package pot

import "time"

// 启动到Ready之间的工作循环
func (p *Pot) loopBeforeReady() {
	// 初始为GetNeighbors

	// timeoutD用于两个阶段的超时
	timeoutD := 4 * time.Second
	timeout := time.NewTicker(timeoutD)

	//ticker := time.Tick(10 * time.Millisecond)

	p.Logf("loopBeforeReady: start collect neighbors\n")

	// 1. 广播getNeighbors消息
	p.broadcastRequestNeighbors()

	// 2. 持续收集邻居节点，直至邻居们都有所回应，或等待超时
	for {
		select {
		case <-p.done:
			return
		case <-p.nWaitChan:
			p.nWait--
			if p.nWait == 0 {
				// 等待结束，发送消息以请求所有节点进度
				goto GET_PROCESSES
			}
		case <-timeout.C:
			p.nWait = 0		// 重置，接下来nWait会在broadcastRequestProcesses()中赋一个值
			// 等待结束，发送消息以请求所有节点进度
			goto GET_PROCESSES
		}
	}
	// 等待结束，发送消息以请求所有节点进度

GET_PROCESSES:

	p.Logf("loopBeforeReady: start collect processes\n")

	// 3. 广播getProcess消息
	p.broadcastRequestProcesses()

	// 4. 持续收集Process消息
	for {
		select {
		case <-p.done:
			return
		case <-p.nWaitChan:
			p.nWait--
			// 等待结束，发送消息以请求所有节点进度
			if p.nWait == 0 {
				// 等待结束，发送消息以请求所有节点进度
				goto GET_BLOCKS
			}
		case <-timeout.C:
			goto GET_BLOCKS
		}
	}

GET_BLOCKS:

	p.Logf("loopBeforeReady: start collect blocks until catch up with the latest progress\n")

	// 5. 广播请求区块消息
	p.broadcastRequestProcesses()

	// 6. 循环直至追上最新进度
	for !p.latest() {
		time.Sleep(5 * time.Millisecond)
	}

	// 切换状态
	p.setState(StateType_ReadyCompete)
}

//// 等待邻居消息的循环
//func (p *Pot) neighborsLoop() {
//
//}
//
//// NotReady状态等待区块消息的循环
//func (p *Pot) waitBlocksLoop() {
//
//}


// 状态切换循环
// 这里要注意这个循环启动的前提是同步到了最新进度，拿到了世界时钟之后才进入状态切换循环
func (p *Pot) stateMachineLoop() {
	for {
		select {
		case <-p.done:
			p.Logf("stateMachineLoop: return ...\n")
			return
		case <-p.ticker.C:
			// 根据当前状态来处理此滴答消息

			state := p.getState()
			switch state {
			case StateType_NotReady:
				// 没准备好，啥也不干，等区块链同步

				// 如果追上进度了则切换状态为ReadyCompete
				if p.latest() {
					p.setState(StateType_ReadyCompete)
				} else {
					// 否则请求快照数据
					p.requestBlocks()
				}

			case StateType_ReadyCompete:
				// 当前是ReadyCompete，则状态切换为Competing

				// 状态切换
				p.setState(StateType_Competing)
				// 发起竞争（广播证明消息）
				p.broadcastProof()

			case StateType_Competing:
				// 当前是Competing，则状态切换为CompetingEnd，并判断竞赛结果，将状态迅速切换为Winner或Loser

				// 状态切换
				p.setState(StateType_CompeteOver)
				// 判断竞赛结果，状态切换
				if p.winner == p.id { // 自己胜出
					p.setState(StateType_CompeteWinner)
					// 广播新区块
					p.broadcastNewBlock(p.maybeNewBlock)
				} else { // 别人胜出
					p.setState(StateType_CompeteLoser)
					// 等待新区块(“逻辑上”的等待，代码中并不需要wait)
				}

			case StateType_CompeteOver:
				// 正常来说，tick时不会是恰好CompeteOver而又没确定是Winner/Loser
				// 所以暂时无视
			case StateType_CompeteWinner:
				// Winner来说的话，立马广播新区块，广播结束后即切换为Ready
				// 所以不太可能tick时状态为Winner
				// 暂时无视
			case StateType_CompeteLoser:
				// Loser等待新区块，接收到tick说明还没得到新区块
				// 状态切换为Ready

				p.setState(StateType_ReadyCompete)
				// 发起竞争（广播证明消息）
				p.broadcastProof()
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
			return
		case msg := <- p.msgin:
			err = p.handleMsg(msg)
			if err != nil {
				p.Errorf("msgHandleLoop: handle msg(%v) fail: %s\n", msg, err)
			}
		}
	}
}
