package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/driver"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/globalstate"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/peerdiscovery"
)

// Command line defaults
const (
	defaultRaftPort = ":12000"
)

// Command line parameters
var nick string
var simPort string
var raftPort string
var joinAddr string

func main() {
	initLogger()
	// Parse arg flags
	flag.StringVar(&nick, "nick", "", "Nickname of this peer")
	flag.StringVar(&simPort, "sim", "", "Listening port of the simulator")
	flag.StringVar(&raftPort, "raftPort", defaultRaftPort, "Set Raft port")
	flag.Parse()
	if simPort != "" {
		logger.Notice("Starting in simulator mode.")
	}
	var ownID = struct {
		ID   string
		Nick string
	}{
		ID: makeUUID(), Nick: peerName(nick),
	}

	// Initialize peer discovery
	peers := make(map[string]string)
	peerCh := make(chan string)
	go peerdiscovery.Start(33324, ownID.Nick, func(id, ip string) {
		fmt.Printf("ID = %v\n", id)
		peerCh <- ip
	})
	time.Sleep(1 * time.Second)
	go func() {
		for {
			select {
			case ip := <-peerCh:
				peers[ip] = ip
			}
		}
	}()

	// Initialize driver
	cfg := driver.Config{
		SimMode: true,
		SimPort: "53566",
		Floors:  4,
		OnBtnPress: func(b driver.Btn) {
			driver.BtnLEDSet(b)
			time.Sleep(5 * time.Second)
			driver.BtnLEDClear(b)
		},
	}
	go driver.Init(cfg)

	// Initalize globalstate
	s := globalstate.New()
	tmpDir, _ := ioutil.TempDir("", "store_test")
	defer os.RemoveAll(tmpDir)
	s.RaftBind = "127.0.0.1:0"
	s.RaftDir = tmpDir
	if err := s.Open(len(peers) == 0); err != nil {
		log.Fatalf("failed to open store: %s", err.Error())
	}
	if len(peers) != 0 {
		var p string
		for _, v := range peers {
			p = v
			break
		}
		if err := join(p, raftPort); err != nil {
			log.Fatalf("failed to join noe at %s: %s", p, err.Error())
		}

	}
	select {}

}

func join(joinAddr, raftPort string) error {
	b, err := json.Marshal(map[string]string{"addr": raftPort})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func peerName(id string) string {
	if id == "" {
		id = strconv.Itoa(os.Getpid())
	}
	localIP, err := peerdiscovery.GetLocalIP()
	if err != nil {
		logger.Warning("Not connected to the internet.")
		localIP = "DISCONNECTED"
	}
	return fmt.Sprintf("%s@%s", id, localIP)
}
