package main

import (
	"fmt"
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/driver"
	"bitbucket.org/halvor_haukvik/ttk4145-elevator/globalstate"
)

func noConsensusAssigner() {
	online := true
	for {

		select {
		case b := <-haveConsensusAssignerCh:
			online = b
		case <-time.After(1 * time.Second):
		}
		if online {
			continue
		}

		floor, dir := ls.GetNextOrder()
		bsu := globalstate.ButtonStatusUpdate{
			Floor:  uint(floor),
			Dir:    dir,
			Status: globalstate.BtnStateAssigned,
		}
		if dir == "up" {
			fmt.Println("up")
			goToCh <- driver.Btn{Floor: floor, Type: driver.HallUp}
			ls.UpdateButtonStatus(bsu)
		} else if dir == "down" {
			fmt.Println("down")
			goToCh <- driver.Btn{Floor: floor, Type: driver.HallDown}
			ls.UpdateButtonStatus(bsu)
		}

	}
}
