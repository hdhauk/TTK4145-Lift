package driver

var btnPressCh = make(chan Btn, 4)
var floorDetectCh = make(chan int)
var dirChangeCh = make(chan string)

func eventHandler() {
	// Initalize datastores
	var hallUpBtns, hallDownBtns, CabBtns map[int]bool
	lastFloor := 0
	lastDir := stop

	for {
		select {

		// Detected button presses
		case btn := <-btnPressCh:
			switch btn.Type {
			case HallUp:
				if hallUpBtns[btn.Floor] == false {
					driver.setBtnLED(btn, true)
					cfg.onBtnPress(btn.Type.String(), btn.Floor)
				}
			case HallDown:
				if hallDownBtns[btn.Floor] == false {
					driver.setBtnLED(btn, true)
					cfg.onBtnPress(btn.Type.String(), btn.Floor)
				}
			case Cab:
				if CabBtns[btn.Floor] == false {
					driver.setBtnLED(btn, true)
					cfg.onBtnPress(btn.Type.String(), btn.Floor)
				}
			}

		// Floor sensor triggered
		case floor := <-floorDetectCh:
			if floor != lastFloor {
				lastFloor = floor
				cfg.onFloorDetect(floor)
				// Pass newly detected floor on to the Autopilot
				apFloor <- floor
			}

		// Direction changed
		case dir := <-dirChangeCh:
			if lastDir != dir {
				lastDir = dir
				cfg.onNewDirection(dir)
			}
		}
	}
}
