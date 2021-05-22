/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 5/21/21 8:41 AM
* @Description: The file is for
***********************************************************************/

package conn_net

type ConnStatus uint8

// 对于ConnStatus，由于每次更新基本是只更新一边（send或recv），所以ConnSTATUS最好是不依赖于当前是什么状态，只根据遇到什么事而更改
const (
	ConnStatus_SHIFT_Send = 0
	ConnStatus_SHIFT_Recv = 1

	ConnStatus_OnlySend ConnStatus = 0x01 << ConnStatus_SHIFT_Send
	ConnStatus_OnlyRecv ConnStatus = 0x01 << ConnStatus_SHIFT_Recv
	ConnStatus_SendRecv = ConnStatus_OnlySend | ConnStatus_OnlyRecv
	ConnStatus_Closed = 0x00

	// 0000 0011 最后一位标志send，倒数第二位标志Recv
	ConnStatus_MASK_Send = 0x01	// 0000 0001
	ConnStatus_MASK_Recv = 0x02	// 0000 0010
)

var connStatusString = map[ConnStatus]string{
	ConnStatus_Closed:  "Closed",
	ConnStatus_OnlySend: "OnlySend",
	ConnStatus_OnlyRecv:"OnlyRecv",
	ConnStatus_SendRecv:"SendRecv",
}

func (cs ConnStatus) String() string {
	return connStatusString[cs]
}

func (cs *ConnStatus) EnableSend() {
	*cs |= ConnStatus_MASK_Send
}

func (cs *ConnStatus) DisableSend() {
	*cs &= ConnStatus(^uint8(ConnStatus_MASK_Send))
}

func (cs *ConnStatus) EnableRecv() {
	*cs |= ConnStatus_MASK_Recv
}

func (cs *ConnStatus) DisableRecv() {
	*cs &= ConnStatus(^uint8(ConnStatus_MASK_Recv))
}

func (cs ConnStatus) CanSend() bool {
	cs1 := cs
	if (cs1 >> ConnStatus_SHIFT_Send) & 0x01 == 0x01 {
		return true
	}
	return false
}

func (cs ConnStatus) CanRecv() bool {
	cs1 := cs
	if (cs1 >> ConnStatus_SHIFT_Recv) & 0x01 == 0x01 {
		return true
	}
	return false
}