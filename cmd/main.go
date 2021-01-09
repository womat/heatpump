package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/womat/debug"

	"heatpump/global"
	"heatpump/pkg/heatpump"
)

type heatPumpRuntime struct {
	sync.RWMutex
	data                      *heatpump.Measurements
	lastState                 heatpump.State
	lastRuntime               float64
	lastStateDate             time.Time
	lastBrinePumpState        heatpump.State
	lastBrinePumpStateDate    time.Time
	lastBrinePumpRuntime      float64
	lastHeatingPumpState      heatpump.State
	lastHeatingPumpStateDate  time.Time
	lastHeatingPumpRuntime    float64
	lastHotWaterPumpState     heatpump.State
	lastHotWaterPumpStateDate time.Time
	lastHotWaterPumpRuntime   float64
}

func main() {
	debug.SetDebug(global.Config.Debug.File, global.Config.Debug.Flag)

	global.Measurements = heatpump.New()
	global.Measurements.SetMeterURL(global.Config.MeterURL)
	global.Measurements.SetUVS232URL(global.Config.UVS232URL)

	if err := loadMeasurements(global.Config.DataFile, global.Measurements); err != nil {
		debug.ErrorLog.Printf("can't open data file: %v\n", err)
		os.Exit(1)
		return
	}

	runtime := &heatPumpRuntime{
		data:          global.Measurements,
		lastState:     heatpump.Off,
		lastRuntime:   global.Measurements.Runtime,
		lastStateDate: time.Now(),

		lastBrinePumpState:     heatpump.Off,
		lastBrinePumpStateDate: time.Now(),
		lastBrinePumpRuntime:   global.Measurements.BrinePumpRuntime,

		lastHeatingPumpState:     heatpump.Off,
		lastHeatingPumpStateDate: time.Now(),
		lastHeatingPumpRuntime:   global.Measurements.HeatingPumpRuntime,

		lastHotWaterPumpState:     heatpump.Off,
		lastHotWaterPumpStateDate: time.Now(),
		lastHotWaterPumpRuntime:   global.Measurements.HotWaterPumpRuntime,
	}

	go runtime.calcRuntime(global.Config.DataCollectionInterval)
	go runtime.backupMeasurements(global.Config.DataFile, global.Config.BackupInterval)

	// capture exit signals to ensure resources are released on exit.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	// wait for am os.Interrupt signal (CTRL C)
	sig := <-quit
	debug.InfoLog.Printf("Got %s signal. Aborting...\n", sig)
	_ = saveMeasurements(global.Config.DataFile, global.Measurements)
	os.Exit(1)
}
