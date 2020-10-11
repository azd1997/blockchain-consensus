/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/9/20 8:09 PM
* @Description: The file is for
***********************************************************************/

package defines

import (
	"bytes"
	"reflect"
	"testing"
)

var testRequest1 = &Request{
	Type:       RequestType_Blocks,
	IndexStart: 3,
	IndexCount: 15,
	Hashes:     [][]byte{
		[]byte("Hash1"),
		[]byte("Hash2"),
	},
}

var testRequest2 = &Request{
	Type:       RequestType_Blocks,
	IndexStart: 3,
	IndexCount: 15,
}

func TestRequest(t *testing.T) {
	var tests = []struct{
		name string
		req *Request
	}{
		{"hashnum!=0", testRequest1},

		{"hashnum==0", testRequest2},
	}

	// 测试逻辑
	for _, test := range tests {
		test := test

		// 调用Len()
		length := test.req.Len()
		// 调用Encode()
		b, err := test.req.Encode()
		if err != nil {
			t.Errorf("[%s] error: %s\n", test.name, err)
		}
		if len(b) != length {
			t.Errorf("[%s] error: length(%d) != lenb(%d)\n", test.name, length, len(b))
		}
		// 调用Decode()
		areq := new(Request)
		err = areq.Decode(bytes.NewReader(b))
		if err != nil {
			t.Errorf("[%s] error: %s\n", test.name, err)
		}
		if !reflect.DeepEqual(test.req, areq) {
			t.Errorf("[%s] error: req(%v) != areq(%v)\n", test.name, test.req, areq)
		}
	}
}

