package driver

import "time"

const (
	red    = "\x1b[31;1m"
	white  = "\x1b[0m"
	yellow = "\x1b[33;1m"
)

func autoPilot(apFloorCh <-chan int, driverInitDone chan struct{}) {
	// lastFloor := 0
	// dstFloor := 0
	currentDir := stop
	// Need function to get correct scoping in for-select loop
	setCurrentDir := func(dir string) {
		currentDir = dir
	}

	<-liftConnDone

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
	cfg.Logger.Printf("[INFO] Ready with elevator stationary in floor: %v\n", lastFloor)
	close(driverInitDone)

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
			cfg.Logger.Printf("[INFO] At floor: %v\t dstFloor: %v\n", f, dstFloor)

			// Case 1
			if f == dstFloor {
				driver.setMotorDir(stop)
				setCurrentDir(stop)
				cfg.OnDstReached(f)
				//openDoor()
				break selector
			}

			// Case 2
			if dirToDst(f, dstFloor) != currentDir {
				newDir := dirToDst(lastFloor, dstFloor)
				cfg.Logger.Printf(yellow+"[WARN] Unexpected direction value. Correcting to: %s"+white, newDir)
				driver.setMotorDir(newDir)
				setCurrentDir(newDir)
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
					cfg.Logger.Printf("[INFO] New destination given (%v). Case 1: Stopping...and opening door.\n", dst)
					cfg.OnDstReached(lastFloor)
					break selector
				// Case 2a
				case up:
					cfg.Logger.Printf("autopilot.go: New destination given (%v). Case 2a: Going down...\n", dst)
					// NOTE: The timer make sure that the elevator have actually left
					// the sensor. Otherwise it will not trigger the floor sensor,
					// and in rare cases will end up going beyond the area of operation.
					time.Sleep(200 * time.Millisecond)

					driver.setMotorDir(down)
					setCurrentDir(down)
					break selector
				// Case 2b
				case down:
					cfg.Logger.Printf("autopilot.go: New destination given (%v). Case 2b: Going up...\n", dst)
					// NOTE: The timer make sure that the elevator have actually left
					// the sensor. Otherwise it will not trigger the floor sensor,
					// and in rare cases will end up going beyond the area of operation.
					time.Sleep(200 * time.Millisecond)

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
