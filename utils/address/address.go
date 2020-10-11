/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/10/20 4:19 PM
* @Description: 网络地址相关的操作
***********************************************************************/

package address

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ParseTCP4 解析TCP4地址
func ParseTCP4(addr string) (*net.TCPAddr, error) {
	slice := strings.Split(addr, ":")
	if len(slice) != 2 {
		return nil, fmt.Errorf("address(%s) parse error: len(slice) != 2", addr)
	}

	// 解析端口
	port, err := strconv.Atoi(slice[1])
	if err != nil {
		return nil, fmt.Errorf("address(%s) parse error: %s", addr, err)
	}

	// 解析IP
	ip := net.ParseIP(slice[0])
	if ip == nil {
		return nil, fmt.Errorf("address(%s) parse error: ip == nil", addr)
	}

	return &net.TCPAddr{
		IP:   ip,
		Port: port,
	}, nil
}
