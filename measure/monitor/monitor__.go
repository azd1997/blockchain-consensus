// Package monitor 负责处理监控数据
package monitor

//// Monitor 监视器接口
//// 决定Monitor基本属性的有两点：
//// 1. 本地如何存储上报数据
//// 2. 本地以何种协议接收上报数据
//type Monitor interface {
//	Init() error
//	Inited() bool
//	Ok() bool // Ok 检查Net所依赖的对象是否初始化好
//	Close() error
//	Closed() bool
//
//	// CollectLoop 收集上报数据的循环
//	CollectLoop()
//	//
//	Draw()
//}
//
//type DefaultMonitor = MemoryTCPMonitor
//
//type MemoryTCPMonitor struct {
//
//}
//
//func (m *MemoryTCPMonitor) loop() {
//	// 接收到
//}