/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/14/20 11:28 PM
* @Description: 集群
***********************************************************************/

package pot

import (
	"errors"
	"fmt"
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"strconv"
)

func handleSignal() {
	//signal.Notify()
	//os.Signal()
	//syscall.S_BLKSIZE
}

const (
	seedIdPrefix   = "seed"
	peerIdPrefix   = "peer"
	seedAddrPrefix = "127.0.0.1:80"
	peerAddrPrefix = "127.0.0.1:90"
	logDestFormat  = "./pot-log-%s.log"
)

// Cluster 模拟数量为两位数的集群
type Cluster struct {
	seeds map[string]*Node
	peers map[string]*Node
}

// DisplayAllNodes 展示所有节点区块链进度信息
func (c *Cluster) DisplayAllNodes() (string, bool) {

	//fmt.Println(c)
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
	return res, allEqual
}

// StartCluster
// shutdownAtTiMap k>0时表示peer idno; k<0时表示seed idno。 cheatAtTiMap同理
func StartCluster(nSeed int, nPeer int,
	shutdownAtTi int, shutdownAtTiMap map[int]int, cheatAtTiMap map[int][]int,
	debug bool, addCaller bool, enableClients bool) (*Cluster, error) {

	seeds, peers, seedsm, peersm := GenIdsAndAddrs(nSeed, nPeer)

	c := &Cluster{
		seeds: map[string]*Node{},
		peers: map[string]*Node{},
	}

	for idx, idaddr := range seeds {
		go func(idx int, idaddr [2]string) {
			id, addr := idaddr[0], idaddr[1]

			fmt.Printf("\n=================== id: %s, addr: %s =======================\n\n", id, addr)

			sat, cat := shutdownAtTi, []int(nil)
			if shutdownAtTiMap != nil {
				if _, ok := shutdownAtTiMap[-idx-1]; ok {
					sat = shutdownAtTiMap[-idx-1]
				}
			}
			if cheatAtTiMap != nil && cheatAtTiMap[-idx-1] != nil {
				cat = cheatAtTiMap[-idx-1]
			}

			var node *Node
			var err error
			if id == "seed01" {
				node, err = StartNode(id, addr, sat, cat, enableClients,
					seedsm, peersm, debug, addCaller, "genesis string")
			} else {
				node, err = StartNode(id, addr, sat, cat, enableClients,
					seedsm, peersm, debug, addCaller)
			}

			if err != nil {
				panic(fmt.Sprintf("%s: %s", id, err))
			}
			c.seeds[id] = node
		}(idx, idaddr)

	}

	for idx, idaddr := range peers { // idx其实就是对应的 idno
		go func(idx int, idaddr [2]string) {
			id, addr := idaddr[0], idaddr[1]
			fmt.Printf("\n=================== id: %s, addr: %s =======================\n\n", id, addr)
			sat, cat := shutdownAtTi, []int(nil)
			if shutdownAtTiMap != nil {
				if _, ok := shutdownAtTiMap[idx+1]; ok {
					sat = shutdownAtTiMap[idx+1]
				}
			}
			if cheatAtTiMap != nil && cheatAtTiMap[idx+1] != nil {
				cat = cheatAtTiMap[idx+1]
			}
			node, err := StartNode(id, addr, sat, cat, enableClients,
				seedsm, peersm, debug, addCaller)
			if err != nil {
				panic(fmt.Sprintf("%s: %s", id, err))
			}
			c.peers[id] = node
		}(idx, idaddr)

	}

	return c, nil
}

//
func (c *Cluster) Shutdown(node string) {

}

func StartNode(id, addr string,
	shutdownAtTi int, cheatAtTi []int, enableClients bool,
	seeds, peers map[string]string, debug bool, addCaller bool,
	genesis ...string) (*Node, error) {

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

	node, err := NewNode(id, duty, addr, shutdownAtTi, cheatAtTi, enableClients, seeds, peers, genesis...)
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
