package main

import (
	"strconv"
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/driver"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/globalstate"
)

func syncBtnLEDs(globalstate globalstate.FSM) {
	online := false
	for {
		select {
		case b := <-haveConsensusBtnSyncCh:
			online = b
		case <-time.After(1 * time.Microsecond):

		}
		if !online {
			continue
		}

		time.Sleep(500 * time.Millisecond)
		s, err := globalstate.GetState()
		if err != nil {
			continue
		}
		for k, v := range s.HallUpButtons {
			f, _ := strconv.Atoi(k)
			if v.LastStatus == "done" {
				driver.BtnLEDClear(driver.Btn{Floor: f, Type: driver.HallUp})
			} else {
				driver.BtnLEDSet(driver.Btn{Floor: f, Type: driver.HallUp})
			}
		}

		for k, v := range s.HallDownButtons {
			f, _ := strconv.Atoi(k)
			if v.LastStatus == "done" {
				driver.BtnLEDClear(driver.Btn{Floor: f, Type: driver.HallDown})
			} else {
				driver.BtnLEDSet(driver.Btn{Floor: f, Type: driver.HallDown})
			}
		}

	}
}
