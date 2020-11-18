/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/17/20 11:15 AM
* @Description: The file is for
***********************************************************************/

package config

import (
	"fmt"
	"testing"
)

func TestConfig_Basic(t *testing.T) {
	// 初始化
	Init("./config.toml")
	// 查看配置
	fmt.Println(Global())
}