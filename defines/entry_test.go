/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/9/20 9:44 PM
* @Description: The file is for
***********************************************************************/

package defines

import (
	"bytes"
	"reflect"
	"testing"
)

var testEntry = &Entry{
	BaseIndex: 3,
	Base:      []byte("BaseHash"),
	Type:      EntryType_Block,
	Data:      []byte("block data"),
}

func TestEntry(t *testing.T) {
	var tests = []struct{
		name string
		ent *Entry
	}{
		{"normal_case", testEntry},
	}

	// 测试逻辑
	for _, test := range tests {
		test := test

		// 调用Len()
		length := test.ent.Len()
		// 调用Encode()
		b, err := test.ent.Encode()
		if err != nil {
			t.Errorf("[%s] error: %s\n", test.name, err)
		}
		if len(b) != length {
			t.Errorf("[%s] error: length(%d) != lenb(%d)\n", test.name, length, len(b))
		}
		// 调用Decode()
		aent := new(Entry)
		err = aent.Decode(bytes.NewReader(b))
		if err != nil {
			t.Errorf("[%s] error: %s\n", test.name, err)
		}
		if !reflect.DeepEqual(test.ent, aent) {
			t.Errorf("[%s] error: ent(%v) != aent(%v)\n", test.name, test.ent, aent)
		}
	}
}
