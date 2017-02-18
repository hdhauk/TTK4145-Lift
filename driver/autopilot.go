package driver

import "time"

const (
	red    = "\x1b[31;1m"
	white  = "\x1b[0m"
	yellow = "\x1b[33;1m"
)

func autoPilot(apFloorCh <-chan int, driverInitDone chan error) {
	currentDir := stop
	// Need function to get correct scoping in for-select loop
	setCurrentDir := func(dir string) {
		currentDir = dir
	}

	<-liftConnDoneCh
	clearAllBtns()
	driverHandle.setDoorLED(false)

	// Drive up to a well defined floor
	var lastFloor int
	var currentDst dst
	select {
	case f := <-apFloorCh:
		lastFloor = f
		currentDst = dst{floor: f, dir: ""}
	case <-time.After(1 * time.Second):
		driverHandle.setMotorDir(up)
		lastFloor = <-apFloorCh
		driverHandle.setMotorDir(stop)
		currentDir = stop
		currentDst.floor = lastFloor
	}
	cfg.Logger.Printf("[INFO] Ready with elevator stationary in floor: %v\n", lastFloor)
	close(driverInitDone)

	// Pickup destination
	var pickup dst
	var setPickup = func(b dst) { pickup = b }

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
				- Case 3: This is NOT my destination, BUT i can pick someone up here.
			*/
			// Case 1
			if f == currentDst.floor {
				setCurrentDir(stop)
				driverHandle.setMotorDir(stop)
				go cfg.OnDstReached(newBtn(currentDst.floor, currentDst.dir))
				currentDst.dir = ""
				openDoor()
				go cfg.OnNewStatus(lastFloor, currentDir, currentDst.floor, currentDst.dir)
				break selector
			}

			// Case 2
			if dirToDst(f, currentDst.floor) != currentDir {
				newDir := dirToDst(lastFloor, currentDst.floor)
				setCurrentDir(newDir)
				cfg.Logger.Printf(yellow+"[WARN] Unexpected direction value. Correcting to: %s"+white, newDir)
				driverHandle.setMotorDir(newDir)
			}

			// Trigger new status callback
			go cfg.OnNewStatus(lastFloor, currentDir, currentDst.floor, currentDst.dir)

		// New destination given
		case d := <-floorDstCh:
			currentDst = d
			/*
				- Case 1: My new destination is coincidentaly the elevator currenly is parked
						-> Open the door
				- Case 2: My new destination is the floor that I just left
						-> Turn back
				- Case 3: My new destination is somewhere else
						-> Determine in what direction the target is -> Go in that direction
			*/
			if currentDst.floor == lastFloor {
				switch currentDir {
				// Case 1
				case stop:
					go cfg.OnDstReached(newBtn(currentDst.floor, currentDst.dir))
					currentDst.dir = ""
					openDoor()
					break selector
				// Case 2a
				case up:
					// NOTE: The timer make sure that the elevator have actually left
					// the sensor. Otherwise it will not trigger the floor sensor,
					// and in rare cases will end up going beyond the area of operation.
					time.Sleep(200 * time.Millisecond)

					driverHandle.setMotorDir(down)
					setCurrentDir(down)
					break selector
				// Case 2b
				case down:
					// NOTE: The timer make sure that the elevator have actually left
					// the sensor. Otherwise it will not trigger the floor sensor,
					// and in rare cases will end up going beyond the area of operation.
					time.Sleep(200 * time.Millisecond)

					driverHandle.setMotorDir(up)
					setCurrentDir(up)
					break selector
				}
			}

			// Case 3
			d2d := dirToDst(lastFloor, d.floor)
			driverHandle.setMotorDir(d2d)
			setCurrentDir(d2d)

			// Trigger new status callback
			go cfg.OnNewStatus(lastFloor, currentDir, currentDst.floor, currentDst.dir)

		// Periodically send status updates
		case <-time.After(4 * time.Second):
			go cfg.OnNewStatus(lastFloor, currentDir, currentDst.floor, currentDst.dir)

		// Pickup requests
		case p := <-stopForPickupCh:
			// Make sure it is safe to stop and that the lift acutally is in this floor.
			atFloor, f := driverHandle.readFloor()
			if !atFloor {
				cfg.Logger.Println(yellow + "[WARN] Cannot stop for pickup outside a floor. Pickup aborted." + white)
				break selector
			} else if f != p.floor {
				cfg.Logger.Printf("%s[WARN] Pickup floor and current floor do not match (%d != %d). Pickup aborted.%s\n", yellow, f, p.floor, white)
				break selector
			}

			// Do the pickup, and carry on.
			setPickup(p)
			driverHandle.setMotorDir(stop)
			go cfg.OnDstReached(newBtn(pickup.floor, pickup.dir))
			go cfg.OnNewStatus(lastFloor, stop, currentDst.floor, currentDst.dir)
			openDoor()
			driverHandle.setMotorDir(currentDir)
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
	driverHandle.setDoorLED(true)
	time.Sleep(3 * time.Second)
	driverHandle.setDoorLED(false)
}

func newBtn(f int, dir string) Btn {
	if dir == "down" || dir == "DOWN" {
		return Btn{Floor: f, Type: HallDown}
	} else if dir == "up" || dir == "UP" {
		return Btn{Floor: f, Type: HallUp}
	}
	return Btn{}
}
