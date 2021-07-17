// Package controller 作为TCP客户端与节点通信修改节点的配置
// 正常来说Controller只需要作为客户端去尝试通知共识节点
// 但是这种做法需要Controller知道所有节点。做法可以有：
// 1. Controller请求Seed，拿到所有节点表信息
// 2. Controller作为服务器，让所有共识节点主动连接Controller，Controller再发送命令
//
// 这里倾向于方案2。把Controller作为一种服务，另外起一个cli工具
package controller

type Controller struct {

}