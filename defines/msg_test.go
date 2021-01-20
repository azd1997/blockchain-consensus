/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/9/20 11:05 PM
* @Description: The file is for
***********************************************************************/

package defines

import (
	"reflect"
	"testing"
)

var testMessage = &Message{
	Version: CodeVersion,
	Type:    MessageType_Blocks,
	Epoch:   8,
	From:    "id_from",
	To:      "id_toto", // From和To长度要一致
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
		// 调用Encode()
		b, err := test.msg.Encode()
		if err != nil {
			t.Errorf("[%s] error: %s\n", test.name, err)
		}

		// 调用Decode()
		amsg := new(Message)
		err = amsg.Decode(b)
		if err != nil {
			t.Errorf("[%s] error: %s\n", test.name, err)
		}

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
