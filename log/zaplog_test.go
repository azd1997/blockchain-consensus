/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/14/20 8:05 PM
* @Description: The file is for
***********************************************************************/

package log

import "testing"

func TestZapLogger(t *testing.T) {
	InitGlobalLogger("id1", true, true, "./test.log")
	defer Sync()

	loggers["id1"].Debug("test sugar log")
	loggers["id1"].Error("some error")

	loggers["id1"].Debugw("some msg", "module", "TESTLOG")
	loggers["id1"].Debugw("[TEST]\tsome msg", "module", "TESTLOG")
}
