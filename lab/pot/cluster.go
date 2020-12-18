/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/14/20 11:28 PM
* @Description: 集群
***********************************************************************/

package pot

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
	_default "github.com/azd1997/blockchain-consensus/requires/default"
	"github.com/azd1997/blockchain-consensus/test"
	"github.com/azd1997/blockchain-consensus/utils/log"
)

const (
	seedIdPrefix   = "seed"
	peerIdPrefix   = "peer"
	seedAddrPrefix = "127.0.0.1:80"
	peerAddrPrefix = "127.0.0.1:90"
	logDestFormat  = "./pot-log-%s.log"
)

// Cluster 模拟数量为两位数的集群
type Cluster struct {
	seeds   map[string]*Node
	peers   map[string]*Node
	clients map[string]*test.TxMaker
}

func StartCluster(nSeed int, nPeer int, debug bool, addCaller bool, enableClients bool) (*Cluster, error) {
	seeds, peers, seedsm, peersm := genIdsAndAddrs(nSeed, nPeer)

	c := &Cluster{
		seeds: map[string]*Node{},
		peers: map[string]*Node{},
		clients: map[string]*test.TxMaker{},
	}


	for _, idaddr := range seeds {
		id, addr := idaddr[0], idaddr[1]
		node, err := StartNode(id, addr, seedsm, peersm, debug, addCaller)
		if err != nil {
			return nil, err
		}
		c.seeds[id] = node
	}

	time.Sleep(2 * time.Second)

	for _, idaddr := range peers {
		id, addr := idaddr[0], idaddr[1]
		node, err := StartNode(id, addr, seedsm, peersm, debug, addCaller)
		if err != nil {
			return nil, err
		}
		c.peers[id] = node
	}

	if enableClients {
		// 创建txmaker
		for id := range peersm {
			id := id
			tos := make([]string, 0, len(peers))
			for to := range peersm {
				if id != to {
					tos = append(tos, to)
				}
			}
			tm := test.NewTxMaker(id, tos, c.peers[id].pot.LocalTxInChan())
			c.clients[id] = tm
			go tm.Start()
		}
	}


	return c, nil
}

func StartNode(id, addr string, seeds, peers map[string]string, debug bool, addCaller bool) (*Node, error) {

	logdest := fmt.Sprintf(logDestFormat, id)

	// 初始化日志单例
	log.InitGlobalLogger(id, debug, addCaller, logdest)

	ln, err := _default.ListenTCP(id, addr)
	if err != nil {
		return nil, err
	}
	d, err := _default.NewDialer(id, addr, 0)
	if err != nil {
		return nil, err
	}
	kv := test.NewStore()
	bc := test.NewBlockChain(id)

	var duty defines.PeerDuty
	if id[:4] == "seed" {
		duty = defines.PeerDuty_Seed
	} else if id[:4] == "peer" {
		duty = defines.PeerDuty_Peer
	} else {
		return nil, errors.New("unknown duty")
	}

	node, err := NewNode(id, duty,
		ln, d, kv, bc, logdest,
		seeds, peers)
	if err != nil {
		return nil, err
	}

	err = node.Init()
	if err != nil {
		return nil, err
	}

	return node, nil
}

func genIdsAndAddrs(nSeed, nPeer int) ([][2]string, [][2]string, map[string]string, map[string]string) {
	seeds := make([][2]string, 0, nSeed)
	peers := make([][2]string, 0, nPeer)
	seedsm := make(map[string]string)
	peersm := make(map[string]string)

	if nSeed >= 100 || nPeer > 100 {
		return nil, nil, nil, nil
	}

	for i := 1; i <= nSeed; i++ {
		idnum := ""
		if i <= 9 {
			idnum = "0" + strconv.Itoa(i)
		} else {
			idnum = strconv.Itoa(i)
		}
		id := seedIdPrefix + idnum
		addr := seedAddrPrefix + idnum
		seeds = append(seeds, [2]string{id, addr})
		seedsm[id] = addr
	}

	for i := 1; i <= nPeer; i++ {
		idnum := ""
		if i <= 9 {
			idnum = "0" + strconv.Itoa(i)
		} else {
			idnum = strconv.Itoa(i)
		}
		id := peerIdPrefix + idnum
		addr := peerAddrPrefix + idnum
		peers = append(peers, [2]string{id, addr})
		peersm[id] = addr
	}

	return seeds, peers, seedsm, peersm
}
