package main

import (
	"log"
	"math"
	"sync"
	"time"
)

func NewAutoscaler() *Autoscaler {
	return &Autoscaler{movingAvgQueueSize: map[string]*movingAvg{}}
}

// Autoscaler is responsible for making continuous adjustments to
// the scale of the backend. It is not responsible for scale-from-zero.
type Autoscaler struct {
	Interval     time.Duration
	AverageCount int

	Scaler *ScalerManager
	// TODO: FIFOQueueManager should manage per backend not per model.
	FIFO *FIFOQueueManager

	movingAvgQueueSizeMtx sync.Mutex
	movingAvgQueueSize    map[string]*movingAvg
}

func (a *Autoscaler) Start() {
	for range time.Tick(a.Interval) {
		log.Println("Calculating scales for all")
		for model, waitCount := range a.FIFO.WaitCounts() {
			if model == "proxy-controller" {
				// TODO: Remove this after selecting models based on labels.
				continue
			}
			avg := a.getMovingAvgQueueSize(model)
			avg.Next(float64(waitCount))
			flt := avg.Calculate()
			ceil := math.Ceil(flt)
			log.Printf("Average for model: %s: %v (ceil: %v), current wait count: %v", model, flt, ceil, waitCount)
			a.Scaler.SetDesiredScale(model, int32(ceil))
		}
	}
}

func (r *Autoscaler) getMovingAvgQueueSize(model string) *movingAvg {
	r.movingAvgQueueSizeMtx.Lock()
	a, ok := r.movingAvgQueueSize[model]
	if !ok {
		a = newSimpleMovingAvg(make([]float64, r.AverageCount))
		r.movingAvgQueueSize[model] = a
	}
	r.movingAvgQueueSizeMtx.Unlock()
	return a
}
