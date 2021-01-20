/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/20/21 6:13 PM
* @Description: The file is for
***********************************************************************/

package math

import (
	"math"
	"time"
)

// RoundToInt 将小数圆整为上下最接近的整数
func RoundToInt(x float64) int {
	y := math.Round(x)
	//fmt.Printf("%.1f\n", y)
	//fmt.Printf("%.0f\n", y)
	//fmt.Printf("%f\n", y)
	//fmt.Printf("%d\n", int(y))
	return int(y)
}

func RoundTickNo(bnTime, b1Time int64, tickms int) int {
	n := float64(bnTime - b1Time) / float64(int64(tickms) * int64(time.Millisecond))	// ns
	return RoundToInt(n)
}
