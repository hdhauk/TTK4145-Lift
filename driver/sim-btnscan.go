package driver

import "time"

var btnPressCh = make(chan Btn, 4)

func simBtnScan() {
	sleeptime := 20 * time.Microsecond
	for {
		// Iterate over all buttons
		for f := 0; f < cfg.floors; f++ {
			if f == 0 { // Special case: no HallDown in first floor
				for _, b := range []BtnType{HallUp, Cab} {
					if readOrderBtnSim(Btn{Floor: f, Type: b}) {
						btnPressCh <- Btn{Floor: f, Type: b}
					}
				}
			} else if f == cfg.floors-1 { // Special case: no HallUp in top floor
				for _, b := range []BtnType{HallDown, Cab} {
					if readOrderBtnSim(Btn{Floor: f, Type: b}) {
						btnPressCh <- Btn{Floor: f, Type: b}
					}
				}
			} else { // All floors inbetween
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
