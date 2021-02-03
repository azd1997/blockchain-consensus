/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/14/20 11:28 AM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"github.com/azd1997/blockchain-consensus/defines"
	"github.com/azd1997/blockchain-consensus/log"
	"testing"
)

func TestNode_PeerInit(t *testing.T) {
	id := "peer1"
	duty := defines.PeerDuty_Peer
	addr := "127.0.0.1:7991"
	logdest := "./pot.log"

	// 初始化日志单例
	log.InitGlobalLogger(id, true, true, logdest)
	defer log.Sync(id)

	node, err := NewNode(
		id, duty, addr, 13, nil, false,
		map[string]string{
			"seed1": "127.0.0.1:8991",
		},
		map[string]string{
			"peer1": "127.0.0.1:7991",
			"peer2": "127.0.0.1:7992",
			"peer3": "127.0.0.1:7993",
		})
	tError(t, err)
	node = node
}

func TestNode_SeedInit(t *testing.T) {
	id := "seed1"
	duty := defines.PeerDuty_Seed
	addr := "127.0.0.1:8991"
	logdest := "./pot-seed1.log"

	// 初始化日志单例
	log.InitGlobalLogger(id, true, true, logdest)
	defer log.Sync(id)

	node, err := NewNode(
		id, duty, addr, 13, nil, false,
		map[string]string{
			"seed1": "127.0.0.1:8991",
		},
		map[string]string{
			"peer1": "127.0.0.1:7991",
			"peer2": "127.0.0.1:7992",
			"peer3": "127.0.0.1:7993",
		})
	tError(t, err)
	node = node
}

func tError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
