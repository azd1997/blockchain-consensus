/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/8/20 6:54 PM
* @Description: 存储节点信息
***********************************************************************/

package bnet

// PeerInfo 节点信息
type PeerInfo struct {
	Id string
	Addr string
}

// PeerInfoTable 节点信息表
type PeerInfoTable struct {

}

// Load 加载
func (pit *PeerInfoTable) Load() error {
	return nil
}

// Store 存储
func (pit *PeerInfoTable) Store() error {
	return nil
}

// Get 按id查询其监听地址
func (pit *PeerInfoTable) Get(id string) (addr string) {
	return ""
}