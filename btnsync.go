package main

import (
	"strconv"
	"time"

	"github.com/hdhauk/TTK4145-Lift/driver"
)

func syncBtnLEDs() {
	consensus := false

	for {
		// For incoming changes in the consensus status.
		select {
		case b := <-haveConsensusBtnSyncCh:
			consensus = b
		case <-time.After(1 * time.Microsecond):

		}
		if !consensus {
			continue
		}
		// Only do sync every half second.
		time.Sleep(500 * time.Millisecond)

		state, err := stateGlobal.GetState()
		if err != nil {
			continue
		}

		// Check all hall buttons.
		for floorStr, status := range state.HallUpButtons {
			f, _ := strconv.Atoi(floorStr)
			if status.LastStatus == "done" {
				driver.BtnLEDClear(driver.Btn{Floor: f, Type: driver.HallUp})
			} else {
				driver.BtnLEDSet(driver.Btn{Floor: f, Type: driver.HallUp})
			}

		}
		for floorStr, status := range state.HallDownButtons {
			f, _ := strconv.Atoi(floorStr)
			if status.LastStatus == "done" {
				driver.BtnLEDClear(driver.Btn{Floor: f, Type: driver.HallDown})
			} else {
				driver.BtnLEDSet(driver.Btn{Floor: f, Type: driver.HallDown})
			}
		}

	}
}
