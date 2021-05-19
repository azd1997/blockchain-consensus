// Package report report负责节点上报监控数据
package report

import "github.com/azd1997/blockchain-consensus/defines"

///////////////////////////// 监控数据格式 ////////////////////////////////

///////////////////////////// 全局变量 ////////////////////////////////

var (
	dest Destination
)

const (
	Destination_DefaultMonitor = iota
	Destination_WebsocketMonitor
)

///////////////////////////// 对外提供的方法 ////////////////////////////////



func SetDestination() {

}

func ReportNewBlock(nb *defines.Block) error {

	// 先向


	return nil
}

///////////////////////////// 日志输出地 ////////////////////////////////

type Destination interface {
	WriteTo(data []byte) (n int, err error)
}

// WebsocketMonitor 通过WebSocket传输给另一进程
type WebsocketMonitor struct {

}

// DefaultMonitor 默认的监控器
//
type DefaultMonitor struct {

}