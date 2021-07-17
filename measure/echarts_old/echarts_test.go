package echarts_old_test

import (
	"testing"

	"github.com/azd1997/blockchain-consensus/measure/echarts_old"
)

func TestEchartsRun(t *testing.T) {
	host := "localhost:9999"
	echarts_old.Run(host, make(<-chan echarts_old.MeasureData))
}