/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/9/20 11:05 PM
* @Description: The file is for
***********************************************************************/

package defines

import (
	"bytes"
	"reflect"
	"testing"
)

var testMessage = &Message{
	Version: CodeVersion,
	Type:    MessageType_Data,
	From:    "id_from",
	To:      "id_toto", // From和To长度要一致
	Sig:     []byte("Signature"),
	Entries: []*Entry{testEntry},
	Reqs:    []*Request{testRequest1, testRequest2},
	Desc:    "description",
}

func TestMessage(t *testing.T) {
	var tests = []struct {
		name string
		msg  *Message
	}{
		{"normal_case", testMessage},
	}

	// 测试逻辑
	for _, test := range tests {
		test := test

		// 调用Len()
		length := test.msg.Len()
		// 调用Encode()
		b, err := test.msg.Encode()
		if err != nil {
			t.Errorf("[%s] error: %s\n", test.name, err)
		}
		if len(b) != length {
			t.Errorf("[%s] error: length(%d) != lenb(%d)\n", test.name, length, len(b))
		}
		// 调用Decode()
		amsg := new(Message)
		r := bytes.NewReader(b)
		err = amsg.Decode(r)
		if err != nil {
			t.Errorf("[%s] error: %s\n", test.name, err)
		}
		if !reflect.DeepEqual(test.msg, amsg) {
			t.Errorf("[%s] error: msg(%v) != amsg(%v)\n", test.name, test.msg, amsg)
		}
	}
}
