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

// initForSeedFirstStart seed初次启动
func (p *Pot) initForSeedFirstStart() error {
	p.Info("initForSeedFirstStart")

	// 尝试与节点表其他seed联系，请求邻居信息
	p.setState(StateType_PreInited_RequestNeighbors)
	rnf, err := p.requestNeighborsFuncGenerator()
	if err != nil {
		return err
	}
	total, errs := p.pit.RangeSeeds(rnf)
	//fmt.Println(total, errs, p.pit.NSeed(), p.pit.Seeds())
	if total == len(errs) {
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
			p.Fatalf("initForSeedFirstStart: create genesis block fail: %s\n", err)
		}
		p.b1Time = genesis.Timestamp
		// 启动时钟
		p.clock.Start(genesis)
		// 更新自己进度
		//p.processes.refresh(genesis)

		// 进入PostPot
		p.setState(StateType_PostPot)
		return nil // 此情况下预启动完成
	} else if total < len(errs) {
		//p.Infof("total=%d, len(errs)=%d, errs=%v", total, len(errs), errs)
		return errors.New("fatal error: impossible")
	}

	// total > len(errs) 部分成功或全部成功.   <的情况不用讨论，不可能出现
	nWait := total - len(errs) // 设置等待数量
	// 等待
	if err := p.wait(nWait); err != nil { // 一个都没等到
		return err // 退出启动
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

	// 其他seed有在线者，那么向其他seed请求1号区块先
	p.setState(StateType_PreInited_RequestFirstBlock)
	firstBlock, err := p.requestOneBlockAndWait(false, 1)
	if err != nil {
		return err
	}
	p.b1Time = firstBlock.Timestamp
	// 初始化时钟
	p.clock.Start(firstBlock)
	// 将第一个区块加入到本地。这里对于1号区块的添加是使用AddNewBlock，特殊处理
	if err := p.bc.AddNewBlock(firstBlock); err != nil {
		p.Errorf("add first block fail: %s", err)
	}

	// 3. 等待一段时间，到达PotStart时刻
	<-p.potStartBeforeReady

	// 4. 向seed或者peer请求最新区块
	p.setState(StateType_PreInited_RequestLatestBlock)
	latestBlock, err := p.requestLatestBlockAndWait(false)
	if err != nil {
		return err
	}
	if err := p.bc.AddNewBlock(latestBlock); err != nil {
		p.Errorf("add latest block fail: %s", err)
	}
	// 驱动时钟，跟上网络时间
	if err := p.clock.Trigger(latestBlock); err != nil {
		return err
	}

	//// 其他seed有在线者，那么向其他seed接着请求最新区块
	//total, errs = p.pit.RangeSeeds(p.requestLatestBlockFuncGenerator())
	//if total - 1 == len(errs) {
	//	// 所有发信都失败了，如果前面请求邻居信息成功了，而这里失败，说明之前活跃的seed挂掉了
	//	// 为了降低启动流程的复杂性，这里直接令其退出
	//	return errors.New("request latest block to seeds all fail")
	//} else if total - 1 < len(errs) {
	//	return errors.New("fatal error: impossible")
	//}
	//// total - 1 > len(errs) 部分发送成功或全部成功.
	//p.nWait = total - 1 - len(errs)		// 设置等待数量
	//// 等待 (此时的handle函数会将所有区块临时存储)
	//if err := p.wait(); err != nil {	// 一个都没等到
	//	return err	// 退出启动
	//}
	//// 由于seed是“可信的”，那么将仅比较index来确定(由于发信/收信时延的不确定性，有可能会出现回复的seed
	//// 给出index和index+1两个连续的区块的情况，取index更大的。
	////
	//// 此外，由于这些seed是可信的，所以它们的区块都是可信的，直接存入区块链即可，并且每次得到区块都用来触
	//// 发网络时钟，当然必须是index更大的区块才行)

	// 切换状态
	p.setState(StateType_NotReady)

	// 请求缺失的区块（如果有缺失的话）
	start, end := firstBlock.Index+1, latestBlock.Index-1
	if end >= start {
		count := end - start + 1
		p.Infof("current is discontinuous, req block %d-%d, %d blocks", start, end, count)
		p.broadcastRequestBlocksByIndex(start, count)
	}

	return nil
}

// initForSeedReStart seed重启动
func (p *Pot) initForSeedReStart() error {
	p.Info("initForSeedReStart")

	// 1. 广播seeds(以及peers)，看有无在线的
	// 尝试与节点表其他seed联系，请求邻居信息
	p.setState(StateType_PreInited_RequestNeighbors)
	seedsAllFail, err := p.requestNeighborsAndWait()
	if err != nil {
		return err
	}

	// 2. 根据本地已有区块，初始化时钟
	localMaxBlock, err := p.bc.GetBlocksByRange(-1, 1)
	if err != nil {
		return err
	}
	p.clock.Start(localMaxBlock[0])

	firstBlock, err := p.bc.GetBlocksByRange(1, 1)
	if err != nil {
		return err
	}
	p.b1Time = firstBlock[0].Timestamp

	// 3. 等待一段时间，到达PotStart时刻
	<-p.potStartBeforeReady

	// 4. 发起请求最新区块。
	// 之所以要在PotStart时刻请求，是为了降低复杂性，2*TickMs能保证回应能在新区块诞生前收到
	p.setState(StateType_PreInited_RequestLatestBlock)
	latestBlock, err := p.requestLatestBlockAndWait(seedsAllFail)
	if err != nil {
		return err
	}
	if err := p.bc.AddNewBlock(latestBlock); err != nil {
		p.Errorf("add latest block fail: %s", err)
	}

	// 5. 驱动时钟，跟上网络时间
	if err := p.clock.Trigger(latestBlock); err != nil {
		return err
	}

	// 6. 设置状态
	p.setState(StateType_NotReady)

	// 请求缺失的区块（如果有缺失的话）
	start, end := localMaxBlock[0].Index+1, latestBlock.Index-1
	if end >= start {
		count := end - start + 1
		p.broadcastRequestBlocksByIndex(start, count)
	}

	return nil
}

// initForPeerFirstStart peer初次启动
func (p *Pot) initForPeerFirstStart() error {
	p.Info("initForPeerFirstStart")

	// 1. 向seeds和预配置的peers请求节点表
	p.setState(StateType_PreInited_RequestNeighbors)
	seedsAllFail, err := p.requestNeighborsAndWait()
	if err != nil {
		return err
	}

	// 2. 请求1号区块，初始化时钟	// TODO: 1号区块距当前最新区块太远导致时间偏差较大问题，待解决
	p.setState(StateType_PreInited_RequestFirstBlock)
	firstBlock, err := p.requestOneBlockAndWait(seedsAllFail, 1)
	if err != nil {
		return err
	}
	p.b1Time = firstBlock.Timestamp
	// 初始化时钟
	p.clock.Start(firstBlock)
	if err := p.bc.AddNewBlock(firstBlock); err != nil {
		p.Errorf("add first block fail: %s", err)
	}

	// 3. 等待一段时间，到达PotStart时刻
	<-p.potStartBeforeReady

	// 4. 向seed或者peer请求最新区块
	p.setState(StateType_PreInited_RequestLatestBlock)
	latestBlock, err := p.requestLatestBlockAndWait(seedsAllFail)
	if err != nil {
		return err
	}
	if err := p.bc.AddNewBlock(latestBlock); err != nil {
		p.Errorf("add latest block fail: %s", err)
	}

	// 5. 驱动时钟，跟上网络时间
	if err := p.clock.Trigger(latestBlock); err != nil {
		return err
	}

	// 6. 设置当前状态
	p.setState(StateType_NotReady)

	// 请求缺失的区块（如果有缺失的话）
	start, end := firstBlock.Index+1, latestBlock.Index-1
	if end >= start {
		count := end - start + 1
		p.broadcastRequestBlocksByIndex(start, count)
	}

	return nil
}

// initForPeerReStart peer重启动
func (p *Pot) initForPeerReStart() error {
	p.Info("initForPeerReStart")

	// 1. 向seeds和预配置的peers请求节点表
	p.setState(StateType_PreInited_RequestNeighbors)
	seedsAllFail, err := p.requestNeighborsAndWait()
	if err != nil {
		return err
	}

	// 2. 根据本地已有区块，初始化时钟
	localMaxBlock, err := p.bc.GetBlocksByRange(-1, 1)
	if err != nil {
		return err
	}
	p.clock.Start(localMaxBlock[0])

	firstBlock, err := p.bc.GetBlocksByRange(1, 1)
	if err != nil {
		return err
	}
	p.b1Time = firstBlock[0].Timestamp

	// 3. 等待一段时间，到达PotStart时刻
	<-p.potStartBeforeReady

	// 4. 向seed或者peer请求最新区块
	p.setState(StateType_PreInited_RequestLatestBlock)
	latestBlock, err := p.requestLatestBlockAndWait(seedsAllFail)
	if err != nil {
		return err
	}
	if err := p.bc.AddNewBlock(latestBlock); err != nil {
		p.Errorf("add latest block fail: %s", err)
	}
	// 5. 驱动时钟，跟上网络时间
	if err := p.clock.Trigger(latestBlock); err != nil {
		return err
	}

	// 6. 设置当前状态
	p.setState(StateType_NotReady)

	// 请求缺失的区块（如果有缺失的话）
	start, end := localMaxBlock[0].Index+1, latestBlock.Index-1
	if end >= start {
		count := end - start + 1
		p.broadcastRequestBlocksByIndex(start, count)
	}

	return nil
}

////////////////////////////////////////////////////

// 请求邻居节点
func (p *Pot) requestNeighborsAndWait() (seedsAllFail bool, err error) {
	all := 0
	rnf, err := p.requestNeighborsFuncGenerator()
	if err != nil {
		return false, err
	}
	total, errs := p.pit.RangeSeeds(rnf)
	all = total - bool2int(p.duty == defines.PeerDuty_Seed)
	if all == len(errs) { // 全部发信失败，则尝试请求peers
		seedsAllFail = true
		total, errs = p.pit.RangePeers(rnf)
		if total == len(errs) { // 还是全部失败，则退出
			return true, errors.New("request neighbors to seeds and peers all fail")
		}
		all = total - bool2int(p.duty == defines.PeerDuty_Seed)
	}
	if all < len(errs) {
		return seedsAllFail, errors.New("fatal error: impossible")
	}
	// 部分成功或全部成功
	nWait := all - len(errs) // 设置等待数量
	// 等待 (此时的handle函数会将邻居节点进行存储，若重复以先存为准)
	if err := p.wait(nWait); err != nil { // 一个都没等到
		return seedsAllFail, err // 退出启动
	}
	return seedsAllFail, nil
}

// 请求最新区块
func (p *Pot) requestLatestBlockAndWait(seedsAllFail bool) (*defines.Block, error) {
	return p.requestOneBlockAndWait(seedsAllFail, -1)
}

// 请求具体某一个区块
func (p *Pot) requestOneBlockAndWait(seedsAllFail bool, index int64) (*defines.Block, error) {
	var total, all int
	var errs map[string]error
	var nWait int
	if seedsAllFail {
		total, errs = p.pit.RangePeers(p.requestOneBlockFuncGenerator(index))
	} else {
		total, errs = p.pit.RangeSeeds(p.requestOneBlockFuncGenerator(index))
	}
	all = total - bool2int(p.duty == defines.PeerDuty_Seed)
	if all == len(errs) { // 全部发信失败，则退出
		return nil, errors.New("request latest block to seeds and peers all fail")
	} else if all < len(errs) {
		return nil, errors.New("fatal error: impossible")
	} else {
		// 部分成功或全部成功
		nWait = all - len(errs) // 设置等待数量
	}
	// 等待该区块，并从众多回复中确定接收哪一个
	if wb, err := p.waitAndDecideOneBlock(index, nWait); err != nil { // 一个都没等到
		return nil, err // 退出启动
	} else {
		return wb, nil
	}
}
