package main

import "github.com/azd1997/blockchain-consensus/measure/monitor"

func main() {
	monitor.Run("127.0.0.1:9998", "127.0.0.1:9999")
}
