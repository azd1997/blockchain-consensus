/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/31/20 2:54 PM
* @Description: The file is for
***********************************************************************/

package log

import (
	"fmt"
	"log"
	"testing"
)

func TestLogger(t *testing.T) {
	l := NewLogger(LogDest_Stderr, "TES", "testid")

	fmt.Println(l.logger.Prefix())

	l.Log("test")
	l.Logf("test: %s, %d, %v\n", "hello", 3, []byte{0})

	l.Error("test")
	l.Errorf("test: %s, %d, %v\n", "hello", 3, []byte{0})

	format := "had [%d], %s\n"
	af := fmt.Sprintf("has: %s", format)
	log.Printf(af, 3, "aaa")
	fmt.Printf(af, 3, "aaa")
}
