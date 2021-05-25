/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/15/20 12:27 AM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"fmt"
	"github.com/azd1997/blockchain-consensus/log"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"
)

// 注意：由于测试阶段，所有数据都是维护在内存中，所以交易不能产生过多，节点数量不能过大，否则会爆内存
// 而且交易数的话，可能会超出5M的区块大小限制，使得广播区块失败。 记得调整UDP缓冲大小

var (
	monitorId = "Monitor"
	monitorHost = "127.0.0.1:9998"
)

func TestCluster(t *testing.T) {
	// 清理当前目录下的日志文件
	if err := clearLogsInCurDir(); err != nil {
		panic(err)
	}

	// 运行集群
	nSeed := 1
	nPeer := 3
	shutdownAtTi := 101
	var cheatAtTiMap map[int][]int = map[int][]int{
		//1: []int{13}, // peer03在t19时伪造证明。 这会导致比E1区块链少出1个区块
	} // 这两张表用于为部分节点设置提前关闭/作弊等行为
	var shutdownAtTiMap map[int]int = map[int]int{
		//1:13,	// 设置第3号peer t19关闭
	}
	E := 1 // E1，全部正常； E2，某节点中间断线； E3，某节点中间作弊
	c, err := StartCluster(nSeed, nPeer, monitorId, monitorHost, shutdownAtTi, shutdownAtTiMap, cheatAtTiMap,
		true, true, true)
	// shutdownAtTiMap优先级比shutdownAtTi高，如果未设置Map，那么所有节点都按照shutdownAtTi关闭
	if err != nil {
		t.Error(err)
		return
	}

	// 等待一定时间后报告集群区块链状态
	time.Sleep(20 * time.Second)
	str, allEqual := c.DisplayAllNodes()
	fmt.Println(str)
	if !allEqual {
		t.Error(allEqual)
	}

	for _, seed := range c.seeds {
		log.Close(seed.id)
	}
	for _, peer := range c.peers {
		log.Close(peer.id)
	}

	// 将本次生成的日志文件转移到文件夹中
	now := time.Now()
	dirname := fmt.Sprintf("./log__%02d-%02d-%02d-%02d-%02d-%02d__%dseed-%dpeer-S%d-E%d-%v",
		now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second(),
		nSeed, nPeer, shutdownAtTi, E, allEqual) // Si表示在ti时刻定时关闭集群所有节点
	if err := createDirAndMoveLogs(dirname); err != nil {
		panic(err)
	}
}

func TestStartNode(t *testing.T) {
	_, _, seedsm, peersm := GenIdsAndAddrs(1, 3)
	peer01 := "peer01"
	_, err := StartNode(peer01, peersm[peer01], monitorId, monitorHost, 13, nil,
		false, seedsm, peersm, false, true)
	if err != nil {
		t.Error(err)
	}
}

// 将"./"目录下的".log"文件迁移到dir中
func createDirAndMoveLogs(dirname string) error {
	if err := os.MkdirAll(dirname, 0777); err != nil {
		return err
	}
	cur, err := os.Getwd() // wd work dir
	if err != nil {
		return err
	}
	flist, err := ioutil.ReadDir(cur)
	if err != nil {
		return err
	}
	for _, v := range flist {
		if strings.HasSuffix(v.Name(), ".log") {
			oldpath := "./" + v.Name()
			newpath := dirname + "/" + v.Name()

			fmt.Println(v.Name(), oldpath, newpath)
			if err := os.Rename(oldpath, newpath); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// 将当前目录下所有"./log"删掉
func clearLogsInCurDir() error {
	cur, err := os.Getwd() // wd work dir
	if err != nil {
		return err
	}
	flist, err := ioutil.ReadDir(cur)
	if err != nil {
		return err
	}
	for _, v := range flist {
		//fmt.Println(v.Name())
		if strings.HasSuffix(v.Name(), ".log") {
			fmt.Println(v.Name()) // Name就是文件名，不带路径
			if err := os.Remove("./" + v.Name()); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// 递归删除目录
//func delDir(dirpath string) error {
//	if os.Stat(dirpath)
//}

func TestUtils1(t *testing.T) {
	if err := clearLogsInCurDir(); err != nil {
		t.Error(err)
	}
}

func TestUtils2(t *testing.T) {

	//  创建一个.log文件
	f, err := os.Create("./test.log")
	if err != nil {
		t.Error(err)
		return
	}
	f.Close()

	// 移动
	if err := createDirAndMoveLogs("./test"); err != nil {
		t.Error(err)
	}

	//if err := os.Remove("./test"); err != nil {
	//	t.Error(err)
	//}
}
