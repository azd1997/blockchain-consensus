/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/14/20 11:28 AM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"github.com/azd1997/blockchain-consensus/defines"
	_default "github.com/azd1997/blockchain-consensus/requires/default"
	"github.com/azd1997/blockchain-consensus/test"
	"github.com/azd1997/blockchain-consensus/utils/log"
	"testing"
)

func TestNode_PeerInit(t *testing.T) {
	id := "peer1"
	duty := defines.PeerDuty_Peer
	addr := "127.0.0.1:7991"
	logdest := "./pot.log"

	// 初始化日志单例
	log.InitGlobalLogger(id, logdest)
	defer log.Sync()

	ln, err := _default.ListenTCP(id, addr)
	tError(t, err)
	d, err := _default.NewDialer(id, addr, 0)
	tError(t, err)

	kv := test.NewStore()
	bc := test.NewBlockChain(id)

	node, err := NewNode(
		id, duty, ln, d, kv, bc, logdest,
		map[string]string{
			"seed1": "127.0.0.1:8991",
		},
		map[string]string{
			"peer1": "127.0.0.1:7991",
			"peer2": "127.0.0.1:7992",
			"peer3": "127.0.0.1:7993",
		})
	tError(t, err)

	err = node.Init()
	tError(t, err)
}

func TestNode_SeedInit(t *testing.T) {
	id := "seed1"
	duty := defines.PeerDuty_Seed
	addr := "127.0.0.1:8991"
	logdest := "./pot-seed1.log"

	// 初始化日志单例
	log.InitGlobalLogger(id, logdest)
	defer log.Sync()

	ln, err := _default.ListenTCP(id, addr)
	tError(t, err)
	d, err := _default.NewDialer(id, addr, 0)
	tError(t, err)

	kv := test.NewStore()
	bc := test.NewBlockChain(id)

	node, err := NewNode(
		id, duty, ln, d, kv, bc, logdest,
		map[string]string{
			"seed1": "127.0.0.1:8991",
		},
		map[string]string{
			"peer1": "127.0.0.1:7991",
			"peer2": "127.0.0.1:7992",
			"peer3": "127.0.0.1:7993",
		})
	tError(t, err)

	err = node.Init()
	tError(t, err)
}

func tError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
