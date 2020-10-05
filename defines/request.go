/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 2020/9/22 10:59
* @Description: The file is for
***********************************************************************/

package defines


type RequestType uint8

const (
	RequestType_Blocks RequestType = 0
	RequestType_Neighbors RequestType = 1
)

type Request struct {
	Type RequestType

	// 根据index区间请求
	IndexStart uint64
	IndexCount uint64

	// 根据哈希请求
	Hashes [][]byte
}
