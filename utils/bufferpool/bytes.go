/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/9/20 9:49 AM
* @Description: 缓冲区复用
***********************************************************************/

package bufferpool

import (
	"bytes"
	"sync"
)

// 缓冲池
var bp = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// Get 从缓冲池获取一份干净的*bytes.Buffer
func Get() *bytes.Buffer {
	res := bp.Get().(*bytes.Buffer)
	res.Reset()
	return res
}

// Return 归还缓冲
func Return(buf *bytes.Buffer) {
	bp.Put(buf)
}