/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/1/20 4:04 PM
* @Description: The file is for
***********************************************************************/

package pot

import (
	"errors"
)

// ErrCannotConnectToSeedsWhenInit 无法联通种子节点
var ErrCannotConnectToSeedsWhenInit = errors.New("cannot connect to seeds when init")

// MsgHandleError 消息处理错误
//type MsgHandleError struct {
//	Duty defines.PeerDuty
//	State StateType
//	Msg *defines.Message
//	Type ErrorType
//	Indexes []int	// 出问题的
//	Result string
//}
//
//func (e MsgHandleError) Error() string {
//	return fmt.Sprintf("%s handle msg")
//}
//
//type ErrorType uint8
//
//const (
//	ErrMsgType ErrorType = iota
//	ErrEntType
//	ErrReqType
//)
