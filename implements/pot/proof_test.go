/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/18/20 12:03 AM
* @Description: The file is for
***********************************************************************/

package pot

import "testing"


//var testProofs = []*Proof{
//	&Proof{
//		Id:        "",
//		TxsNum:    0,
//		BlockHash: nil,
//		Base:      nil,
//		BaseIndex: 0,
//	},
//	&Proof{
//		Id:        "",
//		TxsNum:    0,
//		BlockHash: nil,
//		Base:      nil,
//		BaseIndex: 0,
//	},
//	&Proof{
//		Id:        "",
//		TxsNum:    0,
//		BlockHash: nil,
//		Base:      nil,
//		BaseIndex: 0,
//	}
//}

// 如果TxsNum和BlockHash不相等，那么这样的比较结果肯定是确定的，没必要测试
// 这里的测试主要是测试：当这两项相同时，能否利用id和index来制造公平的比较？
//
// 1. 利用id的奇偶性来交换id的大小比较。 假如说有id1>id2>id3
//	index为奇：
//		在seed1收到三个proof的顺序是 id1,id2,id3 ，那么seed最后得到的winner=id1
//      在seed2收到三个proof的顺序是 id3,id2,id1 ，那么seed最后得到的winner=id1
// 		看上去没啥问题，无论见证者有多少，都只会得到id1是winner
// 	index为偶。 结果会反过来，所有人都会得到winner是id3
//  总结，这个做法不可取，会造成不公平
//
// 2. 第二个做法是，利用 index % len(id) 确定id比较的起始位。这个做法是可行的
// 从公平性来讲，对于整个id的可能域来说，是完全公平的
// 但是，对于实际系统而言，节点数就只有几个、几十个而言，是已经确定的，而且很容易偏颇。
// 也就是说，这个方案仍不太公平
//
// 3. 在2的基础上，要想公平，只能尽量使确定的id变得不确定，变得概率均匀。
// 可以采取将"id+index"加盐再哈希的做法。 这样已经概率公平了。都不需要再使用前面的“起始位”控制

// 测试两个proof在1次
func TestProofCompare1(t *testing.T) {

}
