/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 5/21/21 8:41 AM
* @Description: The file is for
***********************************************************************/

package conn_net

type ConnStatus uint8

const (
	ConnStatus_OnlySend ConnStatus = 0x01 << iota
	ConnStatus_OnlyRecv
	ConnStatus_SendRecv = ConnStatus_OnlySend | ConnStatus_OnlyRecv
	ConnStatus_Closed = 0x00

	//ConnStatus_Ready   ConnStatus = 0
	//ConnStatus_Running ConnStatus = 1
	//ConnStatus_Closed  ConnStatus = 2
)

var connStatusString = map[ConnStatus]string{
	//ConnStatus_Ready:   "Ready",
	//ConnStatus_Running: "Running",
	ConnStatus_Closed:  "Closed",
	ConnStatus_OnlySend: "OnlySend",
	ConnStatus_OnlyRecv:"OnlyRecv",
	ConnStatus_SendRecv:"SendRecv",
}

func (cs ConnStatus) String() string {
	return connStatusString[cs]
}
