package main

import (
	"time"

	"github.com/womat/debug"

	"heatpump/pkg/heatpump"
)

func (r *heatPumpRuntime) backupMeasurements(f string, p time.Duration) {
	for range time.Tick(p) {
		_ = saveMeasurements(f, r.data)
	}
}

func (r *heatPumpRuntime) calcRuntime(p time.Duration) {
	runtime := func(state heatpump.State, lastStateDate *time.Time, lastState *heatpump.State) (runTime float64) {
		if in(state, heatpump.On, heatpump.Heating, heatpump.Cooling, heatpump.HeatingUpHotWater) {
			if !in(*lastState, heatpump.On, heatpump.Heating, heatpump.Cooling, heatpump.HeatingUpHotWater) {
				*lastStateDate = time.Now()
			}
			runTime = time.Since(*lastStateDate).Hours()
			*lastStateDate = time.Now()
		}
		*lastState = state
		return
	}

	ticker := time.NewTicker(p)
	defer ticker.Stop()

	for ; true; <-ticker.C {
		debug.DebugLog.Println("get data")

		temp := heatpump.New()
		*temp = *r.data

		if err := temp.Read(); err != nil {
			debug.ErrorLog.Printf("get heatpump data: %v", err)
			continue
		}

		func() {
			debug.DebugLog.Println("calc runtime")
			r.Lock()
			defer r.Unlock()

			*r.data = *temp
			r.data.Runtime += runtime(r.data.State, &r.lastStateDate, &r.lastState)
			r.data.BrinePumpRuntime += runtime(r.data.BrinePumpState, &r.lastBrinePumpStateDate, &r.lastBrinePumpState)
			r.data.HeatingPumpRuntime += runtime(r.data.HeatingPumpState, &r.lastHeatingPumpStateDate, &r.lastHeatingPumpState)
			r.data.HotWaterPumpRuntime += runtime(r.data.HotWaterPumpState, &r.lastHotWaterPumpStateDate, &r.lastHotWaterPumpState)
		}()
	}
}

func in(s interface{}, pattern ...interface{}) bool {
	for _, p := range pattern {
		if s == p {
			return true
		}
	}
	return false
}
