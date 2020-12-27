package global

import (
	"io"
	"sync"
	"time"
)

// VERSION holds the version information with the following logic in mind
//  1 ... fixed
//  0 ... year 2020, 1->year 2021, etc.
//  7 ... month of year (7=July)
//  the date format after the + is always the first of the month
//
// VERSION differs from semantic versioning as described in https://semver.org/
// but we keep the correct syntax.
//TODO: increase version number to 1.0.1+2020xxyy
const VERSION = "0.0.1+20201227"
const MODULE = "heatpump"

type DebugConf struct {
	File io.WriteCloser
	Flag int
}

type WebserverConf struct {
	Port        int             `yaml:"port"`
	Webservices map[string]bool `yaml:"webservices"`
}

type Configuration struct {
	DataCollectionInterval time.Duration
	Debug                  DebugConf
	Webserver              WebserverConf
}

const (
	on  = "on"
	off = "off"
)

type State string

type HeatPump struct {
	sync.RWMutex
	TimeStamp                    time.Time
	Power                        float64
	State                        State
	Runtime                      time.Duration
	BrineFlow, BrineReturn       float64
	BrinePumpState               State
	BrinePumpRuntime             time.Duration
	HeatingFlow, HeatingReturn   float64
	HeatingPumpState             State
	HeatingPumpRuntime           time.Duration
	HotWaterFlow, HotWaterReturn float64
	HotWaterPumpState            State
	HotWaterPumpRuntime          time.Duration
}

// Config holds the global configuration
var Config Configuration
var Measurements HeatPump

func init() {
	Measurements = HeatPump{}
	Config = Configuration{
		Webserver: WebserverConf{Webservices: map[string]bool{}},
	}
}
