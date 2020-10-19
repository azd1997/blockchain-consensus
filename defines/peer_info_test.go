/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/13/20 1:02 PM
* @Description: The file is for
***********************************************************************/

package defines

import (
	"bytes"
	"reflect"
	"testing"
)

func TestPeerInfo(t *testing.T) {
	pi := &PeerInfo{
		Id:   "id",
		Addr: "addr",
	}
	b, err := pi.Encode()
	if err != nil {
		t.Error(err)
	}

	api := new(PeerInfo)
	err = api.Decode(bytes.NewReader(b))
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(pi, api) {
		t.Error("error")
	}
}
