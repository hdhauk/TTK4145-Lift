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
	driver.GoToFloor(f, dir)
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
	err := globalstate.UpdateButtonStatus(bsu)
	if err != nil {
		fmt.Println("Failed to send button update...")
	}
	driver.BtnLEDSet(b)
}

func onNewStatus(f, dst int, dir string) {
	lsu := globalstate.LiftStatusUpdate{
		Floor: uint(f),
		Dst:   uint(dst),
		Dir:   dir,
	}
	if err := globalstate.UpdateLiftStatus(lsu); err != nil {
		fmt.Println("Failed to send liftupdate...")
	}
}

func onDstReached(b driver.Btn) {
	bsu := globalstate.ButtonStatusUpdate{
		Floor:  uint(b.Floor),
		Dir:    b.Type.String(),
		Status: globalstate.BtnStateDone,
	}
	if err := globalstate.UpdateButtonStatus(bsu); err != nil {
		fmt.Println("Failed to send DONE status")
	}
	driver.BtnLEDClear(b)
}

// Peer discovery callbacks
// =============================================================================
func onNewPeer(p peerdiscovery.Peer) {

}

func onLostPeer(p peerdiscovery.Peer) {
	fmt.Printf("Lost peer: %+v\n", p)
}