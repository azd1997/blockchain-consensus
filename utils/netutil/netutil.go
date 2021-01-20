/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 1/19/21 4:21 PM
* @Description: The file is for
***********************************************************************/

package netutil

import (
	"net"
	"strconv"
	"strings"
)

// ParseIPPort 解析IP:Port
func ParseIPPort(ipPort string) (net.IP, int) {
	s := strings.Split(ipPort, ":")
	if len(s) != 2 {
		return nil, 0
	}

	ip := net.ParseIP(s[0])
	if ip == nil || ip.To4() == nil {
		return nil, 0
	}

	port, err := strconv.Atoi(s[1])
	if err != nil || port <= 0 || port > 65535 {
		return nil, 0
	}

	return ip, port
}
