package driver

import "fmt"

var btnPressCh = make(chan Btn, 4)
var floorDetectCh = make(chan int)
var dirChangeCh = make(chan string)

func eventHandler() {
	// Initalize datastores
	hallUpBtns := make(map[int]bool)
	hallDownBtns := make(map[int]bool)
	CabBtns := make(map[int]bool)

	lastFloor := 0
	lastDir := stop
	<-initDone
	fmt.Printf("eventhandler.go: Entering for-loop\n")
	for {
		select {

		// Detected button presses
		case btn := <-btnPressCh:
			switch btn.Type {
			case HallUp:
				if hallUpBtns[btn.Floor] == false {
					driver.setBtnLED(btn, true)
					cfg.OnBtnPress(btn.Type.String(), btn.Floor)
					hallUpBtns[btn.Floor] = true
				}
			case HallDown:
				if hallDownBtns[btn.Floor] == false {
					driver.setBtnLED(btn, true)
					cfg.OnBtnPress(btn.Type.String(), btn.Floor)
					hallDownBtns[btn.Floor] = true
				}
			case Cab:
				if CabBtns[btn.Floor] == false {
					driver.setBtnLED(btn, true)
					cfg.OnBtnPress(btn.Type.String(), btn.Floor)
					CabBtns[btn.Floor] = true
				}
			}

		// Floor sensor triggered
		case floor := <-floorDetectCh:
			if floor != lastFloor {
				fmt.Printf("eventhandler.go: Found new floor\n")
				lastFloor = floor
				cfg.OnFloorDetect(floor)
				// Pass newly detected floor on to the Autopilot
				apFloor <- floor
			}

		// Direction changed
		case dir := <-dirChangeCh:
			if lastDir != dir {
				lastDir = dir
				cfg.OnNewDirection(dir)
			}
		}
	}
}
