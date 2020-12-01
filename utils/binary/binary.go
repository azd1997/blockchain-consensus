/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/10/20 7:37 PM
* @Description: The file is for
***********************************************************************/

package binary

import (
	"encoding/binary"
	"io"
)

// 字节序
var (
	BigEndian    = binary.BigEndian
	LittleEndian = binary.LittleEndian
)

// Read binary读，传入的data必须是定长类型的指针或定长类型的slice
func Read(r io.Reader, order binary.ByteOrder, data ...interface{}) error {
	n := len(data)
	i := 0
	var err error
	for i < n {
		err = binary.Read(r, order, data[i])
		if err != nil {
			return err
		}
		i++
	}
	return nil
}

// Write binary多写，传入的data必须是定长类型、定长类型的slice、或者上面两种情况的指针
func Write(w io.Writer, order binary.ByteOrder, data ...interface{}) error {
	n := len(data)
	i := 0
	var err error
	for i < n {
		//fmt.Printf("data[%d]: %v\n", i, data[i])
		err = binary.Write(w, order, data[i])
		if err != nil {
			return err
		}
		i++
	}
	return nil
}
