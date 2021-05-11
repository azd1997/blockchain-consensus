package localnet

// Monitor 监视本地网络所有节点的进展情况
// 采取节点将进度报告写到一个公共数据库，Monitor轮询该数据库的做法
type Monitor struct {
	db
}
