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
	Epoch:   8,
	From:    "id_from",
	To:      "id_toto", // From和To长度要一致
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

		err := test.msg.WriteDesc("type", "testmsg")
		if err != nil {
			t.Error(err)
		}

		// 调用Sign()
		err = test.msg.Sign()
		if err != nil {
			t.Error(err)
		}
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
		t.Logf("encoded b: %s\n", string(b))
		t.Logf("encoded b: %v\n", b)

		// 调用Decode()
		amsg := new(Message)
		r := bytes.NewReader(b)
		t.Logf("before decode, r.Len = %d\n", r.Len())
		err = amsg.Decode(r)
		if err != nil {
			t.Errorf("[%s] error: %s\n", test.name, err)
		}

		t.Logf("after decode, r.Len = %d\n", r.Len()) // should be 0
		rest := make([]byte, r.Len())
		r.Read(rest)
		t.Logf("rest of r: %s\n", string(rest)) //      description 	signature
		// 这里的输出结果说明了，在编码过程中不知道为什么，多了许多byte(0)
		t.Logf("rest of r: %v\n", rest)
		// [0 0 0 0 0 11 100 101 115 99 114 105 112 116 105 111 110 0 9 115 105 103 110 97 116 117 114 101]
		// 加上前面加载的desclen和siglen，合计编码过程中多了2+2+4=8个0

		// 调用Verify()
		err = amsg.Verify()
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(test.msg, amsg) {
			t.Errorf("[%s] error: msg != amsg\nmsg:%v\namsg:%v\n", test.name, test.msg, amsg)
		}
	}
}
