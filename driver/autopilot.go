package driver

import "fmt"

var floorDstCh = make(chan int)
var apFloor = make(chan int)

func autoPilot() {
	lastFloor := 0
	dstFloor := <-apFloor
	currentDir := stop

	<-initDone
	for {
		select {
		// Arrived at new floor
		case f := <-apFloor:
			fmt.Printf("autopilot.go/autopilot(): Recived new destination floor %v\n\t, Current destination == %v\n", f, dstFloor)
			if f == dstFloor {
				driver.setMotorDir(stop)
				currentDir = stop
				break
			}
			if d2d := dirToDst(lastFloor, f); d2d != currentDir {
				driver.setMotorDir(d2d)
				currentDir = d2d
			}

		// New destination given
		case dst := <-floorDstCh:
			if dst == dstFloor {
				fmt.Printf("autopilot.go/autopilot(): New destination is the same as before, %v\n", dst)
				break
			}
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
