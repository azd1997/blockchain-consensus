/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/6/20 11:23 AM
* @Description: 初始化与启动
***********************************************************************/

package pot

import (
	"errors"
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"time"
)

const (
	retryMaxTimesWhenInit = 10
	retryDurationWhenInit = 1 * time.Second
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

// 创世区块一定先启动，其他节点必需与创世节点连接上才能启动，其他节点可以暂时连接不上
func (p *Pot) initForGenesis() error {
	p.Info("initForGenesis")

	genesis, err := p.bc.CreateTheWorld()
	if err != nil {
		p.Fatalf("initForGenesis: create genesis block fail: %s\n", err)
	}
	p.b1Time = genesis.Timestamp
	// 启动时钟
	p.clock.Start(genesis)
	p.Infof("start clock succ. b1Time: %d", genesis.Timestamp)

	// 进入PostPot
	p.setStage(StageType_PostPot) // 在这里设置是因为刚刚初始化clock，此时的PotOver时刻信号其实没有，所以手动设置stage
	// 设置状态
	if p.duty == defines.PeerDuty_Seed {
		p.setState(StateType_Judger)
	} else if p.duty == defines.PeerDuty_Peer {
		p.setState(StateType_Winner)
	} else {
		p.Fatalf("initForGenesis: only peer and seed can become a genesis\n")
	}
	return nil
}

// 启动流程应该是：
// 一方面， 节点不断请求节点表内预配置的seed/peer list
	// 当有一个返回之后就

// initForSeedFirstStart seed初次启动
func (p *Pot) initForSeedFirstStart() error {
	p.Info("initForSeedFirstStart")

	retryWhenInit := 0
	// 尝试与节点表其他seed联系，请求邻居信息
	p.setStage(StageType_PreInited_RequestNeighbors)
	_, err := p.requestNeighborsAndWait() // 约2s
	// 本机节点必需至少找到一个节点进行连接
	for err != nil {
		// 不断重试
		retryWhenInit++
		time.Sleep(retryDurationWhenInit * time.Duration(retryWhenInit))
		p.Infof("initForPeerFirstStart: [RN] retryWhenInit: %d", retryWhenInit)
		if retryWhenInit > retryMaxTimesWhenInit {
			return errors.New("retry too much times when init")
		}
		_, err = p.requestNeighborsAndWait()
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
	p.setStage(StageType_PreInited_RequestFirstBlock)
	firstBlock, err := p.requestOneBlockAndWait(false, 1)
	if err != nil {
		return err
	}
	p.b1Time = firstBlock.Timestamp
	// 初始化时钟
	p.clock.Start(firstBlock)
	p.Infof("start clock succ. b1Time: %d", firstBlock.Timestamp)
	// 将第一个区块加入到本地。这里对于1号区块的添加是使用AddNewBlock，特殊处理
	if err := p.bc.AddNewBlock(firstBlock); err != nil {
		p.Errorf("add first block fail: %s", err)
	}

	// 3. 等待一段时间，到达PotStart时刻
	<-p.potStartBeforeReady
	p.setStage(StageType_InPot)   // 因为此时的POTsTART信号已经被用了，所以手动设置stage
	p.setState(StateType_Witness) // 此时只能设置为witness
	return nil
}

// initForSeedReStart seed重启动
func (p *Pot) initForSeedReStart() error {
	p.Info("initForSeedReStart")

	// 1. 广播seeds(以及peers)，看有无在线的
	// 尝试与节点表其他seed联系，请求邻居信息
	retryWhenInit := 0
	// 尝试与节点表其他seed联系，请求邻居信息
	p.setStage(StageType_PreInited_RequestNeighbors)
	_, err := p.requestNeighborsAndWait() // 约2s
	// 本机节点必需至少找到一个节点进行连接
	for err != nil {
		// 不断重试
		retryWhenInit++
		time.Sleep(retryDurationWhenInit * time.Duration(retryWhenInit))
		p.Infof("initForPeerFirstStart: retryWhenInit: %d", retryWhenInit)
		if retryWhenInit > retryMaxTimesWhenInit {
			return errors.New("retry too much times when init")
		}
		_, err = p.requestNeighborsAndWait()
	}

	// 2. 根据本地已有区块，初始化时钟
	//localMaxBlock, err := p.bc.GetBlocksByRange(-1, 1)
	//if err != nil {
	//	return err
	//}
	//p.clock.Start(localMaxBlock[0])

	firstBlock, err := p.bc.GetBlocksByRange(1, 1)
	if err != nil {
		return err
	}
	p.clock.Start(firstBlock[0])
	p.b1Time = firstBlock[0].Timestamp
	p.Infof("start clock succ. b1Time: %d", firstBlock[0].Timestamp)

	// 3. 等待一段时间，到达PotStart时刻
	<-p.potStartBeforeReady

	p.setStage(StageType_InPot)   // 因为此时的POTsTART信号已经被用了，所以手动设置stage
	p.setState(StateType_Witness) // 此时只能设置为witness
	return nil
}

// initForPeerFirstStart peer初次启动
func (p *Pot) initForPeerFirstStart() error {
	p.Info("initForPeerFirstStart")

	// 1. 向seeds和预配置的peers请求节点表
	retryWhenInit := 0
	// 尝试与节点表其他seed联系，请求邻居信息
	p.setStage(StageType_PreInited_RequestNeighbors)
	seedsAllFail, err := p.requestNeighborsAndWait() // 约2s
	// 本机节点必需至少找到一个节点进行连接
	for err != nil {
		// 不断重试
		retryWhenInit++
		time.Sleep(retryDurationWhenInit * time.Duration(retryWhenInit))
		p.Infof("initForPeerFirstStart: [RN] retryWhenInit: %d\n", retryWhenInit)
		if retryWhenInit > retryMaxTimesWhenInit {
			return errors.New("retry too much times when init")
		}
		seedsAllFail, err = p.requestNeighborsAndWait()
	}
	p.Infof("initForPeerFirstStart: seedsAllFail: %v\n", seedsAllFail)

	// 2. 请求1号区块，初始化时钟	// TODO: 1号区块距当前最新区块太远导致时间偏差较大问题，待解决
	retryWhenInit = 0
	p.setStage(StageType_PreInited_RequestFirstBlock)
	firstBlock, err := p.requestOneBlockAndWait(seedsAllFail, 1)
	//if err != nil {
	//	return err
	//}
	for err != nil {
		// 不断重试
		retryWhenInit++
		time.Sleep(retryDurationWhenInit * time.Duration(retryWhenInit))
		p.Infof("initForPeerFirstStart: [RFB] retryWhenInit: %d\n", retryWhenInit)
		if retryWhenInit > retryMaxTimesWhenInit {
			return errors.New("retry too much times when init")
		}
		firstBlock, err = p.requestOneBlockAndWait(seedsAllFail, 1)
	}

	p.b1Time = firstBlock.Timestamp
	// 初始化时钟
	p.clock.Start(firstBlock)
	p.Infof("start clock succ. b1Time: %d", firstBlock.Timestamp)
	if err := p.bc.AddNewBlock(firstBlock); err != nil {
		p.Errorf("add first block fail: %s", err)
	}
	p.Infof("initForPeerFirstStart: firstBlock: %v\n", firstBlock)

	// 3. 等待一段时间，到达PotStart时刻
	<-p.potStartBeforeReady
	p.setStage(StageType_InPot)   // 因为此时的POTsTART信号已经被用了，所以手动设置stage
	p.setState(StateType_Witness) // 此时只能设置为witness
	return nil
}

// initForPeerReStart peer重启动
func (p *Pot) initForPeerReStart() error {
	p.Info("initForPeerReStart")

	// 1. 向seeds和预配置的peers请求节点表
	retryWhenInit := 0
	// 尝试与节点表其他seed联系，请求邻居信息
	p.setStage(StageType_PreInited_RequestNeighbors)
	_, err := p.requestNeighborsAndWait() // 约2s
	// 本机节点必需至少找到一个节点进行连接
	for err != nil {
		// 不断重试
		retryWhenInit++
		time.Sleep(retryDurationWhenInit * time.Duration(retryWhenInit))
		p.Infof("initForPeerFirstStart: retryWhenInit: %d", retryWhenInit)
		if retryWhenInit > retryMaxTimesWhenInit {
			return errors.New("retry too much times when init")
		}
		_, err = p.requestNeighborsAndWait()
	}

	// 2. 根据本地已有区块，初始化时钟
	//localMaxBlock, err := p.bc.GetBlocksByRange(-1, 1)
	//if err != nil {
	//	return err
	//}
	//p.clock.Start(localMaxBlock[0])

	firstBlock, err := p.bc.GetBlocksByRange(1, 1)
	if err != nil {
		return err
	}
	p.clock.Start(firstBlock[0])
	p.b1Time = firstBlock[0].Timestamp
	p.Infof("start clock succ. b1Time: %d", firstBlock[0].Timestamp)

	// 3. 等待一段时间，到达PotStart时刻
	<-p.potStartBeforeReady

	p.setStage(StageType_InPot)   // 因为此时的POTsTART信号已经被用了，所以手动设置stage
	p.setState(StateType_Witness) // 此时只能设置为witness
	return nil
}

////////////////////////////////////////////////////

// requestNeighborsAndWait 广播getNeighbors请求
// seedsAllFail代表种子节点全部请求失败
// err!=nil代表全部节点请求失败
func (p *Pot) requestNeighborsAndWait() (seedsAllFail bool, err error) {

	_, succ, err := p.broadcastRequestNeighbors(true, false)
	if err != nil { // err!=nil意味着可能的发信全部失败
		seedsAllFail = true
		// 尝试向peer请求
		_, succ, err = p.broadcastRequestNeighbors(false, true)
		if err != nil { // 全部失败则退出
			return true, errors.New("requestNeighborsAndWait: request neighbors to seeds and peers all fail")
		}
	}
	nWait := succ                         // 需要等待succ个节点的回信
	if err := p.wait(nWait); err != nil { // 一个都没等到
		return seedsAllFail, fmt.Errorf("requestNeighborsAndWait: %s", err) // 退出启动
	}
	return seedsAllFail, nil
}

// requestOneBlockAndWait 请求具体某一个区块
func (p *Pot) requestOneBlockAndWait(seedsAllFail bool, index int64) (*defines.Block, error) {
	var succ int
	var err error
	var nWait int
	if !seedsAllFail {
		_, succ, err = p.broadcastRequestBlocksByIndex(index, 1, true, false)
	} else {
		_, succ, err = p.broadcastRequestBlocksByIndex(index, 1, false, true)
	}
	if err != nil { // 全部发信失败，则退出
		return nil, errors.New("requestOneBlockAndWait: request one block to seeds and peers all fail")
	}
	// 部分成功或全部成功
	nWait = succ // 设置等待数量
	// 等待该区块，并从众多回复中确定接收哪一个
	if wb, err := p.waitAndDecideOneBlock(index, nWait); err != nil { // 一个都没等到
		return nil, fmt.Errorf("requestOneBlockAndWait: %s", err) // 退出启动
	} else {
		return wb, nil
	}
}
