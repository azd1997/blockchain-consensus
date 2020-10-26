/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/21/20 11:36 AM
* @Description: 测试网络
***********************************************************************/

package lab

import bcc "github.com/azd1997/blockchain-consensus"

// Network 测试网络
type Network struct {
	nodes map[string]*bcc.Node
}

// 启动测试网络
func (tn *Network) Init() {

}

func (tn *Network) AddNode() {

}

func (tn *Network) KillNode() {

}

func (tn *Network) Disconnect(node1, node2 string) {

}

func (tn *Network) Connect(node1, node2 string) {

}

func (tn *Network) SetRandomLongDelay(yes bool) {

}

func (tn *Network) SetNetAbnormal(yes bool) {

}

// 将节点node设置成恶意节点/或非恶意节点
// TODO： 恶意行为有三种类型：
// 		1. 没有按照网络的协议去消息广播/回复/转发
//		2. 恶意篡改自身的广播消息迷惑其他节点
//		3. 恶意篡改自身转发的消息
func (tn *Network) SetMalicious(node string, yes bool) {

}

// 设置节点node随机间隔地生产tx并广播给它的节点表
func (tn *Network) SetRandomGenTx(node string, yes bool) {

}



