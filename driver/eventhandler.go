package driver

var btnPressCh = make(chan Btn, 4)
var floorDetectCh = make(chan int)
var dirChangeCh = make(chan string)

func eventHandler() {
	// Initalize datastores
	hallUpBtns := make(map[int]bool)
	hallDownBtns := make(map[int]bool)
	CabBtns := make(map[int]bool)

	//lastDir := stop
	beenDriving := true
	setBeenDriving := func(b bool) {
		beenDriving = b
	}
	<-initDone
	for {
	selector:
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
			/*
				- Case 1: Incoming positive floor detection
						Case 1a: The carrige have been driving since the last positive detection
							--> Handle the detection as real thing
						Case 1b: The carrige have NOT been driving since the last positive detection
							--> Do nothing. The detection is already been handled
				- Case 2: Incoming negative floor detection
							--> Set beenDriving = true
			*/
			// Case 1
			if floor != -1 {
				// Case 1a
				if beenDriving {
					driver.setFloorLED(floor)
					setBeenDriving(false)
					cfg.OnFloorDetect(floor)
					apFloor <- floor
					break selector
				}
				// Case 1b
				break selector
			}
			// Case 2
			setBeenDriving(true)
		}
	}
}
