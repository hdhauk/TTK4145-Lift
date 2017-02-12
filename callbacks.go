package main

import (
	"fmt"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/peerdiscovery"
)

// Globalstate callbacks
// =============================================================================
func onIncommingCommand(f int, dir string) {
	fmt.Printf("Supposed to go to floor %d, somebody want %s from there\n", f, dir)
}

func onPromotion() {}

func onDemotion() {}

// Peer discovery callbacks
// =============================================================================
func onNewPeer(p peerdiscovery.Peer) {

}

func onLostPeer(p peerdiscovery.Peer) {
	fmt.Printf("Lost peer: %+v\n", p)
}
