package echarts_test

import (
	"testing"

	"github.com/azd1997/blockchain-consensus/measure/mdwriter/echarts"
)

func TestEchartsRun(t *testing.T) {
	host := "localhost:9999"
	echarts.Run(host)
}