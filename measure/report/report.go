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

// ReportNewTx 汇报新交易，只统计数字
func ReportNewTx(tx *defines.Transaction) error {

}

///////////////////////////// 日志输出地 ////////////////////////////////

type Destination interface {
	ReportNewBlock(nb *defines.Block) error
	// ReportNewTx 以后可能要按照类别统计数量，所以把tx传入。暂时不使用
	ReportNewTx(tx *defines.Transaction) error
}

// LocalMemoryMonitor 本机内存监视器，监视器代码也一并写在此处
type LocalMemoryMonitor struct {

}

func (l *LocalMemoryMonitor) ReportNewBlock(nb *defines.Block) error {
	panic("implement me")
}

func (l *LocalMemoryMonitor) ReportNewTx(tx *defines.Transaction) error {
	panic("implement me")
}

// WebsocketMonitor 通过WebSocket传输给另一进程
type WebsocketMonitor struct {

}

// DefaultMonitor 默认的监控器
//
type DefaultMonitor struct {

}

