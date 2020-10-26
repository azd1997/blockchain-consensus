/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/28 11:10
* @Description: The file is for
***********************************************************************/

package lab111

import (
	"fmt"
	"sync"
)

// 暂时先不把公私钥体系引入
// 随便生成账号id

var number int
var lock sync.Mutex

func GenUniqueId() string {
	cur := 0
	lock.Lock()
	number++
	cur = number
	lock.Unlock()
	return fmt.Sprintf("node%d", cur)
}
