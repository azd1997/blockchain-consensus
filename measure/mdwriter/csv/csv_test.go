package csv

import (
	"github.com/azd1997/blockchain-consensus/measure/common"
	"github.com/azd1997/ego/utils"
	"path/filepath"
	"testing"
)

func TestCSVFile(t *testing.T) {
	filename := "./tmp/test.csv"
	fileDir := filepath.Dir(filename)
	if exists, _ := utils.DirExists(fileDir); !exists {
		utils.MkdirAll(fileDir)
	}

	f := NewFile(filename)
	f.MDWrite(common.GenerateRandomMD())
	f.MDWrite(common.GenerateRandomMD())
}
