/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/13/20 3:47 PM
* @Description: The file is for
***********************************************************************/

package bnet

import (
	"bytes"
	"errors"
	"log"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/requires"
)

// 测试用的requires.Store实现
type testStore struct {
	cfs map[requires.CF]bool
	kvs map[string]string
}

func (t *testStore) Open() error {
	log.Printf("testStore open ...\n")
	return nil
}

func (t *testStore) Close() error {
	log.Printf("testStore close ...\n")
	return nil
}

func (t *testStore) Get(cf requires.CF, key []byte) ([]byte, error) {
	k := string(cf[:]) + string(key)
	v := new(defines.PeerInfo)
	v.Decode(bytes.NewReader([]byte(t.kvs[k])))
	log.Printf("testStore Get: {key: %s, value: %s}\n", string(key), v.String())
	//time.Sleep(50 * time.Millisecond)
	return []byte(t.kvs[k]), nil
}

func (t *testStore) Set(cf requires.CF, key, value []byte) error {
	k := string(cf[:]) + string(key)
	t.kvs[k] = string(value)
	v := new(defines.PeerInfo)
	v.Decode(bytes.NewReader(value))
	log.Printf("testStore Set: {key: %s, value: %s}\n", string(key), v.String())
	//time.Sleep(300 * time.Millisecond)
	return nil
}

func (t *testStore) Del(cf requires.CF, key []byte) error {
	delete(t.kvs, string(cf[:]) + string(key))
	log.Printf("testStore Del: {key: %s}\n", string(key))
	//time.Sleep(300 * time.Millisecond)
	return nil
}

func (t *testStore) RegisterCF(cf requires.CF) error {
	t.cfs[cf] = true
	return nil
}

func (t *testStore) RangeCF(cf requires.CF, f func(key, value []byte) error) error {
	var firstErr, err error
	for k, v := range t.kvs {
		if strings.HasPrefix(k, string(cf[:])) {
			err = f([]byte(k[requires.CFLen:]), []byte(v))
			if err != nil {
				if firstErr == nil {
					firstErr = err
				} else {
					continue
				}
			}
		}
	}

	return firstErr
}

//////////////////////////////////////////////

// 测试节点信息表的创建、初始化、增删改查、持久化
func TestPeerInfoTable(t *testing.T) {
	// 创建
	tkv := &testStore{
		cfs: map[requires.CF]bool{},
		kvs: map[string]string{},
	}
	pit := NewPeerInfoTable(tkv)

	// 修改merge间隔，以使得在测试函数执行期间能够执行merge
	pit.mergeInterval = 100 * time.Millisecond

	// 初始化
	err := pit.Init()
	if err != nil {
		t.Error(err)
	}
	defer pit.Close()

	// 插入三条数据
	err = pit.Set(&defines.PeerInfo{
		Id:   "id1",
		Addr: "addr1",
	})
	handleError(t, err, pit, tkv)
	err = pit.Set(&defines.PeerInfo{
		Id:   "id2",
		Addr: "addr2",
	})
	handleError(t, err, pit, tkv)
	err = pit.Set(&defines.PeerInfo{
		Id:   "id3",
		Addr: "addr3",
	})
	handleError(t, err, pit, tkv)

	// 睡眠触发merge
	time.Sleep(150 * time.Millisecond)

	// 删除数据
	err = pit.Del("id3")
	handleError(t, err, pit, tkv)

	// 修改数据
	err = pit.Set(&defines.PeerInfo{
		Id:   "id2",
		Addr: "addr22222",
	})
	handleError(t, err, pit, tkv)

	// 查数据
	info, err := pit.Get("id2")
	handleError(t, err, pit, tkv)
	if !reflect.DeepEqual(info, &defines.PeerInfo{
		Id:   "id2",
		Addr: "addr22222",
	}) {
		t.Error("errorrrrr")
	}

	// 睡眠触发merge
	time.Sleep(150 * time.Millisecond)
}

func handleError(t *testing.T, err error, pit *PeerInfoTable, tkv *testStore) {
	if err != nil {
		t.Error(err)
	}
	if err = checkPitAndKv(pit, tkv); err != nil {
		t.Error(err)
	}
}

/////////////////////////// / ////////////////////////////

// 检查pit.peers和kv两处的数据是否保持了一致
func checkPitAndKv(pit *PeerInfoTable, kv *testStore) error {
	pit.peersLock.RLock()
	defer pit.peersLock.RUnlock()
	if len(pit.peers) != len(kv.kvs) {	// 测试过程中只有1个cf
		return errors.New("len(pit.peers) != len(kv.kvs)")
	}
	for id, pi := range pit.peers {
		api, err := kv.Get(pit.cf, []byte(id))
		if err != nil {
			return err
		}
		pib, err := pi.Encode()
		if err != nil {
			return err
		}
		if !bytes.Equal(pib, api) {
			return errors.New("!bytes.Equal(pib, api)")
		}
	}
	return nil
}