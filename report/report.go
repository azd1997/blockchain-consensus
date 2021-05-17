// Package report report负责节点上报监控数据
package report

import "github.com/azd1997/blockchain-consensus/defines"

///////////////////////////// 监控数据格式 ////////////////////////////////



///////////////////////////// 对外提供的方法 ////////////////////////////////

func ReportNewBlock(nb *defines.Block) error {
	return nil
}