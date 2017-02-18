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
	switch dir {
	case "up":
		goToCh <- driver.Btn{Floor: f, Type: driver.HallUp}
	case "down":
		goToCh <- driver.Btn{Floor: f, Type: driver.HallDown}
	}
}

func onPromotion() {}

func onDemotion() {}

func onAquiredConsensus() {
	mainlogger.Println("Aquired consensus")
	haveConsensusBtnSyncCh <- true
	haveConsensusAssignerCh <- true
}
func onLostConsensus() {
	mainlogger.Println("Lost consensus")
	haveConsensusBtnSyncCh <- false
	haveConsensusAssignerCh <- false
}

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
		mainlogger.Printf("[INFO] Unable to send button update to network. Storing locally.\n")
		if err := ls.UpdateButtonStatus(bsu); err != nil {
			mainlogger.Printf("[ERROR] Unable to handle button press: %v", err.Error())
			return
		}
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
		// fmt.Println("Failed to send liftupdate...")
	}

}

func onDstReached(b driver.Btn, pickup bool) {
	bsu := globalstate.ButtonStatusUpdate{
		Floor:  uint(b.Floor),
		Dir:    b.Type.String(),
		Status: globalstate.BtnStateDone,
	}
	if err := gs.UpdateButtonStatus(bsu); err != nil {
		if err := ls.UpdateButtonStatus(bsu); err != nil {
			mainlogger.Printf("[ERROR] Unable to handle button press: %v", err.Error())
			return
		}
	}
	driver.BtnLEDClear(b)
	if !pickup {
		fmt.Println("done with order")
		orderDoneCh <- struct{}{}
	}
}

// Peer discovery callbacks
// =============================================================================
func onNewPeer(p peerdiscovery.Peer) {

}

func onLostPeer(p peerdiscovery.Peer) {
	fmt.Printf("Lost peer: %+v\n", p)
}
