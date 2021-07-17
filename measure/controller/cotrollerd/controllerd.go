package cotrollerd

import (
	"github.com/azd1997/blockchain-consensus/modules/bnet"
)

// Controller 控制器服务
// 一方面与共识集群保持连接；一方面接受cli的消息
type Controller struct {
	id string
	host string
	server bnet.BNet
}

func Run(id, host string) {

}
