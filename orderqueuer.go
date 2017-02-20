package main

import (
	"time"

	"bitbucket.org/halvor_haukvik/ttk4145-elevator/driver"
)

type btnQueue struct {
	btns []driver.Btn
}

func (bq *btnQueue) Queue(b driver.Btn) {
	bq.btns = append(bq.btns, b)
}
func (bq *btnQueue) Dequeue() (b driver.Btn) {
	b = bq.btns[0]
	bq.btns = bq.btns[1:]
	return
}

func (bq *btnQueue) IsEmpty() bool {
	return len(bq.btns) == 0
}

func orderQueuer() {
	outsideQueue := btnQueue{}
	insideQueue := btnQueue{}
	ready := true
	insideTimeout := time.Now()

	for {
		select {
		// Listen for incomming orders
		case dst := <-goToCh:
			outsideQueue.Queue(dst)
		case dst := <-goToFromInsideCh:
			insideQueue.Queue(dst)

		// Listen for message that the last destination is reached.
		case <-orderDoneCh:
			ready = true
			insideTimeout = time.Now()
		case <-time.After(100 * time.Millisecond):
		}

		if !insideQueue.IsEmpty() && ready {
			dst := insideQueue.Dequeue()
			driver.GoToFloor(dst.Floor, "")
			ready = false
		} else if ready &&
			!outsideQueue.IsEmpty() &&
			// Give people inside priority if they press within 3 sec.
			time.Since(insideTimeout) > 3*time.Second {

			dst := outsideQueue.Dequeue()
			driver.GoToFloor(dst.Floor, dst.Type.String())
			ready = false
		}
	}
}
