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
	InitGlobalLogger("testid", "./test.log")
	defer Sync()

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
