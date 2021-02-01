package main

import (
	"flag"
	"github.com/azd1997/blockchain-consensus/lab/pot"
)

var (
	flagNseed = flag.Int("nseed", 1, "number of seeds")
	flagNpeer = flag.Int("npeer", 3, "number of peers")
	flagMyno = flag.Int("myno", 0, "my id number")
	flagMyduty = flag.String("myduty", "peer", "my duty")
	flagShutdownAt = flag.Int("shutdown", 21, "shutdown at ti")
	flagDebug = flag.Bool("debug", false, "enable debug?")
	flagAddCaller = flag.Bool("addcaller", false, "enable addCaller in log")
	flagEnableClients = flag.Bool("clients", false, "enable clients")
)

// ./pot -nseed=1 -npeer=3 -myno=1 -myduty=seed -shutdown=21 -debug -addcaller -enableClients
func main() {
	flag.Parse()

	_, _, seedsm, peersm := pot.GenIdsAndAddrs(*flagNseed, *flagNpeer)
	idnum := pot.Idnum(*flagMyno)
	id, addr :=

	_, err := pot.StartNode(id, addr, shutdownAtTi, cheatAtTi,
		seedsm, peersm, debug, addCaller)
	if err != nil {
		panic(err)
	}
}


