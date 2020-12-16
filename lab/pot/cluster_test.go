/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/15/20 12:27 AM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"testing"
	"time"
)

func TestCluster(t *testing.T) {
	c, err := StartCluster(1, 3)
	if err != nil {
		t.Error(err)
	}
	c = c

	time.Sleep(10 * time.Second)
}

func TestStartNode(t *testing.T) {
	_, _, seedsm, peersm := genIdsAndAddrs(1, 3)
	peer01 := "peer01"
	_, err := StartNode(peer01, peersm[peer01], seedsm, peersm)
	if err != nil {
		t.Error(err)
	}
}
