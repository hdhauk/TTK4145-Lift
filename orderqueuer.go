package main

import (
	"time"

	"github.com/hdhauk/TTK4145-Lift/driver"
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
	ready := true

	for {
		select {
		// Listen for incoming orders
		case dst := <-goToCh:
			outsideQueue.Queue(dst)
		// Listen for message that the last destination is reached.
		case <-orderDoneCh:
			ready = true
		case <-time.After(100 * time.Millisecond):
		}
		if !outsideQueue.IsEmpty() && ready {
			dst := outsideQueue.Dequeue()
			driver.GoToFloor(dst.Floor, dst.Type.String())
			ready = false
		}
	}
}
