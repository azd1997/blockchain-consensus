/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/20 20:50
* @Description: The file is for
***********************************************************************/

package defines

// Version 共识节点版本
// 不同共识协议的节点版本号不能放在一起比较
type Version uint8


const (
	CodeVersion Version = 0x0
)