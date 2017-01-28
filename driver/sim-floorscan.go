package driver

import "time"

var floorDetectCh = make(chan int)

func simFloorDetect() {
	sleeptime := 20 * time.Microsecond
	for {
		if atFloor, floor := readFloorSim(); atFloor {
			floorDetectCh <- floor
		}
		time.Sleep(sleeptime)
	}
}
