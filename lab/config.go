/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/27 22:01
* @Description: The file is for
***********************************************************************/

package lab

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	bcc "github.com/azd1997/blockchain-consensus"
	"github.com/azd1997/blockchain-consensus/lab/labrpc"
)

type config struct {
	mu sync.Mutex
	t  *testing.T

	// net其实可以看做是种子节点，因为种子节点总是最新的状态
	net *labrpc.Network

	// 集群节点数量
	n int
	// 通知节点挂掉
	done int32
	// 共识类型
	consensusType string

	// 所有共识节点的状态
	nodes map[string]*nodeConfig
}

// 节点配置
type nodeConfig struct {
	id        string
	css       bcc.Consensus
	applyErr  string
	connected bool
	saved     *bcc.ConsensusLog
	nodeLog   string            // 日志文件名
	endnames  map[string]string // 与节点相连的End列表. 代表着：id -> endname 的传输通道， v则表示用来表示二者传输的通道名/临时文件名
}

var ncpu_once sync.Once

func make_config(t *testing.T, n int, unreliable bool, consensusType string) *config {
	// CPU设置
	ncpu_once.Do(func() {
		if runtime.NumCPU() < 2 {
			fmt.Printf("warning: only one CPU, which may conceal locking bugs\n")
		}
	})
	runtime.GOMAXPROCS(4)

	// 创建测试用的Config
	cfg := &config{}
	cfg.t = t
	cfg.net = labrpc.MakeNetwork()
	cfg.n = n
	cfg.consensusType = consensusType
	cfg.nodes = make(map[string]*nodeConfig)

	// 创建n个id，及其本地日志
	for i := 0; i < n; i++ {
		id := GenUniqueId()
		nc := new(nodeConfig)
		nc.id = id
		cfg.nodes[id] = nc
	}

	cfg.setunreliable(unreliable)

	cfg.net.LongDelays(true)

	// 创建所有共识状态机
	for nodeid, nodeCfg := range cfg.nodes {
		ni, nc := nodeid, nodeCfg
		nc.nodeLog = "./id_" + ni
		nc.css = bcc.NewConsensus(consensusType)
	}

	// 连接所有节点
	for _, idnc := range cfg.nodes {
		cfg.connect(idnc.id)
	}

	return cfg
}

// 关闭某状态机服务，但是保存其状态
func (cfg *config) crash1(nodeid string) {
	// 从网络中断开
	cfg.disconnect(nodeid)
	// 删掉该服务，阻止客户端再与其建立连接
	cfg.net.DeleteServer(nodeid) // disable client connections to the server.

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	nc := cfg.nodes[nodeid]

	// a fresh persister, in case old instance
	// continues to update the Persister.
	// but copy old persister's content so that we always
	// pass Make() the last persisted state.
	if nc.saved != nil {
		nc.saved = nc.saved.Copy()
	}

	// 关闭状态机
	css := nc.css
	if css != nil {
		cfg.mu.Unlock()
		css.Kill()
		cfg.mu.Lock()
		nc.css = nil
	}

	// 保存状态
	if nc.saved != nil {
		clog := nc.saved.ReadConsensusState()
		nc.saved = &bcc.ConsensusLog{}
		nc.saved.SaveConsensusState(clog)
	}
}

//
// start or re-start a Raft.
// if one already exists, "kill" it first.
// allocate new outgoing port file names, and a new
// state persister, to isolate previous instance of
// this server. since we cannot really kill it.
//
func (cfg *config) start1(nodeid string) {
	cfg.crash1(nodeid)

	nc := cfg.nodes[nodeid]

	// a fresh set of outgoing ClientEnd names.
	// so that old crashed instance's ClientEnds can't send.
	if nc.endnames == nil {
		nc.endnames = make(map[string]string)
	}
	for _, node := range cfg.nodes { // 遍历所有节点，定义好自己去其他节点的连接名，方便之后根据该名称建立连接
		nc.endnames[node.id] = fmt.Sprintf("%s-%s", nc.id, node.id)
	}

	// a fresh set of ClientEnds.
	// 建立一批ClientEnds与之（前面的名字）对应
	ends := make(map[string]*labrpc.ClientEnd) // ends是自己向所有节点建立的单向连接集合
	for _, node := range cfg.nodes {
		ends[node.id] = cfg.net.MakeEnd(nc.endnames[node.id])
		cfg.connect(nc.endnames[node.id])
	}

	cfg.mu.Lock()

	// a fresh persister, so old instance doesn't overwrite
	// new instance's persisted state.
	// but copy old persister's content so that we always
	// pass Make() the last persisted state.
	if nc.saved != nil {
		nc.saved = nc.saved.Copy()
	} else {
		nc.saved = new(bcc.ConsensusLog)
	}

	cfg.mu.Unlock()

	// listen to messages from Raft indicating newly committed messages.
	// 新起goroutine，处理外部请求消息
	applyCh := make(chan ApplyMsg)
	go func() {
		for m := range applyCh {
			err_msg := ""
			if m.UseSnapshot {
				// ignore the snapshot
			} else if v, ok := (m.Command).(int); ok {
				cfg.mu.Lock()
				for j := 0; j < len(cfg.logs); j++ {
					if old, oldok := cfg.logs[j][m.Index]; oldok && old != v {
						// some server has already committed a different value for this entry!
						err_msg = fmt.Sprintf("commit index=%v server=%v %v != server=%v %v",
							m.Index, i, m.Command, j, old)
					}
				}
				_, prevok := cfg.logs[i][m.Index-1]
				cfg.logs[i][m.Index] = v
				cfg.mu.Unlock()

				if m.Index > 1 && prevok == false {
					err_msg = fmt.Sprintf("server %v apply out of order %v", i, m.Index)
				}
			} else {
				err_msg = fmt.Sprintf("committed command %v is not an int", m.Command)
			}

			if err_msg != "" {
				log.Fatalf("apply error: %v\n", err_msg)
				cfg.applyErr[i] = err_msg
				// keep reading after error so that Raft doesn't block
				// holding locks...
			}
		}
	}()

	rf := Make(ends, i, cfg.saved[i], applyCh)

	cfg.mu.Lock()
	cfg.rafts[i] = rf
	cfg.mu.Unlock()

	svc := labrpc.MakeService(rf)
	srv := labrpc.MakeServer()
	srv.AddService(svc)
	cfg.net.AddServer(i, srv)
}

// cleanup 杀死所有共识状态机，然后发出关闭信号
func (cfg *config) cleanup() {
	for id := range cfg.nodes {
		if _, ok := cfg.nodes[id]; ok && cfg.nodes[id].css != nil {
			cfg.nodes[id].css.Kill()
		}
	}
	atomic.StoreInt32(&cfg.done, 1)
}

// connect 将节点 nodeid 连接到共识网络net.
func (cfg *config) connect(nodeid string) {
	fmt.Printf("connect (%s)\n", nodeid)

	if _, ok := cfg.nodes[nodeid]; !ok {
		log.Printf("node <%s> doesn't exist\n", nodeid)
		return // 节点不存在，直接返回
	}

	nc := cfg.nodes[nodeid]

	nc.connected = true

	// outgoing ClientEnds
	// 将自己连向所有节点(包括自己)
	for _, idnc := range cfg.nodes {
		if idnc.connected {
			endname := nc.endnames[idnc.id]
			cfg.net.Enable(endname, true)
		}
	}

	// incoming ClientEnds
	// 将所有节点连向自己
	for _, idnc := range cfg.nodes {
		if idnc.connected {
			endname := idnc.endnames[nc.id]
			cfg.net.Enable(endname, true)
		}
	}
}

// disconnect 将nodeid这台节点移除出整个网络
func (cfg *config) disconnect(nodeid string) {
	fmt.Printf("disconnect(%s)\n", nodeid)

	nc := cfg.nodes[nodeid]

	// 标记为未连接
	nc.connected = false

	// outgoing ClientEnds
	// 将自己连向所有节点(包括自己)的连接断开
	for _, idnc := range cfg.nodes {
		if idnc.connected {
			endname := nc.endnames[idnc.id]
			cfg.net.Enable(endname, false)
		}
	}

	// incoming ClientEnds
	// 将所有节点连向自己的连接断开
	for _, idnc := range cfg.nodes {
		if idnc.connected {
			endname := idnc.endnames[nc.id]
			cfg.net.Enable(endname, false)
		}
	}
}

func (cfg *config) rpcCount(server int) int {
	return cfg.net.GetCount(server)
}

func (cfg *config) setunreliable(unrel bool) {
	cfg.net.Reliable(!unrel)
}

func (cfg *config) setlongreordering(longrel bool) {
	cfg.net.LongReordering(longrel)
}

// check that there's exactly one leader.
// try a few times in case re-elections are needed.
//func (cfg *config) checkOneLeader() int {
//	for iters := 0; iters < 10; iters++ {
//		time.Sleep(500 * time.Millisecond)
//		leaders := make(map[int][]int)
//		for i := 0; i < cfg.n; i++ {
//			if cfg.connected[i] {
//				if t, leader := cfg.rafts[i].GetState(); leader {
//					leaders[t] = append(leaders[t], i)
//				}
//			}
//		}
//		lastTermWithLeader := -1
//		for t, leaders := range leaders {
//			if len(leaders) > 1 {
//				cfg.t.Fatalf("term %d has %d (>1) leaders", t, len(leaders))
//			}
//			if t > lastTermWithLeader {
//				lastTermWithLeader = t
//			}
//		}
//
//		if len(leaders) != 0 {
//			return leaders[lastTermWithLeader][0]
//		}
//	}
//	cfg.t.Fatalf("expected one leader, got none")
//	return -1
//}

// check that everyone agrees on the term.
func (cfg *config) checkTerms() int {
	term := -1
	for i := 0; i < cfg.n; i++ {
		if cfg.connected[i] {
			xterm, _ := cfg.rafts[i].GetState()
			if term == -1 {
				term = xterm
			} else if term != xterm {
				cfg.t.Fatalf("servers disagree on term")
			}
		}
	}
	return term
}

// check that there's no leader
func (cfg *config) checkNoLeader() {
	for i := 0; i < cfg.n; i++ {
		if cfg.connected[i] {
			_, is_leader := cfg.rafts[i].GetState()
			if is_leader {
				cfg.t.Fatalf("expected no leader, but %v claims to be leader", i)
			}
		}
	}
}

// how many servers think a log entry is committed?
func (cfg *config) nCommitted(index int) (int, interface{}) {
	count := 0
	cmd := -1
	// fmt.Printf("cfg.logs: %v\n", cfg.logs)
	for i := 0; i < len(cfg.rafts); i++ {
		if cfg.applyErr[i] != "" {
			cfg.t.Fatal(cfg.applyErr[i])
		}

		cfg.mu.Lock()
		cmd1, ok := cfg.logs[i][index]
		// fmt.Printf("cfg.logs: %v\n", cfg.logs)
		// fmt.Printf("index: %d, cmd1: %d\n", index, cmd1)
		cfg.mu.Unlock()

		if ok {
			if count > 0 && cmd != cmd1 {
				cfg.t.Fatalf("committed values do not match: index %v, %v, %v\n",
					index, cmd, cmd1)
			}
			count += 1
			cmd = cmd1
		}
	}
	// fmt.Printf("count: %d, cmd: %v\n", count, cmd)
	return count, cmd
}

// wait for at least n servers to commit.
// but don't wait forever.
func (cfg *config) wait(index int, n int, startTerm int) interface{} {
	to := 10 * time.Millisecond
	for iters := 0; iters < 30; iters++ {
		nd, _ := cfg.nCommitted(index)
		if nd >= n {
			break
		}
		time.Sleep(to)
		if to < time.Second {
			to *= 2
		}
		if startTerm > -1 {
			for _, r := range cfg.rafts {
				if t, _ := r.GetState(); t > startTerm {
					// someone has moved on
					// can no longer guarantee that we'll "win"
					return -1
				}
			}
		}
	}
	nd, cmd := cfg.nCommitted(index)
	if nd < n {
		cfg.t.Fatalf("only %d decided for index %d; wanted %d\n",
			nd, index, n)
	}
	return cmd
}

// do a complete agreement.
// it might choose the wrong leader initially,
// and have to re-submit after giving up.
// entirely gives up after about 10 seconds.
// indirectly checks that the servers agree on the
// same value, since nCommitted() checks this,
// as do the threads that read from applyCh.
// returns index.
func (cfg *config) one(cmd int, expectedServers int) int {
	t0 := time.Now()
	starts := 0
	for time.Since(t0).Seconds() < 10 {
		// try all the servers, maybe one is the leader.
		index := -1
		for si := 0; si < cfg.n; si++ {
			starts = (starts + 1) % cfg.n
			var rf *Raft
			cfg.mu.Lock()
			if cfg.connected[starts] {
				rf = cfg.rafts[starts]
			}
			cfg.mu.Unlock()
			if rf != nil {
				cmdBytes := GetBytes(cmd)
				sig := signature(cmdBytes)
				index1, _, ok := rf.Start(cmd, sig)
				// fmt.Printf("index1: %d, ok: %v\n", index1, ok)
				if ok {
					index = index1
					// fmt.Printf("index: %d\n", index)
					break
				}
			}
		}

		if index != -1 {
			// somebody claimed to be the leader and to have
			// submitted our command; wait a while for agreement.
			t1 := time.Now()
			for time.Since(t1).Seconds() < 2 {
				nd, cmd1 := cfg.nCommitted(index)
				// fmt.Printf("index: %d, nd: %d. cmd1: %d\n", index, nd, cmd1)
				if nd > 0 && nd >= expectedServers {
					// committed
					if cmd2, ok := cmd1.(int); ok && cmd2 == cmd {
						// and it was the command we submitted.
						return index
					}
				}
				time.Sleep(20 * time.Millisecond)
			}
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
	cfg.t.Fatalf("one(%v) failed to reach agreement", cmd)
	return -1
}
