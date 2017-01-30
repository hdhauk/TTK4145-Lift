package main

import (
	"fmt"
	"os"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/peerdiscovery"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("example")

func initLogger() {
	stdBackend := logging.NewLogBackend(os.Stderr, "", 0)
	stdFormat := logging.MustStringFormatter(
		`%{color}%{time:15:04:05.00} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)
	stdFormatter := logging.NewBackendFormatter(stdBackend, stdFormat)
	logging.SetBackend(stdFormatter)

}

func logPeerUpdate(p peerdiscovery.PeerUpdate) {
	logger.Notice("Peer update:\n")
	//fmt.Printf("\tPeers:\t%q\n", p.Peers)
	printPeerArr(p.Peers)
	fmt.Printf("\tNew:\t%q\n", p.New)
	fmt.Printf("\tLost:\t%q\n", p.Lost)
}

func printPeerArr(p []string) {
	if len(p) == 1 {
		fmt.Printf("\tPeers:\t%q\n", p)
		return
	}
	fmt.Printf("\tPeers:")
	ret := "\t["
	for i, elem := range p {
		if i == 0 {
			ret = ret + fmt.Sprintf("\"%s\"\n", elem)
			continue
		} else if i == len(p)-1 {
			ret = ret + fmt.Sprintf("\t\t \"%s\"]\n", elem)
			continue
		}
		ret = ret + fmt.Sprintf("\t\t \"%s\"\n", elem)
	}
	fmt.Printf(ret)
}
