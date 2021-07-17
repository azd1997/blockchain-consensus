package mdwriter

import "github.com/azd1997/blockchain-consensus/measure/common"

// MDWriter 测量数据MeasureData的writer
type MDWriter interface {
	MDWrite(md common.MeasureData) error
}
