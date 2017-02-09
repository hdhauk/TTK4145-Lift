package main

import (
	"fmt"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/driver"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/globalstate"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/peerdiscovery"
)

// Globalstate callbacks
// =============================================================================
func onIncommingCommand(f int, dir string) {
	fmt.Printf("Supposed to go to floor %d, somebody want %s from there\n", f, dir)
	driver.GoToFloor(f)
}

func onPromotion() {}

func onDemotion() {}

// Driver callbacks
// =============================================================================
func onBtnPress(b driver.Btn) {
	bsu := globalstate.ButtonStatusUpdate{
		Floor:  uint(b.Floor),
		Dir:    b.Type.String(),
		Status: globalstate.BtnStateUnassigned,
	}
	globalstate.UpdateButtonStatus(bsu)
	driver.BtnLEDSet(b)
}

func onDstReached(f int) {

}

func onFloorDetect(f int) {

}

func onNewDirection(dir string) {

}

// Peer discovery callbacks
// =============================================================================
func onNewPeer(p peerdiscovery.Peer) {

}

func onLostPeer(p peerdiscovery.Peer) {
	fmt.Printf("Lost peer: %+v\n", p)
}
