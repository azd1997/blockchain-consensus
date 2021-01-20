/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/20/21 6:21 PM
* @Description: The file is for
***********************************************************************/

package math

import "testing"

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