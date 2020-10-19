/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/8/20 6:54 PM
* @Description: 节点信息表
***********************************************************************/

package peerinfo

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
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
)

// PeerInfoTable 节点信息表
// 节点信息表实际在使用时，一般是不会删除数据的，只会新增
// 但是后序会考虑到节点的断连、恶意等，需要删除操作
type PeerInfoTable struct {
	// seeds目前设定为固定的配置，运行期间不修改其数据
	// 因此使用期间不加锁
	seeds map[string]*defines.PeerInfo

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

	done chan struct{}
}

// NewPeerInfoTable
func NewPeerInfoTable(kv requires.Store) *PeerInfoTable {
	return &PeerInfoTable{
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
	}
}

// Init NewPeerInfoTable之后需要调用Init来初始化，并启动mergeLoop
func (pit *PeerInfoTable) Init() error {
	// 开启kv(与之建立连接)
	err := pit.kv.Open()
	if err != nil {
		return err
	}

	err = pit.load()
	if err != nil {
		return err
	}

	go pit.mergeLoop()

	return nil
}

// load 加载
// 启动时，从pit.kv中加载所有节点(包括种子)信息到pit.peers
func (pit *PeerInfoTable) load() error {
	f := func(key, value []byte) error {
		pi := new(defines.PeerInfo)
		err := pi.Decode(bytes.NewReader(value))
		if err != nil {
			return err
		}
		pit.peers[string(key)] = pi
		return nil
	}

	err := pit.kv.RangeCF(pit.seedsCF, f)
	if err != nil {
		return err
	}
	err = pit.kv.RangeCF(pit.peersCF, f)
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
				log.Printf("PeerInfoTable merge into kvstore fail: %s\n", err)
			}
			log.Printf("PeerInfoTable merge into kvstore succ\n")
		case <-pit.done:
			if err = pit.merge(); err != nil {
				log.Printf("PeerInfoTable merge into kvstore fail: %s\n", err)
			}
			log.Printf("PeerInfoTable merge into kvstore succ\n")
			return
		}
	}
}

// Get 按id查询
func (pit *PeerInfoTable) Get(id string) (*defines.PeerInfo, error) {
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
		log.Printf("PeerInfoTable Get: {id:%s, PeerInfo:%s}\n", id, dpi.info.String())
		return dpi.info, nil
	}
	return nil, errors.New("unknown error")
}

// Set 添加或修改
// 只写到dirty中
func (pit *PeerInfoTable) Set(info *defines.PeerInfo) error {
	id := info.Id
	// 不论dirty中对应键值对是否存在，都直接写dirty
	pit.dirtyLock.Lock()
	pit.dirty[id] = &dirtyPeerInfo{
		op:   OpSet,
		info: info,
	}
	pit.dirtyLock.Unlock()

	log.Printf("PeerInfoTable Set: {id:%s, PeerInfo:%s}\n", id, info.String())
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

	log.Printf("PeerInfoTable Det: {id:%s}\n", id)
	return nil
}

// Close 关闭PeerInfoTable：关闭其内与kv的连接，通知mergeLoop退出
func (pit *PeerInfoTable) Close() error {
	close(pit.done)
	return pit.kv.Close()
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
func (pit *PeerInfoTable) RangePeers(f func(peer *defines.PeerInfo) error) error {
	var firstErr, err error
	peers := pit.Peers()
	for _, peer := range peers {
		peer := peer // 复制一份
		err = f(peer)
		if err != nil {
			if firstErr != nil {
				continue
			} else {
				firstErr = err
			}
		}
	}
	return firstErr
}

// RangeSeeds 对pit.seeds执行某项操作
func (pit *PeerInfoTable) RangeSeeds(f func(peer *defines.PeerInfo) error) error {
	var firstErr, err error
	for _, seed := range pit.seeds {
		seed := seed // 复制一份
		err = f(seed)
		if err != nil {
			if firstErr != nil {
				continue
			} else {
				firstErr = err
			}
		}
	}
	return firstErr
}

///////////////////// 节点信息加载函数 //////////////////////

// 节点信息加载函数
// 外部调用者可以通过定义自己的PeerInfoLoadFunc来加载数据，具体数据从哪加载，如何加载自行解释
//type PeerInfoLoadFunc func() ([]*PeerInfo, error)
