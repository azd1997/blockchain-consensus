/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 5/22/21 11:24 AM
* @Description: The file is for
***********************************************************************/

package modules

type Module interface {
	Init() error
	Inited() bool
	Ok() bool // Ok 检查Net所依赖的对象是否初始化好
	Close() error
	Closed() bool
}
