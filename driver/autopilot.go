package driver

import (
	"time"
)

const (
	red    = "\x1b[31;1m"
	white  = "\x1b[0m"
	yellow = "\x1b[33;1m"
)

func autoPilot(apFloorCh <-chan int, driverInitDone chan error) {
	// State variables
	currentDir := stop
	var lastFloor int
	currentInsideDst := -1
	var currentOutsideDst dst
	insideBtns := []bool{}
	for i := 0; i < cfg.Floors; i++ {
		insideBtns = append(insideBtns, false)
	}

	// Wait for driver connection
	<-liftConnDoneCh
	clearAllBtns()
	driverHandle.setDoorLED(false)

	// Make sure we are in a well-defined known floor
	select {
	case f := <-apFloorCh:
		lastFloor = f
		currentOutsideDst = dst{floor: f, dir: ""}
	case <-time.After(1 * time.Second):
		driverHandle.setMotorDir(up)
		lastFloor = <-apFloorCh
		driverHandle.setMotorDir(stop)
		currentDir = stop
		currentOutsideDst.floor = lastFloor
	}
	cfg.Logger.Printf("[INFO] Ready with lift stationary in floor: %v\n", lastFloor)
	close(driverInitDone)

	for {
	selector:
		select {
		case lastFloor = <-apFloorCh:
			if lastFloor == currentInsideDst {
				insideBtns[lastFloor] = false
				driverHandle.setBtnLED(Btn{lastFloor, Cab}, false)
				currentInsideDst = -1

				// Check if happend to also be the outside destination
				if currentOutsideDst.floor == lastFloor {
					go cfg.OnDstReached(newBtn(lastFloor, currentOutsideDst.dir), false)
					currentOutsideDst = dst{-1, ""}
				}
				openDoor()
			} else if insideBtns[lastFloor] {
				insideBtns[lastFloor] = false
				driverHandle.setBtnLED(Btn{lastFloor, Cab}, false)

				// Check if happend to also be the outside destination
				if currentOutsideDst.floor == lastFloor {
					go cfg.OnDstReached(newBtn(lastFloor, currentOutsideDst.dir), false)
					currentOutsideDst = dst{-1, ""}
				}
				openDoor()
			} else if currentOutsideDst.floor == lastFloor {
				go cfg.OnDstReached(newBtn(lastFloor, currentOutsideDst.dir), false)
				currentOutsideDst = dst{-1, ""}
				openDoor()
			}

		case d := <-floorDstCh:
			currentOutsideDst = dst{d.floor, d.dir}
		case p := <-stopForPickupCh:
			// Make sure that it is safe to stop and that the lift actually is at this floor
			atFloor, f := driverHandle.readFloor()
			if !atFloor {
				cfg.Logger.Println(yellow + "[WARN] Cannot stop for pickup outside a floor. Pickup aborted." + white)
				break selector
			} else if f != p.floor {
				cfg.Logger.Printf("%s[WARN] Pickup floor and current floor do not match (%d != %d). Pickup aborted.%s\n", yellow, f, p.floor, white)
				break selector
			}

			// Otherwise do the pickup and carry on
			driverHandle.setMotorDir(stop)
			go cfg.OnDstReached(newBtn(p.floor, p.dir), true)
			go cfg.OnNewStatus(lastFloor, stop, currentOutsideDst.floor, currentOutsideDst.dir)
			if insideBtns[f] {
				driverHandle.setBtnLED(Btn{f, Cab}, false)
				insideBtns[f] = false
			}
			openDoor()
			driverHandle.setMotorDir(currentDir)

		case b := <-insideBtnPressCh:
			insideBtns[b.Floor] = true
		case <-time.After(4 * time.Second):
			go cfg.OnNewStatus(lastFloor, currentDir, currentOutsideDst.floor, currentOutsideDst.dir)

		}

		// Determine what to do next
		if currentInsideDst != -1 {
			currentDir = dirToDst(lastFloor, currentInsideDst)
			if currentDir == stop {
				currentInsideDst = -1
				driverHandle.setBtnLED(Btn{lastFloor, Cab}, false)
				openDoor()
			}
		} else if f := getFurthestAway(insideBtns, lastFloor); f != -1 {
			currentInsideDst = f
			currentDir = dirToDst(lastFloor, f)
		} else if currentOutsideDst.floor != -1 {
			currentDir = dirToDst(lastFloor, currentOutsideDst.floor)
			// If the door should open here
			if currentDir == stop {
				go cfg.OnDstReached(newBtn(currentOutsideDst.floor, currentOutsideDst.dir), false)
				currentOutsideDst = dst{-1, ""}
				openDoor()
			}
		} else {
			currentDir = stop
		}
		driverHandle.setMotorDir(currentDir)
	}
}

func getFurthestAway(btns []bool, currentFloor int) int {
	min := 1000
	max := -1

	for k, v := range btns {
		if v && k < min {
			min = k
		}
		if v && k > max {
			max = k
		}
	}
	// Check if there were any btns pressed.
	if max == -1 {
		return -1
	}

	if max-currentFloor >= currentFloor-min {
		return max
	}
	return min
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

func clearAllBtns() {
	for i := 0; i < cfg.Floors-1; i++ {
		b := Btn{i, HallUp}
		driverHandle.setBtnLED(b, false)
	}
	for i := 1; i < cfg.Floors; i++ {
		driverHandle.setBtnLED(Btn{i, HallDown}, false)
	}
	for i := 0; i < cfg.Floors; i++ {
		driverHandle.setBtnLED(Btn{i, Cab}, false)
	}
}
