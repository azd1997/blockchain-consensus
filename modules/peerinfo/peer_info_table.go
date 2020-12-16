/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/8/20 6:54 PM
* @Description: 节点信息表
***********************************************************************/

package peerinfo

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
	"github.com/azd1997/blockchain-consensus/utils/log"
)

// dirtyPeerInfo 脏的、与从存储引擎不一致的数据
type dirtyPeerInfo struct {
	op   Op
	info *defines.PeerInfo
}

// Op 对PeerInfoTable的操作 None Del Set
type Op uint8

const (
	OpNone Op = 0 // 默认值，因此每次创建DirtyPeerInfo，都需要指定Op
	OpDel  Op = 1 // 删除
	OpSet  Op = 2 // 新增或修改
)

////////////////////// PeerInfoTable //////////////////////////
// 暂时只维护在内存中

const (
	// 长度不能超过requires.CFLen (6)
	// 将种子节点与普通节点分开存储
	PeerInfoKeyPrefix    = "peers-"
	SeedInfoKeyPrefix    = "seeds-"
	DefaultMergeInterval = 5 * time.Minute
	Module_Pit           = "PIT"
)

// PeerInfoTable 节点信息表
// 节点信息表实际在使用时，一般是不会删除数据的，只会新增
// 但是后序会考虑到节点的断连、恶意等，需要删除操作
type PeerInfoTable struct {
	id string // 自己的账号

	// seeds目前设定为固定的配置，运行期间不修改其数据
	// 因此使用期间不加锁
	seeds map[string]*defines.PeerInfo
	nSeed uint32 // 必须使用atomic包的方法加载和变更

	nPeer     uint32 // 必须使用atomic包的方法加载和变更
	peers     map[string]*defines.PeerInfo
	dirty     map[string]*dirtyPeerInfo // 被修改的/新增的/删除的
	peersLock *sync.RWMutex
	dirtyLock *sync.RWMutex

	// 存储引擎
	kv requires.Store
	// 存储前缀
	peersCF requires.CF
	seedsCF requires.CF

	// 自动merge的间隔
	mergeInterval time.Duration

	inited bool // 是否已初始化
	done   chan struct{}

	*log.Logger
}

// NewPeerInfoTable
func NewPeerInfoTable(id string, kv requires.Store) (*PeerInfoTable, error) {
	logger := log.NewLogger(Module_Pit, id)
	if logger == nil {
		return nil, errors.New("nil logger, please init logger first")
	}

	return &PeerInfoTable{
		id:            id,
		seeds:         make(map[string]*defines.PeerInfo),
		peers:         make(map[string]*defines.PeerInfo),
		dirty:         make(map[string]*dirtyPeerInfo),
		peersLock:     new(sync.RWMutex),
		dirtyLock:     new(sync.RWMutex),
		kv:            kv,
		peersCF:       requires.String2CF(PeerInfoKeyPrefix),
		seedsCF:       requires.String2CF(SeedInfoKeyPrefix),
		mergeInterval: DefaultMergeInterval,
		done:          make(chan struct{}),
		Logger:        logger,
	}, nil
}

// Init NewPeerInfoTable之后需要调用Init来初始化，并启动mergeLoop
func (pit *PeerInfoTable) Init() error {

	if pit.Inited() {
		return nil
	}

	pit.Info("Init start")

	// 开启kv(与之建立连接)
	err := pit.kv.Open()
	if err != nil {
		return err
	}

	// 注册列族，若已存在则do nothing
	if err := pit.kv.RegisterCF(pit.peersCF); err != nil {
		return nil
	}
	if err := pit.kv.RegisterCF(pit.seedsCF); err != nil {
		return nil
	}

	err = pit.load()
	if err != nil {
		return err
	}

	go pit.mergeLoop()

	pit.inited = true
	pit.Info("Init succ")
	return nil
}

// Inited 是否已初始化
func (pit *PeerInfoTable) Inited() bool {
	return pit.inited
}

// load 加载
// 启动时，从pit.kv中加载所有节点(包括种子)信息到pit.peers
func (pit *PeerInfoTable) load() error {

	err := pit.kv.RangeCF(pit.seedsCF, func(key, value []byte) error {
		pi := new(defines.PeerInfo)
		err := pi.Decode(value)
		if err != nil {
			return err
		}
		pit.seeds[string(key)] = pi
		pit.nSeedIncr() // 计数加1
		return nil
	})
	if err != nil {
		return err
	}
	err = pit.kv.RangeCF(pit.peersCF, func(key, value []byte) error {
		pi := new(defines.PeerInfo)
		err := pi.Decode(value)
		if err != nil {
			return err
		}
		pit.peers[string(key)] = pi
		pit.nPeerIncr() // 计数加1
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

//// store 存储/持久化
//func (pit *PeerInfoTable) store() error {
//
//	return nil
//}

// merge 将dirty刷合并到peers去，并且先更新到kv中
// peers代表的是与kv存储引擎中一致的数据
func (pit *PeerInfoTable) merge() error {

	// 1. 将seeds合并
	// 由于seeds数量少，直接全量覆盖
	err := pit.kv.RangeCF(pit.seedsCF, func(key, value []byte) error {
		pi := new(defines.PeerInfo)
		err := pi.Decode(value)
		if err != nil {
			return err
		}
		pit.seeds[string(key)] = pi
		return nil
	})
	if err != nil {
		return err
	}
	for id := range pit.seeds {
		b, err := pit.seeds[id].Encode()
		if err != nil {
			return err
		}
		err = pit.kv.Set(pit.seedsCF, []byte(id), b)
		if err != nil {
			return err
		}
	}

	// 2. 将peers合并

	// 将dirty合并到peers
	pit.dirtyLock.RLock()
	pit.peersLock.Lock()
	for id, dinfo := range pit.dirty {
		_, ok := pit.peers[id]
		switch dinfo.op {
		case OpNone:
			return errors.New("only OpDel or OpSet is permitted")
		case OpDel:
			if ok {
				// 删除kv中该项
				err := pit.kv.Del(pit.peersCF, []byte(id))
				if err != nil {
					return err
				}
				// 删除pit.peers中该项
				delete(pit.peers, id)
			} else {
				return errors.New("OpDel occurred on a non-exist id")
			}
		case OpSet:
			// 设置kv中该项
			infobytes, err := dinfo.info.Encode()
			if err != nil {
				return err
			}
			err = pit.kv.Set(pit.peersCF, []byte(id), infobytes)
			if err != nil {
				return err
			}
			// 设置pit.peers中该项
			pit.peers[id] = dinfo.info
		}
	}
	pit.peersLock.Unlock()
	pit.dirtyLock.RUnlock()

	// 重置dirty
	pit.dirtyLock.Lock()
	pit.dirty = make(map[string]*dirtyPeerInfo)
	pit.dirtyLock.Unlock()

	return nil
}

// mergeLoop
func (pit *PeerInfoTable) mergeLoop() {
	var err error
	ticker := time.Tick(pit.mergeInterval)
	for {
		select {
		case <-ticker:
			if err = pit.merge(); err != nil {
				pit.Errorf("PeerInfoTable merge into kvstore fail: %s\n", err)
			}
			pit.Infof("PeerInfoTable merge into kvstore succ\n")
		case <-pit.done:
			if err = pit.merge(); err != nil {
				pit.Errorf("PeerInfoTable merge into kvstore fail: %s\n", err)
			}
			pit.Infof("PeerInfoTable merge into kvstore succ\n")
			return
		}
	}
}

// Get 按id查询
func (pit *PeerInfoTable) Get(id string) (*defines.PeerInfo, error) {
	// 首先看是否是seed
	if v, ok := pit.seeds[id]; ok {
		return v, nil
	}

	// 从从普通节点里找
	pit.peersLock.RLock()
	pi, ok1 := pit.peers[id]
	pit.peersLock.RUnlock()
	pit.dirtyLock.RLock()
	dpi, ok2 := pit.dirty[id]
	pit.dirtyLock.RUnlock()
	if !ok1 && !ok2 || (ok2 && dpi.op == OpDel) {
		return nil, fmt.Errorf("no such id: %s", id)
	}

	if !ok2 {
		return pi, nil
	}
	if ok2 && dpi.op == OpSet {
		pit.Debugf("PeerInfoTable Get: {id:%s, PeerInfo:%s}\n", id, dpi.info.String())
		return dpi.info, nil
	}
	return nil, errors.New("unknown error")
}

// Set 添加或修改
// 只写到dirty中
func (pit *PeerInfoTable) Set(info *defines.PeerInfo) error {

	if info == nil {
		return nil
	}

	// 如果是seed
	if info.Duty == defines.PeerDuty_Seed {
		if pit.seeds[info.Id] == nil {
			pit.nSeedIncr() // 计数加1
		}
		pit.seeds[info.Id] = info
		return nil
	}

	// 如果是普通节点

	id := info.Id

	// 先查看id是否存在
	if info, _ := pit.Get(id); info == nil {
		pit.nPeerIncr() // 计数加1
	}

	// 不论dirty中对应键值对是否存在，都直接写dirty
	pit.dirtyLock.Lock()
	pit.dirty[id] = &dirtyPeerInfo{
		op:   OpSet,
		info: info,
	}
	pit.dirtyLock.Unlock()

	pit.Debugf("PeerInfoTable Set: {id:%s, PeerInfo:%s}\n", id, info.String())
	return nil
}

func (pit *PeerInfoTable) AddPeers(peers map[string]string) error {
	for id, addr := range peers {
		info := defines.PeerInfo{
			Id:   id,
			Addr: addr,
			Attr: 0,
			Duty: defines.PeerDuty_Peer,
			Data: nil,
		}
		if err := pit.Set(&info); err != nil {
			fmt.Printf("AddPeers: err: %s\n", err)
		}
	}
	return nil
}

func (pit *PeerInfoTable) AddSeeds(seeds map[string]string) error {
	for id, addr := range seeds {
		info := defines.PeerInfo{
			Id:   id,
			Addr: addr,
			Attr: 0,
			Duty: defines.PeerDuty_Seed,
			Data: nil,
		}
		if err := pit.Set(&info); err != nil {
			fmt.Printf("AddSeeds: err: %s\n", err)
		}
	}
	return nil
}

// Del 删除
// 只写到dirty中
// 如果peers/dirty均没有该记录则报错
func (pit *PeerInfoTable) Del(id string) error {
	pit.peersLock.RLock()
	_, ok1 := pit.peers[id]
	pit.peersLock.RUnlock()
	pit.dirtyLock.RLock()
	dpi, ok2 := pit.dirty[id]
	pit.dirtyLock.RUnlock()
	if !ok1 && !ok2 || (ok2 && dpi.op == OpDel) {
		return fmt.Errorf("no such id: %s", id)
	}
	pit.dirtyLock.Lock()
	pit.dirty[id] = &dirtyPeerInfo{op: OpDel}
	pit.dirtyLock.Unlock()
	pit.nPeerDecr() // 计数减1

	pit.Debugf("PeerInfoTable Del: {id:%s}\n", id)
	return nil
}

// Close 关闭PeerInfoTable：关闭其内与kv的连接，通知mergeLoop退出
func (pit *PeerInfoTable) Close() error {
	close(pit.done)
	err := pit.kv.Close()
	if err != nil {
		return err
	}
	pit.inited = false
	return nil
}

// Peers 生成Peers的快照
func (pit *PeerInfoTable) Peers() map[string]*defines.PeerInfo {

	// 先生成peers和dirty的快照
	peersSnapshot := map[string]*defines.PeerInfo{}
	dirtySnapshot := map[string]*dirtyPeerInfo{}
	pit.peersLock.RLock()
	pit.dirtyLock.RLock()
	for id := range pit.peers {
		peersSnapshot[id] = pit.peers[id]
	}
	for id := range pit.dirty {
		dirtySnapshot[id] = pit.dirty[id]
	}
	pit.dirtyLock.RUnlock()
	pit.peersLock.RUnlock()

	// 合并
	// 这个地方的合并并不适合与pit.merge去复用代码
	// 原因是merge还需要将修改持久化，耗时较久，而且可能会失败，因此比较适合异步地进行
	// 而Peers是个同步的API，需要获取当前的所有节点信息，没必要去做数据持久化，只在内存中处理也更可控
	for id, dinfo := range dirtySnapshot {
		_, ok := peersSnapshot[id]
		switch dinfo.op {
		case OpNone:
			// 不会出现这种情况，不管它
		case OpDel:
			if ok {
				delete(peersSnapshot, id)
			}
		case OpSet:
			peersSnapshot[id] = dinfo.info
		}
	}

	return peersSnapshot
}

// Seeds 生成Seeds的快照
func (pit *PeerInfoTable) Seeds() map[string]*defines.PeerInfo {
	snapshot := map[string]*defines.PeerInfo{}
	for id := range pit.seeds {
		snapshot[id] = pit.seeds[id]
	}
	return snapshot
}

// RangePeers 对pit当前记录的所有peers执行某项操作
func (pit *PeerInfoTable) RangePeers(f func(peer *defines.PeerInfo) error) (
	total int, errs map[string]error) {

	if f == nil {
		pit.Fatalf("PeerInfoTable RangePeers fail: nil func\n")
	}

	errs = make(map[string]error)
	peers := pit.Peers()
	total = len(peers)
	for _, peer := range peers {
		peer := peer // 复制一份
		if err := f(peer); err != nil {
			errs[peer.Id] = err
		}
	}
	return total, errs
}

// RangeSeeds 对pit.seeds执行某项操作
func (pit *PeerInfoTable) RangeSeeds(f func(peer *defines.PeerInfo) error) (
	total int, errs map[string]error) {

	if f == nil {
		pit.Fatalf("PeerInfoTable RangeSeeds fail: nil func\n")
	}

	errs = make(map[string]error)
	for _, seed := range pit.seeds {
		seed := seed // 复制一份
		if err := f(seed); err != nil {
			errs[seed.Id] = err
		}
	}
	return pit.NSeed(), errs
}

// IsSeed 判断id是否是seed节点
func (pit *PeerInfoTable) IsSeed(id string) bool {
	_, ok := pit.seeds[id]
	return ok
}

func (pit *PeerInfoTable) NSeed() int {
	return int(atomic.LoadUint32(&pit.nSeed))
}

func (pit *PeerInfoTable) NPeer() int {
	return int(atomic.LoadUint32(&pit.nPeer))
}

func (pit *PeerInfoTable) nPeerIncr() {
	atomic.AddUint32(&pit.nPeer, 1)
}

func (pit *PeerInfoTable) nPeerDecr() {
	atomic.AddUint32(&pit.nPeer, ^uint32(0))
}

func (pit *PeerInfoTable) nSeedIncr() {
	atomic.AddUint32(&pit.nSeed, 1)
}

func (pit *PeerInfoTable) nSeedDecr() {
	atomic.AddUint32(&pit.nSeed, ^uint32(0))
}

///////////////////// 节点信息加载函数 //////////////////////

// 节点信息加载函数
// 外部调用者可以通过定义自己的PeerInfoLoadFunc来加载数据，具体数据从哪加载，如何加载自行解释
//type PeerInfoLoadFunc func() ([]*PeerInfo, error)
