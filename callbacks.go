package main

import (
	"fmt"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/driver"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/globalstate"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/peerdiscovery"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/statetools"
)

// Globalstate callbacks
// =============================================================================
func onIncomingCommand(f int, dir string) {
	// TODO: Doublecheck if the lift isn't currently busy
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
	err := gs.UpdateButtonStatus(bsu)
	if err != nil {
		fmt.Println("Failed to send button update...")
		// TODO: Add to local store
	}
	driver.BtnLEDSet(b)
}

func onNewStatus(f int, dir string, dstFloor int, dstDir string) {
	// Check if there are anyone to pick up.
	state, _ := gs.GetState()
	if statetools.ShouldStopAndPickup(state, f, dir) {
		driver.StopForPickup(f, dir)
		fmt.Printf("Pickup was available in Floor=%d Dir=%s. Stopping!\n", f, dir)
	}

	// Send status update
	lsu := globalstate.LiftStatusUpdate{
		CurrentFloor: uint(f),
		CurrentDir:   dir,
		DstFloor:     uint(dstFloor),
		DstBtnDir:    dstDir,
	}
	if err := gs.UpdateLiftStatus(lsu); err != nil {
		fmt.Println("Failed to send liftupdate...")
	}
	// Check lift reached a foor and if it can pick someone up there
	// TODO
	// --> GetState
	// --> statetools.ShouldStopAndPickup(state, f, dir)

}

func onDstReached(b driver.Btn) {
	bsu := globalstate.ButtonStatusUpdate{
		Floor:  uint(b.Floor),
		Dir:    b.Type.String(),
		Status: globalstate.BtnStateDone,
	}
	if err := gs.UpdateButtonStatus(bsu); err != nil {
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
