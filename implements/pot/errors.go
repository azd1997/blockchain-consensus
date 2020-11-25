/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/1/20 4:04 PM
* @Description: The file is for
***********************************************************************/

package pot

import "errors"

// ErrCannotConnectToSeedsWhenInit 无法联通种子节点
var ErrCannotConnectToSeedsWhenInit = errors.New("cannot connect to seeds when init")
