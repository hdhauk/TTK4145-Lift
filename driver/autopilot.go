package driver

import (
	"fmt"
	"time"
)

var floorDstCh = make(chan int)
var apFloor = make(chan int)

func autoPilot() {
	// lastFloor := 0
	// dstFloor := 0
	currentDir := stop
	// Need function to get correct scoping in for-select loop
	setCurrentDir := func(dir string) {
		currentDir = dir
	}

	<-initDone

	// Drive up to a well defined floor
	var lastFloor, dstFloor int
	select {
	case f := <-apFloor:
		lastFloor = f
		dstFloor = f
	case <-time.After(1 * time.Second):
		driver.setMotorDir(up)
		lastFloor = <-apFloor
		driver.setMotorDir(stop)
		currentDir = stop
		dstFloor = lastFloor
	}

	fmt.Printf("autopilot.go: Entering for-loop with [lastfloor:%v dstFloor:%v]\n", lastFloor, dstFloor)
	for {
	selector:
		select {
		// Arrived at new floor
		case f := <-apFloor:
			lastFloor = f
			/*
				- Case 1: This is my destination
						-> stop -> open door
				- Case 2: This is NOT my destination
						-> Make sure the target destination is in the direction of travel
						-> If everything is OK carry on
			*/
			fmt.Printf("autopilot.go: At floor: %v\t dstFloor: %v\n", f, dstFloor)

			// Case 1
			if f == dstFloor {
				driver.setMotorDir(stop)
				setCurrentDir(stop)
				openDoor()
				break selector
			}

			// Case 2
			if dirToDst(lastFloor, dstFloor) != currentDir {
				fmt.Printf("autopilot.go: Something unexpected have happend. Somehow ended up in a wrong direction. Turning around")
				driver.setMotorDir(dirToDst(lastFloor, dstFloor))
				setCurrentDir(dirToDst(lastFloor, dstFloor))
			}

		// New destination given
		case dst := <-floorDstCh:
			dstFloor = dst
			/*
				- Case 1: My new destination is coincidentaly the elevator currenly is parked
						-> Open the door
				- Case 2: My new destination is the floor that I just left
						-> Turn back
				- Case 3: My new destination is somewhere else
						-> Determine in what direction the target is -> Go in that direction
			*/
			if dst == lastFloor {
				switch currentDir {
				// Case 1
				case stop:
					fmt.Printf("autopilot.go: New destination given (%v). Case 1\n", dst)
					openDoor()
					break selector
				// Case 2a
				case up:
					fmt.Printf("autopilot.go: New destination given (%v). Case 2a\n", dst)
					driver.setMotorDir(down)
					setCurrentDir(down)
					break selector
				// Case 2b
				case down:
					fmt.Printf("autopilot.go: New destination given (%v). Case 2b\n", dst)
					driver.setMotorDir(up)
					setCurrentDir(up)
					break selector
				}
			}

			// Case 3
			fmt.Printf("autopilot.go: New destination given (%v). Case 3\n", dst)
			d2d := dirToDst(lastFloor, dst)
			driver.setMotorDir(d2d)
			setCurrentDir(d2d)
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

func openDoor() {
	driver.setDoorLED(true)
	time.Sleep(3 * time.Second)
	driver.setDoorLED(false)
}
