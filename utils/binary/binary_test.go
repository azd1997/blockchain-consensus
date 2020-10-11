/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/10/20 8:14 PM
* @Description: The file is for
***********************************************************************/

package binary

import (
	"bytes"
	"encoding/binary"
	"testing"
)

///////////////// 测试数据 /////////////////////

// 定长类型数据
var data1 uint32 = 34

// 定长类型的slice
var data2 = []byte("slice")

// 定长类型指针
var data3 = &data1

// slice指针
var data4 = &data2

///////////////// 测试逻辑 /////////////////////

func TestReadAndWrite(t *testing.T) {

	buf := bytes.NewBuffer(make([]byte, 0))
	err := Write(buf, binary.BigEndian, data1, data2, data3, data4)
	if err != nil {
		t.Error(err)
	}
	t.Log(buf.Bytes())

	d1 := uint32(0)
	d2 := make([]byte, 5)
	d3 := new(uint32)
	x := make([]byte, 5)
	d4 := &x
	err = Read(buf, binary.BigEndian, &d1, d2, d3, d4)
	if err != nil {
		t.Error(err)
	}
	if d1 != data1 {
		t.Errorf("d1(%d) != data1(%d)\n", d1, data1)
	}
	if !bytes.Equal(d2, data2) {
		t.Errorf("d2(%s) != data2(%s)\n", string(d2), string(data2))
	}
	if *d3 != *data3 {
		t.Errorf("d3(%d) != data3(%d)\n", *d3, *data3)
	}
	if !bytes.Equal(*d4, *data4) {
		t.Errorf("d4(%s) != data4(%s)\n", string(*d4), string(*data4))
	}
}