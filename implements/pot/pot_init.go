/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/6/20 11:23 AM
* @Description: 初始化与启动
***********************************************************************/

package pot

import (
	"errors"
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

	// 尝试与节点表其他seed联系，请求邻居信息
	total, errs := p.pit.RangeSeeds(func(seed *defines.PeerInfo) error {
		if seed.Id == p.id {return nil}
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Req,
			Epoch:   0,		// 本地没有区块
			From:    p.id,
			To:      seed.Id,
			Reqs:    []*defines.Request{&defines.Request{
				Type:       defines.RequestType_Neighbors,
				Data:       nil,	// 由于自己是seed，对方已知晓，没必要将自己的信息携带过去
			}},
		}
		return p.signAndSendMsg(msg)
	})
	if total - 1 == len(errs) {
		// 与所有其他seed的发信都失败了
		// 存在两种可能：(a)自己是整个网络中第一个启动的seed; (b)自己不是第一个启动的seed，别人是，但是别人目前没在线
		// 目前只按(a)情况处理
		// 为了处理(b)情况，要求所有peer需要定期广播给所有seed，来报告进度，seed可以借此发现自己落后从而纠正
		// 而这个定期广播恰好由证明消息所承担了
		//
		// (a)情况下，当前seed需要创建区块链了：

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
		return nil	// 此情况下预启动完成
	}

	// total - 1 > len(errs) 部分成功或全部成功.   <的情况不用讨论，不可能出现
	p.nWait = total - 1 - len(errs)		// 设置等待数量
	// 等待
	if err := p.wait(); err != nil {	// 一个都没等到
		return err	// 退出启动
	}

	/*
		注意：
		1. 原先的方案中，在收集到邻居表之后是请求最新区块，但是这里存在一个问题：
			假设所有其他节点都是诚实的，那么请求的响应中区块index的情况可能会是1个，也可能是2个（因为时间的问题）
			但如果再考虑恶意节点，那么总的收集到的区块index情形数就会呈现 > 2种情况，且由于诚实答案可能有2种，
			导致没法判断到底是哪个或者哪2个是正确的最新区块index
		2. 为了解决这个问题，只能想办法约束请求回来的数据的诚实的区块index只有1种，
			这样的话，在汇总时直接根据多数获胜原则就可以确定最新区块是哪个。
			这要求节点启动时不能直接在此时就去请求最新区块，而应该尽可能在网络potstart时刻发起请求
			这样的话

		注意，在收集邻居节点之后，自己的节点信息已经被广播出去了。
			其他正在运行周期中的节点可以在有新区块时发送给自己，没必要自己去请求最新区块
			尽管此时自己并没有

		// 当peer集群还不满足条件去共识出块的时候，seed需要按照时间间隔广播伪区块，
		// 提供给其他启动的节点用以同步时钟

		// 节点启动时借助1号区块（可以指定，暂时按1号处理）初始化时钟，
		// 尽管此时时钟偏差会累积，但是偏差较小，可以忽视，，，






	*/

	// 其他seed有在线者，那么向其他seed接着请求最新区块
	total, errs = p.pit.RangeSeeds(func(seed *defines.PeerInfo) error {
		if seed.Id == p.id {return nil}
		msg := &defines.Message{
			Version: defines.CodeVersion,
			Type:    defines.MessageType_Req,
			Epoch:   0,		// 本地没有区块
			From:    p.id,
			To:      seed.Id,
			Reqs:    []*defines.Request{&defines.Request{
				Type:       defines.RequestType_Blocks,
				IndexStart: -1,		// 反向请求1个区块，也即请求最新区块
				IndexCount: 1,
			}},
		}
		return p.signAndSendMsg(msg)
	})
	if total - 1 == len(errs) {		// 所有发信都失败了
		return errors.New("request latest block to seeds all fail")
	}
	// total - 1 > len(errs) 部分发送成功或全部成功.
	p.nWait = total - 1 - len(errs)		// 设置等待数量
	// 等待
	if err := p.wait(); err != nil {	// 一个都没等到
		return err	// 退出启动
	}
	// TODO：这里要求bc有个区块的待定池，进度表有收集最新区块（进度）的能力并且能确定到底哪个是最合适的
	// 这样的话，在收集最新区块，并且判断跟随哪个区块这个地方就比较简单，由进度表决定，接下来在这里只需要
	// 拿进度表决定的最新区块，来初始化时钟
	p.processes.

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