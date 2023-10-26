package main

import (
	"sync"
)

// movingAvg keeps track of a history of measurements and returns the average.
// One important feature of this implementation is that the average can go to zero.
// All methods are thread safe.
//
// Alternative: consider exponential moving average where near-zero values are treated
// as zero (for scale to zero):
//
//	func MovingExpAvg(value, oldValue, fdtime, ftime float64) float64 {
//	 alpha := 1.0 - math.Exp(-fdtime/ftime)
//	 r := alpha * value + (1.0 - alpha) * oldValue
//	 return r
//	}
type movingAvg struct {
	mtx     sync.Mutex
	history []float64
	index   int
}

func newSimpleMovingAvg(seed []float64) *movingAvg {
	return &movingAvg{
		history: seed,
	}
}

func (a *movingAvg) Next(next float64) {
	a.mtx.Lock()
	a.history[a.index] = next
	a.index++
	if a.index == len(a.history) {
		a.index = 0
	}
	a.mtx.Unlock()
}

func (a *movingAvg) Calculate() (result float64) {
	a.mtx.Lock()
	for _, p := range a.history {
		result += p
	}
	result /= float64(len(a.history))
	a.mtx.Unlock()

	return result
}
