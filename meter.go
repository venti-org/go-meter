package gmeter

import (
	"fmt"
	"time"
)

type Meter struct {
	id           int
	children     int
	start        time.Time
	finish       time.Time
	lastStart    time.Time
	successItems []int64
	failedItems  []int64
	finishNum    int
}

func NewMeter(id int) *Meter {
	return &Meter{
		id:    id,
		start: time.Now(),
	}
}

func (meter *Meter) getClientCount() int {
	n := 0
	if !meter.lastStart.IsZero() {
		n += 1
	}
	return n + meter.children
}

func (meter *Meter) Start() {
	meter.lastStart = time.Now()
}

func (meter *Meter) Finish(res *Response) {
	meter.finish = time.Now()
	meter.finishNum += 1
	ms := time.Since(meter.lastStart).Milliseconds()
	if res.Error != nil {
		meter.failedItems = append(meter.failedItems, ms)
		meter.Failed(res)
	} else {
		meter.successItems = append(meter.successItems, ms)
		meter.Success(res)
	}
}

func (meter *Meter) Success(res *Response) {
	fmt.Println(res.String())
}

func (meter *Meter) Failed(res *Response) {
	ErrPrintln(res.String())
}

func (meter *Meter) Extend(other *Meter) {
	if other == nil || other.finishNum == 0 {
		return
	}
	meter.children += other.getClientCount()
	if other.start.Before(meter.start) {
		meter.start = other.start
	}
	if other.finish.After(meter.finish) {
		meter.finish = other.finish
	}
	meter.finishNum += other.finishNum
	meter.successItems = append(meter.successItems, other.successItems...)
	meter.failedItems = append(meter.failedItems, other.failedItems...)
}

func (meter *Meter) Summary() {
	if meter.finishNum == 0 {
		return
	}
	successCost := sum(meter.successItems)
	minSuccessCost := minWithDefault(0, meter.successItems...)
	maxSuccessCost := maxWithDefault(0, meter.successItems...)
	failedCost := sum(meter.failedItems)
	minFailedCost := minWithDefault(0, meter.failedItems...)
	maxFailedCost := maxWithDefault(0, meter.failedItems...)

	costMs := meter.finish.Sub(meter.start).Milliseconds()

	ErrPrintf("client%v: (%v clients real cost %vms process %v request qps %.2f\n", meter.id,
		meter.getClientCount(), costMs, meter.finishNum, div(int64(meter.finishNum)*1000, costMs))

	ErrPrintf("    all cost %vms process %v request averagy %vms max %vms min %vms\n",
		successCost+failedCost, meter.finishNum, div(successCost+failedCost, int64(meter.finishNum)),
		max(maxSuccessCost, maxFailedCost), min(minSuccessCost, minFailedCost))
	ErrPrintf("    success cost %vms process %v request averagy %vms max %vms min %vms\n",
		successCost, len(meter.successItems), div(successCost, int64(len(meter.successItems))),
		maxSuccessCost, minSuccessCost)
	ErrPrintf("    failed cost %vms process %v request averagy %vms max %vms min %vms\n",
		failedCost, len(meter.failedItems), div(failedCost, int64(len(meter.failedItems))),
		maxFailedCost, minFailedCost)
}
