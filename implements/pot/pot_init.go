/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/6/20 11:23 AM
* @Description: 初始化与启动
***********************************************************************/

package pot

import (
	"github.com/azd1997/blockchain-consensus/defines"
)

// Init 初始化
// 启动流程：
//		1. 启动消息处理循环
//		2. 先后请求Neighbors/Processes/Blocks，直至自身节点进入Ready状态
// 		3. 在取到“最新区块”的时候，按照自身时间戳与最新区块的构建时间/索引，以及响应侧epoch，判断当前处于何种阶段，启动定时器
//		4. 启动状态切换循环
//
// 种子节点启动和非种子节点启动有所不同:
// 		1. 种子节点启动时需要向其他种子节点与所有一直在线的非种子节点获取信息
//		2. 非种子节点启动时可以只通过种子节点来获取信息直至到达最新状态
//
// ** 整个网络启动时一定是先启动种子节点而后启动非种子节点；种子节点允许中间重启
//
func (p *Pot) Init() error {

	// 启动消息处理循环
	go p.msgHandleLoop()
	// 启动状态切换循环(没有clock触发)
	go p.stateMachineLoop()

	// 收集并更新节点表
	// 如果自己是种子节点且种子总数只有1，且当前没有peer，那么直接退出，此时说明自己是全网第一个节点.这种情况下不必
	if !(p.duty == defines.PeerDuty_Seed && p.pit.NSeed() == 1 && p.pit.NPeer() == 0) {
		err := p.loopCollectNeighbors(true, 0)
		if err != nil {
			return err
		}
	}

	//

	// 区块链的最新状态
	bc := p.bc.GetMaxIndex()
	// 根据节点duty和其本地区块链状态决定以何种逻辑启动
	if p.duty == defines.PeerDuty_Seed {	// seed
		if bc == 0 {	// 初次启动
			return p.init_Seed_FirstStart()
		} else {	// 重启动
			return p.init_Seed_ReStart()
		}
	} else {	// peer
		if bc == 0 {	// 初次启动
			return p.init_Peer_FirstStart()
		} else {	// 重启动
			return p.init_Peer_ReStart()
		}
	}



	//// 阻塞直到追上最新进度
	//p.loopBeforeReady()
	//
	//// 切换状态，准备进入状态切换循环
	//p.setState(StateType_ReadyCompete)
	//
	//// 启动世界时钟
	//if
	//p.clock = time.NewTimer(time.Second)
	//
	//

	//return nil
}

// 作为Seed初始化
// 包括初始化启动(还没有区块链)与重启动两种
// 	1. 初始化启动： 启动消息处理循环/创造0号区块/启动时钟/进入
//	2. 重启动：启动消息处理循环/等待直到收到NewBlock
//func (p *Pot) init_Seed() {
//	// 启动消息处理循环
//	go p.msgHandleLoop()
//	// 判断是初始化启动还是重启动
//	if p.bc.GetMaxIndex() == 0 {	// 说明还没有区块，也就是初始化启动
//		// 创建创世区块(1号区块)
//		genesis, err := p.bc.CreateTheWorld()
//		if err != nil {
//			p.Fatalf("init_seed: create genesis block fail: %s\n", err)
//		}
//		// 启动时钟
//		c := StartClock(genesis)
//		if c == nil {
//			p.Fatalf("init_seed: start clock fail\n")
//		}
//		//p.setState(Sta)
//	}
//
//}

// TODO 暂时只考虑1台seed，

// seed初次启动
func (p *Pot) init_Seed_FirstStart() error {

	//

	// 创建创世区块(1号区块)
	genesis, err := p.bc.CreateTheWorld()
	if err != nil {
		p.Fatalf("init_Seed_FirstStart: create genesis block fail: %s\n", err)
	}
	// 启动时钟
	c := StartClock(genesis)
	if c == nil {
		p.Fatalf("init_Seed_FirstStart: start clock fail\n")
	}
	// 进入PreInited状态，直到检测到网络中存在3个及以上最新进度的、在线的节点
	p.setState(StateType_Seed_PreInited)

	return nil
}

// seed重启动
func (p *Pot) init_Seed_ReStart() error {

	// 1. 广播peers，请求节点表信息

	// 2. 广播peers，请求最新区块信息

	// 3. 确定最新区块，初始化时钟

	return nil
}

// peer初次启动
func (p *Pot) init_Peer_FirstStart() error {

	// 1. 向seeds请求节点表

	return nil
}

// peer重启动
func (p *Pot) init_Peer_ReStart() error {
return nil
}


////////////////////////////////////////////////////

//
func (p *Pot) collectNeighbors() {

}