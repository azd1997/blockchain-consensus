/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/14/20 8:44 PM
* @Description: 节点信息表模块提供的API
***********************************************************************/

/*
	用法：
	1. 初始化 err := peerinfo.Init(kv)
	2. 调用其他API 如：info, err := peerinfo.Get(id)
*/

package peerinfo

import (
	"github.com/azd1997/blockchain-consensus/modules/peerinfo/simplepit"
	"github.com/azd1997/blockchain-consensus/requires"
)

// 节点信息表全局单例
var pit Pit

// Init 初始化节点信息表
// 在使用本文件其他API之前，必须调用Init初始化pit
func Init(id string, kv requires.Store) error {
	var err error
	pit, err := simplepit.NewSimplePit(id)
	if err != nil {
		return err
	}
	return pit.Init()
}

// Global 获取节点信息表全局单例
func Global() Pit {
	return pit
}

//// Get 查询id的节点信息
//func Get(id string) (*defines.PeerInfo, error) {
//	return pit.Get(id)
//}
//
//// Del 删除id的节点信息
//func Del(id string) error {
//	return pit.Del(id)
//}
//
//// Set 新增或修改节点信息
//func Set(info *defines.PeerInfo) error {
//	return pit.Set(info)
//}
//
//// Close 关闭pit
//func Close() error {
//	return pit.Close()
//}
//
//// Peers 生成Peers的快照
//func Peers() map[string]*defines.PeerInfo {
//	return pit.Peers()
//}
//
//// Seeds 生成Seeds的快照
//func Seeds() map[string]*defines.PeerInfo {
//	return pit.Seeds()
//}
//
//// RangePeers 对pit当前记录的所有peers执行某项操作
//func RangePeers(f func(peer *defines.PeerInfo) error) (total int, errs map[string]error) {
//	return pit.RangePeers(f)
//}
//
//// RangeSeeds 对pit.seeds执行某项操作
//func RangeSeeds(f func(peer *defines.PeerInfo) error) (total int, errs map[string]error) {
//	return pit.RangeSeeds(f)
//}
//
//func NSeed() int {
//	return pit.NSeed()
//}
//
//func NPeer() int {
//	return pit.NPeer()
//}
