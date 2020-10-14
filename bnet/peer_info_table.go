/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/8/20 6:54 PM
* @Description: 存储节点信息
***********************************************************************/

package bnet

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
	op Op
	info *defines.PeerInfo
}

// Op 对PeerInfoTable的操作 None Del Set
type Op uint8

const (
	OpNone Op = 0	// 默认值，因此每次创建DirtyPeerInfo，都需要指定Op
	OpDel Op = 1	// 删除
	OpSet Op = 2	// 新增或修改
)

////////////////////// PeerInfoTable //////////////////////////
// 暂时只维护在内存中

const (
	PeerInfoKeyPrefix = "pi"
	DefaultMergeInterval = 5 * time.Minute
)

// PeerInfoTable 节点信息表
type PeerInfoTable struct {
	peers map[string]*defines.PeerInfo
	dirty map[string]*dirtyPeerInfo	// 被修改的/新增的/删除的
	peersLock *sync.RWMutex
	dirtyLock *sync.RWMutex

	// 存储引擎
	kv requires.Store
	// 存储前缀
	cf requires.CF

	// 自动merge的间隔
	mergeInterval time.Duration

	done chan struct{}
}

// NewPeerInfoTable
func NewPeerInfoTable(kv requires.Store) *PeerInfoTable {
	return &PeerInfoTable{
		peers: make(map[string]*defines.PeerInfo),
		dirty:make(map[string]*dirtyPeerInfo),
		peersLock:    new(sync.RWMutex),
		dirtyLock:    new(sync.RWMutex),
		kv:kv,
		cf:requires.String2CF(PeerInfoKeyPrefix),
		mergeInterval:DefaultMergeInterval,
		done:make(chan struct{}),
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
// 启动时，从pit.kv中加载所有节点信息到pit.peers
func (pit *PeerInfoTable) load() error {
	return pit.kv.RangeCF(pit.cf, func(key, value []byte) error {
		pi := new(defines.PeerInfo)
		err := pi.Decode(bytes.NewReader(value))
		if err != nil {
			return err
		}
		pit.peers[string(key)] = pi
		return nil
	})
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
				err := pit.kv.Del(pit.cf, []byte(id))
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
			err = pit.kv.Set(pit.cf, []byte(id), infobytes)
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
	pit.dirty[id] = &dirtyPeerInfo{op:OpDel}
	pit.dirtyLock.Unlock()

	log.Printf("PeerInfoTable Det: {id:%s}\n", id)
	return nil
}

// Close 关闭PeerInfoTable：关闭其内与kv的连接，通知mergeLoop退出
func (pit *PeerInfoTable) Close() error {
	close(pit.done)
	return pit.kv.Close()
}

///////////////////// 节点信息加载函数 //////////////////////

// 节点信息加载函数
// 外部调用者可以通过定义自己的PeerInfoLoadFunc来加载数据，具体数据从哪加载，如何加载自行解释
//type PeerInfoLoadFunc func() ([]*PeerInfo, error)
