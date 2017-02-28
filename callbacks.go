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
	haveConsensusBtnSyncCh <- true
	haveConsensusAssignerCh <- true

	// Send share all unhandled orders
	failed := 0
	buttonUpdates := stateLocal.GetShareworthyUpdates()
	for _, bsu := range buttonUpdates {
		err := stateGlobal.UpdateButtonStatus(bsu)
		if err != nil {
			mainlogger.Printf("[INFO] Unable to send button update to network. Storing locally.\n")
			if err := stateLocal.UpdateButtonStatus(bsu); err != nil {
				mainlogger.Printf("[ERROR] Unable to handle button press: %v", err.Error())
				failed++
				continue
			}

		}
	}
	mainlogger.Printf("[INFO] Aquired consensus.\n")
	mainlogger.Printf("[INFO] Shared relevant button statuses with peers, %d in total. %d failed.\n", len(buttonUpdates), failed)
}

func onLostConsensus() {
	mainlogger.Println("[WARN] Lost consensus. Falling back to local non-consensus mode.")
	haveConsensusBtnSyncCh <- false
	haveConsensusAssignerCh <- false
}

// Driver callbacks
// =============================================================================
func onBtnPress(b driver.Btn) {
	if b.Type != driver.Cab {
		bsu := globalstate.ButtonStatusUpdate{
			Floor:  uint(b.Floor),
			Dir:    b.Type.String(),
			Status: globalstate.BtnStateUnassigned,
		}
		err := stateGlobal.UpdateButtonStatus(bsu)
		if err != nil {
			mainlogger.Printf("[INFO] Unable to send button update to network. Storing locally.\n")
			if err := stateLocal.UpdateButtonStatus(bsu); err != nil {
				mainlogger.Printf("[ERROR] Unable to handle button press: %v", err.Error())
				return
			}
		}
	}

	driver.BtnLEDSet(b)
}

func onNewStatus(f int, dir string, dstFloor int, dstDir string) {
	// Check if there are anyone to pick up.
	state, _ := stateGlobal.GetState()
	if statetools.ShouldStopAndPickup(state, f, dir) {
		driver.StopForPickup(f, dir)
		mainlogger.Printf("[INFO] Pickup was available in Floor=%d Dir=%s. Stopping!\n", f, dir)
	}

	// Send status update
	lsu := globalstate.LiftStatusUpdate{
		CurrentFloor: uint(f),
		CurrentDir:   dir,
		DstFloor:     uint(dstFloor),
		DstBtnDir:    dstDir,
	}
	if err := stateGlobal.UpdateLiftStatus(lsu); err != nil {
		mainlogger.Println("[WARN] Failed to send liftupdate.")
	}

}

func onDstReached(b driver.Btn, pickup bool) {
	bsu := globalstate.ButtonStatusUpdate{
		Floor:  uint(b.Floor),
		Dir:    b.Type.String(),
		Status: globalstate.BtnStateDone,
	}
	if err := stateGlobal.UpdateButtonStatus(bsu); err != nil {
		if err := stateLocal.UpdateButtonStatus(bsu); err != nil {
			mainlogger.Printf("[ERROR] Unable to handle button press: %v", err.Error())
			return
		}
	}
	driver.BtnLEDClear(b)
	if !pickup {
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
