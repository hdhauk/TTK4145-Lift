package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/driver"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/peerdiscovery"
)

const (
	on  = true
	off = false
)

func main() {
	initLogger()
	// Parse arg flags
	var nick string
	var simPort string
	flag.StringVar(&nick, "nick", "", "nick name of this peer")
	flag.StringVar(&simPort, "sim", "", "listening port of the simulator")
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

	// Start peer discovery
	go peerdiscovery.Start(33324, ownID.Nick, func(id, IP string) {
		// Callback on new peer
		fmt.Printf("ID = %v\n", id)
		fmt.Printf("IP = %v\n", IP)
	})

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
	fmt.Println("initializing")
	go driver.Init(cfg)
	time.Sleep(4 * time.Second)

	var i int
	for {
		fmt.Printf("Go to floor: \n")
		fmt.Scanf("%d", &i)
		fmt.Printf("Goint to: %v\n", i)
		driver.GoToFloor(i)
	}

	// for {
	// 	select {
	// 	case <-time.After(1 * time.Second):
	// 		//logPeerUpdate(p)
	//
	// 	}
	// }
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
