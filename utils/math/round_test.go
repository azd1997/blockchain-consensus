/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/20/21 6:21 PM
* @Description: The file is for
***********************************************************************/

package math

import (
	"testing"
	"time"
)

func TestRoundToInt(t *testing.T) {
	tests := []struct {
		x float64
		y int
	}{
		{12.1, 12},
		{11.9, 12},
		{12.001, 12},
		{11.877, 12},
		{12.0, 12},
		{11.5, 12},	// special
	}
	for _, tt := range tests {
		if y := RoundToInt(tt.x); y != tt.y {
			t.Errorf("error. tt.x=%f, tt.y=%d, but y=%d\n", tt.x, tt.y, y)
		}
	}
}

func TestRoundTickNo(t *testing.T) {
	now := time.Now().UnixNano()
	tickms := 500
	tick := int64(time.Duration(tickms) * time.Duration(time.Millisecond))

	tests := []struct {
		tn, t1 int64
		y int
	}{
		{now, now-3*tick, 3},
	}
	for _, tt := range tests {
		if y := RoundTickNo(tt.tn, tt.t1, 500); y != tt.y {
			t.Errorf("error. tt.tn=%d, tt.t1=%d, tt.y=%d, but y=%d\n", tt.tn, tt.t1, tt.y, y)
		}
	}
}