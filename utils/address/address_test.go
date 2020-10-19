/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/10/20 7:39 PM
* @Description: The file is for
***********************************************************************/

package address

import (
	"testing"
)

func TestParseTCP4(t *testing.T) {
	addrstr := "127.0.0.1:8080"
	addr, err := ParseTCP4(addrstr)
	if err != nil {
		t.Error(err)
	}
	t.Log(addr)

	str2 := addr.String()

	if addrstr != str2 {
		t.Errorf("addrstr(%s) != str2(%s)\n", addrstr, str2)
	}
}
