package common

import (
	"fmt"
	"testing"
)

func TestMeasureData_CSV(t *testing.T) {
	md := MeasureData{
		BlockKey: "xxx",
		BlockTime: 123,
		RangeBlockNum: 20,
	}	// 默认值
	csvHeader := md.CSVHeader()
	csv := md.CSV()
	fmt.Println(csvHeader)
	fmt.Println(csv)
}
