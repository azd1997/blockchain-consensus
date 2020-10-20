/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/19/20 5:34 PM
* @Description: 区块链进度
***********************************************************************/

package defines


// Process 进度
type Process struct {
	Index uint64
	Hash []byte
	LatestMaker string
}