package driver

import (
	"time"
)

func btnPressHandler(btnPressCh <-chan Btn) {
	// Initialize button registers
	hallUpBtns := make(map[int]time.Time)
	hallDownBtns := make(map[int]time.Time)
	CabBtns := make(map[int]time.Time)
	cbTriggerInterval := 250 * time.Millisecond

	// Block until connection is established
	<-liftConnDoneCh

	for {
		select {
		case btn := <-btnPressCh:
			switch btn.Type {
			case HallUp:
				if time.Since(hallUpBtns[btn.Floor]) > cbTriggerInterval {
					cfg.OnBtnPress(btn)
					hallUpBtns[btn.Floor] = time.Now()
				}
			case HallDown:
				if time.Since(hallDownBtns[btn.Floor]) > cbTriggerInterval {
					cfg.OnBtnPress(btn)
					hallDownBtns[btn.Floor] = time.Now()
				}
			case Cab:
				if time.Since(CabBtns[btn.Floor]) > cbTriggerInterval {
					cfg.OnBtnPress(btn)
					insideBtnPressCh <- btn
					CabBtns[btn.Floor] = time.Now()
				}
			}
		}
	}
}

func floorDetectHandler(floorDetectCh <-chan int, apFloor chan<- int) {
	// Initialization
	beenDriving := true
	setBeenDriving := func(b bool) {
		beenDriving = b
	}

	// Block until connection is established
	<-liftConnDoneCh

	/*
		== WORKER LOOP ==
		Case 1: Incoming positive floor detection
			Case 1a: 	The carriage have been driving since the last positive detection
								--> Handle the detection as real thing
			Case 1b: 	The carriage have NOT been driving since the last positive detection
								--> Do nothing. The detection is already been handled

		Case 2: Incoming negative floor detection
						--> Set beenDriving = true
	*/
	for {
	selector:
		select {
		case floor := <-floorDetectCh:
			// Case 1
			if floor != -1 {
				// Case 1a
				if beenDriving {
					driverHandle.setFloorLED(floor)
					setBeenDriving(false)
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
