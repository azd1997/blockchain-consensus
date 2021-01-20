/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 11/3/20 10:12 PM
* @Description: The file is for
***********************************************************************/

package pot

//func Test_binarySearch(t *testing.T) {
//	type args struct {
//		holes  [][2]int64
//		target int64
//	}
//	tests := []struct {
//		name string
//		args args
//		want int
//	}{
//		{"normal",
//			args{holes: [][2]int64{{1, 3}, {4, 6}, {7, 10}}, target: 5},
//			1},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			if got := binarySearch(tt.args.holes, tt.args.target); got != tt.want {
//				t.Errorf("binarySearch() = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
//
//func Test_processTable_fill(t *testing.T) {
//	type fields struct {
//		id        string
//		maxIndex  int64
//		processes map[string]*defines.Process
//		lock      *sync.RWMutex
//		holes     [][2]int64
//	}
//	type args struct {
//		bIndex int64
//	}
//	tests := []struct {
//		name   string
//		fields fields
//		args   args
//	}{
//		{"normal", fields{
//			holes: [][2]int64{{1, 3}, {4, 6}, {7, 10}},
//		}, args{bIndex: 5}},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			pt := &processTable{
//				id:        tt.fields.id,
//				maxIndex:  tt.fields.maxIndex,
//				processes: tt.fields.processes,
//				lock:      tt.fields.lock,
//				holes:     tt.fields.holes,
//			}
//			pt.fill(tt.args.bIndex)
//			t.Log(pt.holes)
//		})
//	}
//}
