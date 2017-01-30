package driver

import "time"

func simBtnScan() {
	sleeptime := 20 * time.Microsecond
	for {
		// Iterate over all buttons
		for f := 0; f < cfg.Floors; f++ {
			if f == 0 { // Special case: no HallDown in first floor
				for _, b := range []BtnType{HallUp, Cab} {
					if readOrderBtnSim(Btn{Floor: f, Type: b}) {
						btnPressCh <- Btn{Floor: f, Type: b}
					}
				}
			} else if f == cfg.Floors-1 { // Special case: no HallUp in top floor
				for _, b := range []BtnType{HallDown, Cab} {
					if readOrderBtnSim(Btn{Floor: f, Type: b}) {
						btnPressCh <- Btn{Floor: f, Type: b}
					}
				}
			} else { // All floors in between
				for _, b := range []BtnType{HallUp, HallDown, Cab} {
					if readOrderBtnSim(Btn{Floor: f, Type: b}) {
						btnPressCh <- Btn{Floor: f, Type: b}
					}
				}
			}
		}
		time.Sleep(sleeptime)
	}
}

func simFloorDetect() {
	sleeptime := 1 * time.Millisecond
	for {
		if atFloor, floor := readFloorSim(); atFloor {
			floorDetectCh <- floor
		} else {
			floorDetectCh <- -1
		}
		time.Sleep(sleeptime)
	}
}
