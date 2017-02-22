package main

import (
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/driver"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/globalstate"
)

func noConsensusAssigner() {
	consensus := true
	for {
		select {
		// Listen for updates on the consensus.
		case b := <-haveConsensusAssignerCh:
			consensus = b
		case <-time.After(1 * time.Second):
		}

		// Proceed with actual work only of there are no consensus.
		if consensus {
			continue
		}

		// Assign orders from local state.
		floor, dir := stateLocal.GetNextOrder()
		bsu := globalstate.ButtonStatusUpdate{
			Floor:  uint(floor),
			Dir:    dir,
			Status: globalstate.BtnStateAssigned,
		}
		if dir == "up" {
			goToCh <- driver.Btn{Floor: floor, Type: driver.HallUp}
			stateLocal.UpdateButtonStatus(bsu)
		} else if dir == "down" {
			goToCh <- driver.Btn{Floor: floor, Type: driver.HallDown}
			stateLocal.UpdateButtonStatus(bsu)
		}

	}
}
