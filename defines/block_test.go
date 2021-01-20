/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/15/20 2:08 PM
* @Description: The file is for
***********************************************************************/

package defines

import (
	"reflect"
	"testing"
)

func TestBlock(t *testing.T) {
	tx, err := NewTransactionAndSign("from", "to", 10, nil, "this is a tx")
	if err != nil {
		t.Error(err)
	}
	b, err := NewBlockAndSign(1, "id", []byte("prevhash"), []*Transaction{tx}, "this is a block")
	if err != nil {
		t.Error(err)
	}

	bBytes, err := b.Encode()
	if err != nil {
		t.Error(err)
	}

	nb := new(Block)
	err = nb.Decode(bBytes)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(b, nb) {
		t.Error("error")
	}
}
