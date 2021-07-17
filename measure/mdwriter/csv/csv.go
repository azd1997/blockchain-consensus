package csv

import (
	"encoding/csv"
	"errors"
	"os"

	"github.com/azd1997/blockchain-consensus/measure/common"
)

type File struct {
	file string
}

func NewFile(filename string) *File {
	// 创建文件。如果filename存在，那么会清空原文件内容
	//
	f, err := os.Create(filename)
	if nil != err {
		panic(err)
	}
	defer f.Close()

	// 写入header
	err = mdWrite(f, common.MeasureData{}, true)
	if nil != err {
		panic(err)
	}

	return &File{
		file: filename,
	}
}

// MDWrite 写数据到文件。每写一次就保存一次，避免处理程序异常崩溃的情况
func (f *File) MDWrite(md common.MeasureData) error {
	file, err := os.OpenFile(f.file, os.O_WRONLY | os.O_APPEND, 0666)
	if nil != err {
		panic(err)
	}
	defer file.Close()

	err = mdWrite(file, md, false)
	if nil != err {
		panic(err)
	}
	return nil
}

func mdWrite(file *os.File, md common.MeasureData, header bool) error {
	if file == nil {
		return errors.New("file == nil")
		//log.Fatalln("error writing csv: ")
	}
	w := csv.NewWriter(file)

	var row []string
	if header {
		row = md.CSVHeader()
	} else {
		row = md.CSV()
	}

	w.WriteAll([][]string{row}) // calls Flush internally

	if err := w.Error(); err != nil {
		return err
		//log.Fatalln("error writing csv:", err)
	}
	return nil
}