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
	"github.com/azd1997/blockchain-consensus/log"
	"github.com/azd1997/blockchain-consensus/test"
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

// 展示所有节点区块链进度信息
func (c *Cluster) DisplayAllNodes() string {

	var ref string   // 参照
	allEqual := true // 所有节点是否有相同的区块链
	res := "\nThe Cluster:"
	for id := range c.seeds { // 遍历种子节点集合
		bc := c.seeds[id].bc.Display()
		if ref == "" {
			ref = bc[19:] // 去掉前面的 BlockChain(peer01) 不比较
		}
		if ref != bc[19:] {
			allEqual = false
		}
		res += bc
	}
	for id := range c.peers { // 遍历种子节点集合
		bc := c.peers[id].bc.Display()
		if ref == "" {
			ref = bc[19:] // 去掉前面的 BlockChain(peer01) 不比较
		}
		if ref != bc[19:] {
			allEqual = false
		}
		res += bc
	}
	res += fmt.Sprintf("allEqual = %v\n", allEqual)
	return res
}

func StartCluster(nSeed int, nPeer int, shutdownAtTi, cheatAtTi int,
	debug bool, addCaller bool, enableClients bool) (*Cluster, error) {

	seeds, peers, seedsm, peersm := GenIdsAndAddrs(nSeed, nPeer)

	c := &Cluster{
		seeds:   map[string]*Node{},
		peers:   map[string]*Node{},
		clients: map[string]*test.TxMaker{},
	}

	for _, idaddr := range seeds {
		id, addr := idaddr[0], idaddr[1]
		node, err := StartNode(id, addr, shutdownAtTi, cheatAtTi,
			seedsm, peersm, debug, addCaller)
		if err != nil {
			return nil, err
		}
		c.seeds[id] = node
	}

	time.Sleep(2 * time.Second)

	for _, idaddr := range peers {
		id, addr := idaddr[0], idaddr[1]
		node, err := StartNode(id, addr, shutdownAtTi, cheatAtTi,
			seedsm, peersm, debug, addCaller)
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

//
func (c *Cluster) Shutdown(node string) {

}

func StartNode(id, addr string, shutdownAtTi, cheatAtTi int,
	seeds, peers map[string]string, debug bool, addCaller bool) (*Node, error) {

	logdest := fmt.Sprintf(logDestFormat, id)

	// 初始化日志单例
	log.InitGlobalLogger(id, debug, addCaller, logdest)

	var duty defines.PeerDuty
	if id[:4] == "seed" {
		duty = defines.PeerDuty_Seed
	} else if id[:4] == "peer" {
		duty = defines.PeerDuty_Peer
	} else {
		return nil, errors.New("unknown duty")
	}

	node, err := NewNode(id, duty, addr, shutdownAtTi, cheatAtTi,
		seeds, peers)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func GenIdsAndAddrs(nSeed, nPeer int) ([][2]string, [][2]string, map[string]string, map[string]string) {
	seeds := make([][2]string, 0, nSeed)
	peers := make([][2]string, 0, nPeer)
	seedsm := make(map[string]string)
	peersm := make(map[string]string)

	if nSeed >= 100 || nPeer > 100 {
		return nil, nil, nil, nil
	}

	for i := 1; i <= nSeed; i++ {
		idnum := Idnum(i)
		id := seedIdPrefix + idnum
		addr := seedAddrPrefix + idnum
		seeds = append(seeds, [2]string{id, addr})
		seedsm[id] = addr
	}

	for i := 1; i <= nPeer; i++ {
		idnum := Idnum(i)
		id := peerIdPrefix + idnum
		addr := peerAddrPrefix + idnum
		peers = append(peers, [2]string{id, addr})
		peersm[id] = addr
	}

	return seeds, peers, seedsm, peersm
}

// Idnum i范围1~99
func Idnum(i int) string {
	idnum := ""
	if i <= 9 {
		idnum = "0" + strconv.Itoa(i)
	} else {
		idnum = strconv.Itoa(i)
	}
	return idnum
}
