package gmeter

import (
	"fmt"
	"time"
)

type Meter struct {
	id           int
	lastStart    time.Time
	successItems []int64
	failedItems  []int64
	finishNum    int
}

func NewMeter(id int) *Meter {
	return &Meter{
		id: id,
	}
}

func (meter *Meter) Start() {
	meter.lastStart = time.Now()
}

func (meter *Meter) Finish(result map[string]interface{}, err error) {
	meter.finishNum += 1
	ms := time.Since(meter.lastStart).Milliseconds()
	if err != nil {
		meter.failedItems = append(meter.failedItems, ms)
		meter.Failed(err)
	} else {
		meter.successItems = append(meter.successItems, ms)
		meter.Success(result)
	}
}

func (meter *Meter) Success(result map[string]interface{}) {
}

func (meter *Meter) Failed(err error) {
}

func (meter *Meter) Extend(other *Meter) {
	if other == nil || other.finishNum == 0 {
		return
	}
	meter.finishNum += other.finishNum
	meter.successItems = append(meter.successItems, other.successItems...)
	meter.failedItems = append(meter.failedItems, other.failedItems...)
}

func div(num1 int64, num2 int64) float64 {
	if num2 == 0 {
		return 0
	}
	return float64(num1) / float64(num2)
}

func (meter *Meter) Summary() {
	if meter.finishNum == 0 {
		return
	}
	successCost := sum(meter.successItems)
	failedCost := sum(meter.failedItems)

	fmt.Printf("client: %v\n", meter.id)

	fmt.Printf("    all cost %vms process %v request averagy %vms\n",
		successCost+failedCost, meter.finishNum, div(successCost+failedCost, int64(meter.finishNum)))
	fmt.Printf("    success cost %vms process %v request averagy %vms\n",
		successCost, len(meter.successItems), div(successCost, int64(len(meter.successItems))))
	fmt.Printf("    failed cost %vms process %v request averagy %vms\n",
		failedCost, len(meter.failedItems), div(failedCost, int64(len(meter.failedItems))))
}

func sum(nums []int64) int64 {
	sum := int64(0)
	for _, num := range nums {
		sum += num
	}
	return sum
}
