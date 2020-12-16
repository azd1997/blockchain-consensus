/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/9/20 9:44 PM
* @Description: The file is for
***********************************************************************/

package defines

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

var testEntry = &Entry{
	BaseIndex: 3,
	Base:      []byte("BaseHash"),
	Type:      EntryType_Block,
	Data:      []byte("block data"),
}

var testEntry2 = func() *Entry {
	testPeerInfo := &PeerInfo{
		Id:   "id",
		Addr: "127.0.0.1:9090",
		Duty: PeerDuty_Seed,
		Attr: 0,
		Data: nil,
	}
	b, err := testPeerInfo.Encode()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	e := &Entry{
		BaseIndex: 0,
		Base:      []byte("base"),
		Type:      EntryType_Neighbor,
		Data:      b,
	}
	fmt.Println("entry2: ", e.BaseIndex, e.Base, e.Type, e.Data)
	fmt.Println("origin b length: ", len(b))
	return e
}

func TestEntry(t *testing.T) {
	te2 := testEntry2()

	var tests = []struct {
		name string
		ent  *Entry
	}{
		//{"normal_case", testEntry},
		{"case2", te2},
	}

	// 测试逻辑
	for _, test := range tests {
		test := test

		// 调用Len()
		length := test.ent.Len()
		fmt.Println("length: ", length)
		// 调用Encode()
		b, err := test.ent.Encode()
		if err != nil {
			t.Errorf("[%s] error: %s\n", test.name, err)
		}
		if len(b) != length {
			t.Errorf("[%s] error: length(%d) != lenb(%d)\n", test.name, length, len(b))
		}

		fmt.Println("b: ", b)

		// 调用Decode()
		aent := new(Entry)
		err = aent.Decode(bytes.NewReader(b))
		if err != nil {
			t.Errorf("[%s] error: %s\n", test.name, err)
		}
		fmt.Println(aent.BaseIndex, aent.Base, aent.Type, aent.Data)

		if !reflect.DeepEqual(test.ent, aent) {
			t.Errorf("[%s] error: ent(%v) != aent(%v)\n", test.name, test.ent, aent)
		}

		if aent.Type == EntryType_Neighbor {
			nei := new(PeerInfo)
			err := nei.Decode(aent.Data)
			if err != nil {
				t.Error(err)
			}
			fmt.Println(nei)
		}

	}
}

func TestEntry_Decode(t *testing.T) {
	testPeerInfo := &PeerInfo{
		Id:   "id",
		Addr: "127.0.0.1:9090",
		Duty: PeerDuty_Seed,
		Attr: 0,
		Data: nil,
	}
	b, err := testPeerInfo.Encode()
	if err != nil {
		t.Error(err)
	}
	e := &Entry{
		BaseIndex: 0,
		Base:      []byte("base"),
		Type:      EntryType_Neighbor,
		Data:      b,
	}
	fmt.Println("entry2: ", e.BaseIndex, e.Base, e.Type, e.Data)
	fmt.Println("origin b length: ", len(b), "origin b: ", b)

	entryBytes, err := e.Encode()
	if err != nil {
		t.Error(err)
	}

	ae := new(Entry)
	err = ae.Decode(bytes.NewReader(entryBytes))
	if err != nil {
		t.Error(err)
	}
	fmt.Println("ae: ", ae.BaseIndex, ae.Base, ae.Type, ae.Data)
}
