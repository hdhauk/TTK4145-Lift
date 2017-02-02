package driver

import "time"

func autoPilot(apFloorCh <-chan int) {
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
	case f := <-apFloorCh:
		lastFloor = f
		dstFloor = f
	case <-time.After(1 * time.Second):
		driver.setMotorDir(up)
		lastFloor = <-apFloorCh
		driver.setMotorDir(stop)
		currentDir = stop
		dstFloor = lastFloor
	}
	cfg.Logger.Printf("autopilot started with elevator stationary in floor: %v\n", lastFloor)

	// Start autopilot service
	for {
	selector:
		select {
		// Arrived at new floor
		case f := <-apFloorCh:
			lastFloor = f
			/*
				- Case 1: This is my destination
						-> stop -> open door
				- Case 2: This is NOT my destination
						-> Make sure the target destination is floorDstChin the direction of travel
						-> If everything is OK carry on
			*/
			cfg.Logger.Printf("autopilot.go: At floor: %v\t dstFloor: %v\n", f, dstFloor)

			// Case 1
			if f == dstFloor {
				driver.setMotorDir(stop)
				setCurrentDir(stop)
				openDoor()
				break selector
			}

			// Case 2
			if dirToDst(lastFloor, dstFloor) != currentDir {
				cfg.Logger.Printf("autopilot.go: Something unexpected have happend. Somehow ended up in a wrong direction. Turning around")
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
					cfg.Logger.Printf("autopilot.go: New destination given (%v). Case 1\n", dst)
					openDoor()
					break selector
				// Case 2a
				case up:
					cfg.Logger.Printf("autopilot.go: New destination given (%v). Case 2a\n", dst)
					driver.setMotorDir(down)
					setCurrentDir(down)
					break selector
				// Case 2b
				case down:
					cfg.Logger.Printf("autopilot.go: New destination given (%v). Case 2b\n", dst)
					driver.setMotorDir(up)
					setCurrentDir(up)
					break selector
				}
			}

			// Case 3
			cfg.Logger.Printf("autopilot.go: New destination given (%v). Case 3\n", dst)
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
	cfg.Logger.Println("Opening door...")
	driver.setDoorLED(true)
	time.Sleep(3 * time.Second)
	cfg.Logger.Println("Closing door...")
	driver.setDoorLED(false)
}
