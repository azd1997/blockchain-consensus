/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 19:52
* @Description: client客户端定义
***********************************************************************/

package bcc

// Client 客户端实现
// 客户端不关心共识集群采用何种共识
// 它只需要维护“足够的”共识节点地址，需要发起请求时向共识节点发起请求
type Client struct {
	// 维护的共识节点集合。key为服务器节点的ID，其实就是通过公私钥产生的
	Servers map[string]Consensus

	// 除此之外，Client查询区块链数据，需要根据自身是否是ClientOnly还是也部署了共识节点
	// 如果ClientOnly，查数据需要向
	//ClientOnly bool


}
