package driver

var floorDstCh = make(chan int)
var apFloor = make(chan int)

func autoPilot() {
	lastFloor := 0
	dstFloor := 0
	currentDir := stop

	for {
		select {
		// Arrived at new floor
		case f := <-apFloor:
			if f == dstFloor {
				driver.setMotorDir(stop)
				currentDir = stop
			}
			if d2d := dirToDst(lastFloor, f); d2d != currentDir {
				driver.setMotorDir(d2d)
				currentDir = d2d
			}
		// New destination given
		case dst := <-floorDstCh:
			dstFloor = dst
			if d2d := dirToDst(lastFloor, dstFloor); d2d != currentDir {
				driver.setMotorDir(d2d)
				currentDir = d2d
			}
		}
	}
}

func dirToDst(lastFloor, dst int) string {
	if lastFloor < dst {
		return up
	}
	if lastFloor > dst {
		return down
	}
	return stop
}
