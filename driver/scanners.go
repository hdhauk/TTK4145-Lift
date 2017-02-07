package driver

import "time"

func btnScan(btnPressCh chan<- Btn) {
	sleeptime := 20 * time.Microsecond
	for {
		// Iterate over all buttons
		for f := 0; f < cfg.Floors; f++ {
			if f == 0 { // Special case: no HallDown in first floor
				for _, b := range []BtnType{HallUp, Cab} {
					if driver.readOrderBtn(Btn{Floor: f, Type: b}) {
						btnPressCh <- Btn{Floor: f, Type: b}
					}
				}
			} else if f == cfg.Floors-1 { // Special case: no HallUp in top floor
				for _, b := range []BtnType{HallDown, Cab} {
					if driver.readOrderBtn(Btn{Floor: f, Type: b}) {
						btnPressCh <- Btn{Floor: f, Type: b}
					}
				}
			} else { // All floors in between
				for _, b := range []BtnType{HallUp, HallDown, Cab} {
					if driver.readOrderBtn(Btn{Floor: f, Type: b}) {
						btnPressCh <- Btn{Floor: f, Type: b}
					}
				}
			}
		}
		time.Sleep(sleeptime)
	}
}

func floorDetect(floorDetectCh chan<- int) {
	sleeptime := 1 * time.Millisecond
	for {
		if atFloor, floor := driver.readFloor(); atFloor {
			floorDetectCh <- floor
		} else {
			floorDetectCh <- -1
		}
		time.Sleep(sleeptime)
	}
}
