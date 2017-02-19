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
		case dst := <-goToCh:
			outsideQueue.Queue(dst)
			mainlogger.Println("Added to outside queue")
		case dst := <-goToFromInsideCh:
			mainlogger.Println("Added to inside queue")
			insideQueue.Queue(dst)
		case <-orderDoneCh:
			mainlogger.Println("liftDriver ready!")
			ready = true
			insideTimeout = time.Now()
		case <-time.After(100 * time.Millisecond):
		}

		if !insideQueue.IsEmpty() && ready {
			dst := insideQueue.Dequeue()
			mainlogger.Println("Took order from inside")
			driver.GoToFloor(dst.Floor, "")
			ready = false
		} else if ready &&
			!outsideQueue.IsEmpty() &&
			time.Since(insideTimeout) > 3*time.Second {

			mainlogger.Println("Took order from outside")
			dst := outsideQueue.Dequeue()
			driver.GoToFloor(dst.Floor, dst.Type.String())
			ready = false
		}
	}
}
