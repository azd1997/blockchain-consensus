/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/15/20 12:27 AM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"fmt"
	"testing"
	"time"
)

func TestCluster(t *testing.T) {
	c, err := StartCluster(1, 7, 45, 0,
		false, false, true)
	if err != nil {
		t.Error(err)
	}
	c = c

	time.Sleep(20 * time.Second)

	fmt.Println(c.DisplayAllNodes())
}

func TestStartNode(t *testing.T) {
	_, _, seedsm, peersm := GenIdsAndAddrs(1, 3)
	peer01 := "peer01"
	_, err := StartNode(peer01, peersm[peer01], 13, 0, seedsm, peersm, false, true)
	if err != nil {
		t.Error(err)
	}
}
