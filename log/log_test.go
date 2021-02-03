/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/31/20 2:54 PM
* @Description: The file is for
***********************************************************************/

package log

import (
	"fmt"
	"testing"
)

func TestLogger(t *testing.T) {
	InitGlobalLogger("testid", true, true, "./test.log")
	defer Sync("testid")

	l := NewLogger("TES", "testid")

	fmt.Println(l.prefix)

	l.Info("test")
	l.Infof("test: %s, %d, %v\n", "hello", 3, []byte{0})
	l.Infow("test infow", "err", "some err")

	//l.Error("test")
	//l.Errorf("test: %s, %d, %v\n", "hello", 3, []byte{0})
	//
	//format := "had [%d], %s\n"
	//af := fmt.Sprintf("has: %s", format)
	//log.Printf(af, 3, "aaa")
	//fmt.Printf(af, 3, "aaa")
}

func Test2Loggers(t *testing.T) {
	InitGlobalLogger("id1", true, true)
	defer Sync("id1")
	InitGlobalLogger("id2", true, true)
	defer Sync("id2")

	l1 := NewLogger("TES", "id1")
	l2 := NewLogger("TES", "id2")
	l1.Info("test1")
	l2.Info("test2")
}
