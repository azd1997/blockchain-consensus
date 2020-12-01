/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/10/20 2:01 PM
* @Description: 默认的requires中网络相关接口的实现
***********************************************************************/

package bcc

import (
	"time"

	"github.com/azd1997/blockchain-consensus/requires"
	_default "github.com/azd1997/blockchain-consensus/requires/default"
)

// DefaultListener 默认的Listener
func DefaultListener(id, addr string) (requires.Listener, error) {
	return _default.ListenTCP(id, addr)
}

// DefaultDialer 默认的Dialer
func DefaultDialer(id, addr string) (requires.Dialer, error) {
	return _default.NewDialer(id, addr, 0)
}

// DefaultDialerTimeout 默认的Dialer，带超时
func DefaultDialerTimeout(id, addr string, timeout time.Duration) (requires.Dialer, error) {
	return _default.NewDialer(id, addr, timeout)
}
