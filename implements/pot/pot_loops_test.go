/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/30/20 4:38 PM
* @Description: The file is for
***********************************************************************/

package pot

// import (
// 	"testing"

// 	"github.com/azd1997/blockchain-consensus/defines"
// 	"github.com/azd1997/blockchain-consensus/modules/peerinfo"
// 	"github.com/azd1997/blockchain-consensus/test"
// 	"github.com/azd1997/blockchain-consensus/utils/log"
// )

// var seed1 = &defines.PeerInfo{
// 	Id:   "seed1",
// 	Addr: "seed1-addr",
// 	Attr: defines.PeerAttr_Normal,
// 	Duty: defines.PeerDuty_Seed,
// 	Data: nil,
// }

// var seed2 = &defines.PeerInfo{
// 	Id:   "seed2",
// 	Addr: "seed2-addr",
// 	Attr: defines.PeerAttr_Normal,
// 	Duty: defines.PeerDuty_Seed,
// 	Data: nil,
// }

// var peer1 = &defines.PeerInfo{
// 	Id:   "peer1",
// 	Addr: "peer1-addr",
// 	Attr: defines.PeerAttr_Normal,
// 	Duty: defines.PeerDuty_Worker,
// 	Data: nil,
// }

// var peer2 = &defines.PeerInfo{
// 	Id:   "peer2",
// 	Addr: "peer2-addr",
// 	Attr: defines.PeerAttr_Normal,
// 	Duty: defines.PeerDuty_Worker,
// 	Data: nil,
// }

// var peer3 = &defines.PeerInfo{
// 	Id:   "peer3",
// 	Addr: "peer3-addr",
// 	Attr: defines.PeerAttr_Normal,
// 	Duty: defines.PeerDuty_Worker,
// 	Data: nil,
// }

// func handleError(t *testing.T, err error) {
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

// func preparePIT(t *testing.T, infos ...*defines.PeerInfo) *peerinfo.PeerInfoTable {
// 	kv := test.NewStore()
// 	pit := peerinfo.NewPeerInfoTable(kv)
// 	if err := pit.Init(); err != nil {
// 		panic(err)
// 	}

// 	// 准备数据
// 	for _, info := range infos {
// 		info := info
// 		err := pit.Set(info)
// 		handleError(t, err)
// 	}

// 	return pit
// }

// //////////////////////////////////////////////////

// //////////////////////////// loopCollectNeighbors /////////////////////////////

// // 验证：
// // 	1. [peer(非种子节点)启动流程] 确实向外发送了请求消息
// //  2. [peer(非种子节点)启动流程] 确实会在nWait等待完毕时退出
// func TestPot_loopCollectNeighbors_Peer(t *testing.T) {
// 	// 节点目前只知道种子节点
// 	pit := preparePIT(t, seed1, peer1)	// peer1是自身信息

// 	pot, err := New(&Option{
// 		Id:      peer1.Id,
// 		Duty:    peer1.Duty,
// 		LogDest: log.LogDest_Stderr,
// 		Pit:     pit,
// 	})
// 	handleError(t, err)

// 	// 启动消息循环
// 	go pot.msgHandleLoop()

// 	// 启动收集邻居循环
// 	done := make(chan struct{})
// 	go func() {
// 		pot.loopCollectNeighbors(pot.duty != defines.PeerDuty_Seed, 0)
// 		close(done)
// 	}()

// 	// 尝试得到这个请求消息
// 	getNeighbors := <-pot.msgout

// 	t.Logf("send msg: %v\n", getNeighbors)		// 注意检查该消息是否符合要求

// 	// 给msgin塞入一个节点表消息（看上去好像是seed回应了）
// 	b, err := peer2.Encode()
// 	handleError(t, err)
// 	ent := &defines.Entry{
// 		Type:      defines.EntryType_Neighbor,
// 		Data:      b,
// 	}
// 	resp := &defines.Message{
// 		Version: defines.CodeVersion,
// 		Type:    defines.MessageType_Data,
// 		From:    seed1.Id,
// 		To:      peer1.Id,
// 		Entries: []*defines.Entry{ent},
// 	}
// 	err = resp.Sign()
// 	handleError(t, err)
// 	pot.msgin <- resp

// 	<-done
// 	// 关闭循环
// 	close(pot.done)
// 	// 关闭节点表
// 	handleError(t, pit.Close())
// }

// // 验证：
// // 	1. [第一个seed(种子节点)启动流程] 重试两次后退出
// func TestPot_loopCollectNeighbors_FristSeed(t *testing.T) {
// 	// 节点目前只知道种子节点
// 	pit := preparePIT(t, seed1)	// seed1是自身信息

// 	pot, err := New(&Option{
// 		Id:      seed1.Id,
// 		Duty:    seed1.Duty,
// 		LogDest: log.LogDest_Stderr,
// 		Pit:     pit,
// 	})
// 	handleError(t, err)

// 	// 启动消息循环
// 	go pot.msgHandleLoop()

// 	// 启动收集邻居循环
// 	done := make(chan struct{})
// 	go func() {
// 		pot.loopCollectNeighbors(pot.duty != defines.PeerDuty_Seed, 0)
// 		close(done)
// 	}()

// 	// 由于结点表没有其他seed/peer信息，所以没有消息发送，也就没有回应

// 	<-done
// 	// 关闭循环
// 	close(pot.done)
// 	// 关闭节点表
// 	handleError(t, pit.Close())
// }

// // 验证：
// // 	1. [seed(种子节点)启动流程] 确实向外发送了请求消息
// //  2. [seed(种子节点)启动流程] 确实会在nWait等待超时后退出
// func TestPot_loopCollectNeighbors_Seed(t *testing.T) {
// 	// 节点目前只知道种子节点
// 	pit := preparePIT(t, seed1, seed2, peer1)	// seed1是自身信息

// 	pot, err := New(&Option{
// 		Id:      seed1.Id,
// 		Duty:    seed1.Duty,
// 		LogDest: log.LogDest_Stderr,
// 		Pit:     pit,
// 	})
// 	handleError(t, err)

// 	// 启动消息循环
// 	go pot.msgHandleLoop()

// 	// 启动收集邻居循环
// 	done := make(chan struct{})
// 	go func() {
// 		pot.loopCollectNeighbors(pot.duty != defines.PeerDuty_Seed, 0)
// 		close(done)
// 	}()

// 	// 尝试得到请求消息，应该会有两个
// 	getNeighbors1 := <-pot.msgout
// 	t.Logf("send msg: %v\n", getNeighbors1)		// 注意检查该消息是否符合要求
// 	getNeighbors2 := <-pot.msgout
// 	t.Logf("send msg: %v\n", getNeighbors2)		// 注意检查该消息是否符合要求

// 	// 给msgin塞入一个节点表消息（看上去好像是seed2回应了），但peer1出问题了没回应
// 	b, err := peer2.Encode()
// 	handleError(t, err)
// 	ent := &defines.Entry{
// 		Type:      defines.EntryType_Neighbor,
// 		Data:      b,
// 	}
// 	resp := &defines.Message{
// 		Version: defines.CodeVersion,
// 		Type:    defines.MessageType_Data,
// 		From:    seed2.Id,
// 		To:      seed1.Id,
// 		Entries: []*defines.Entry{ent},
// 	}
// 	err = resp.Sign()
// 	handleError(t, err)
// 	pot.msgin <- resp

// 	<-done
// 	// 关闭循环
// 	close(pot.done)
// 	// 关闭节点表
// 	handleError(t, pit.Close())
// }

// //////////////////////////// loopCollectProcesses /////////////////////////////

// // 验证：
// // 	1. 确实向外发送了请求消息
// //  2. 确实会在nWait等待完毕时退出
// func TestPot_loopCollectProcesses(t *testing.T) {
// 	// 节点目前只知道种子节点
// 	pit := preparePIT(t, seed1, peer1)	// peer1是自身信息

// 	pot, err := New(&Option{
// 		Id:      peer1.Id,
// 		Duty:    peer1.Duty,
// 		LogDest: log.LogDest_Stderr,
// 		Pit:     pit,
// 	})
// 	handleError(t, err)

// 	pot.setState(StateType_Init_GetProcesses)

// 	// 启动消息循环
// 	go pot.msgHandleLoop()

// 	// 启动收集邻居循环
// 	done := make(chan struct{})
// 	go func() {
// 		pot.loopCollectProcesses(pot.duty != defines.PeerDuty_Seed, 0)
// 		close(done)
// 	}()

// 	// 尝试得到这个请求消息
// 	getProcesses := <-pot.msgout

// 	t.Logf("send msg: %v\n", getProcesses)		// 注意检查该消息是否符合要求

// 	// 给msgin塞入一个进度消息（看上去好像是seed回应了）
// 	pr := &defines.Process{
// 		Index:       3,
// 		Hash:        nil,
// 		LatestMaker: "eiger",
// 		Id:          peer2.Id,		// 这是seed1记录的peer2的进度
// 	}
// 	b, err := pr.Encode()
// 	handleError(t, err)
// 	ent := &defines.Entry{
// 		Type:      defines.EntryType_Process,
// 		Data:      b,
// 	}
// 	resp := &defines.Message{
// 		Version: defines.CodeVersion,
// 		Type:    defines.MessageType_Data,
// 		From:    seed1.Id,
// 		To:      peer1.Id,
// 		Entries: []*defines.Entry{ent},
// 	}
// 	err = resp.Sign()
// 	handleError(t, err)
// 	pot.msgin <- resp

// 	<-done
// 	// 关闭循环
// 	close(pot.done)
// 	// 关闭节点表
// 	handleError(t, pit.Close())
// }

// //////////////////////////// loopCollectProcesses /////////////////////////////

// // 验证：
// // 	1. 确实向外发送了请求消息
// //  2.
// func TestPot_loopCollectBlocks(t *testing.T) {
// 	// 节点目前只知道种子节点
// 	pit := preparePIT(t, seed1, peer1)	// peer1是自身信息

// 	pot, err := New(&Option{
// 		Id:      peer1.Id,
// 		Duty:    peer1.Duty,
// 		LogDest: log.LogDest_Stderr,
// 		Pit:     pit,
// 	})
// 	handleError(t, err)

// 	pot.setState(StateType_Init_GetProcesses)

// 	// 启动消息循环
// 	go pot.msgHandleLoop()

// 	// 启动收集邻居循环
// 	done := make(chan struct{})
// 	go func() {
// 		pot.loopCollectProcesses(pot.duty != defines.PeerDuty_Seed, 0)
// 		close(done)
// 	}()

// 	// 尝试得到这个请求消息
// 	getProcesses := <-pot.msgout

// 	t.Logf("send msg: %v\n", getProcesses)		// 注意检查该消息是否符合要求

// 	// 给msgin塞入一个进度消息（看上去好像是seed回应了）
// 	pr := &defines.Process{
// 		Index:       3,
// 		Hash:        nil,
// 		LatestMaker: "eiger",
// 		Id:          peer2.Id,		// 这是seed1记录的peer2的进度
// 	}
// 	b, err := pr.Encode()
// 	handleError(t, err)
// 	ent := &defines.Entry{
// 		Type:      defines.EntryType_Process,
// 		Data:      b,
// 	}
// 	resp := &defines.Message{
// 		Version: defines.CodeVersion,
// 		Type:    defines.MessageType_Data,
// 		From:    seed1.Id,
// 		To:      peer1.Id,
// 		Entries: []*defines.Entry{ent},
// 	}
// 	err = resp.Sign()
// 	handleError(t, err)
// 	pot.msgin <- resp

// 	<-done
// 	// 关闭循环
// 	close(pot.done)
// 	// 关闭节点表
// 	handleError(t, pit.Close())
// }
