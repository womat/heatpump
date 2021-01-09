package heatpump

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/womat/debug"
)

const httpRequestTimeout = 10 * time.Second

const (
	On                State = "on"
	Off               State = "off"
	Heating           State = "heating"
	Cooling           State = "cooling"
	HeatingUpHotWater State = "heating up hot water"
	HeatRecovery      State = "heat recovery"

	ThresholdHeatPump     = 2000
	ThresholdBrinePump    = 200
	ThresholdHeatingPump  = 30
	ThresholdHotWaterPump = 0
)

type State string

type Measurements struct {
	sync.RWMutex
	Timestamp                    time.Time
	Power                        float64
	State                        State
	Runtime                      float64
	BrineFlow, BrineReturn       float64
	BrinePumpState               State
	BrinePumpRuntime             float64
	HeatingFlow, HeatingReturn   float64
	HeatingPumpState             State
	HeatingPumpRuntime           float64
	HotWaterFlow, HotWaterReturn float64
	HotWaterPumpState            State
	HotWaterPumpRuntime          float64
	config                       struct {
		uvs232URL, meterURL string
	}
}

type uvs232URLBody struct {
	Timestamp time.Time `json:"Timestamp"`
	Runtime   float64   `json:"Runtime"`
	Measurand struct {
		Temperature1, Temperature2, Temperature3, Temperature4 float64
		Out1, Out2                                             bool
		RotationSpeed                                          float64
	} `json:"Data"`
}

type meterURLBody struct {
	Timestamp time.Time `json:"Time"`
	Runtime   float64   `json:"Runtime"`
	Measurand struct {
		E float64 `json:"e"`
		P float64 `json:"p"`
	} `json:"Measurand"`
}

func New() *Measurements {
	return &Measurements{}
}

func (m *Measurements) SetUVS232URL(url string) {
	m.config.uvs232URL = url
}

func (m *Measurements) SetMeterURL(url string) {
	m.config.meterURL = url
}

func (m *Measurements) Read() (err error) {
	var wg sync.WaitGroup
	start := time.Now()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := m.readUVS232(); e != nil {
			err = e
		}

		debug.TraceLog.Printf("runtime to request UVS232 data: %vs", time.Since(start).Seconds())
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if e := m.readMeter(); e != nil {
			err = e
		}

		debug.TraceLog.Printf("runtime to request meter data: %vs", time.Since(start).Seconds())
	}()

	wg.Wait()
	debug.DebugLog.Printf("runtime to request data: %vs", time.Since(start).Seconds())

	m.Lock()
	defer m.Unlock()
	if m.Power > ThresholdHeatingPump {
		m.HeatingPumpState = On
	} else {
		m.HeatingPumpState = Off
	}

	if m.Power > ThresholdBrinePump {
		m.BrinePumpState = On
	} else {
		m.BrinePumpState = Off
	}

	switch {
	case m.Power > ThresholdHeatPump:
		m.State = On
	case m.Power > ThresholdBrinePump:
		m.State = HeatRecovery
	default:
		m.State = Off
	}

	m.Timestamp = time.Now()
	return
}

func (m *Measurements) readUVS232() (err error) {
	var r uvs232URLBody

	if err = read(m.config.uvs232URL, &r); err != nil {
		return
	}

	m.Lock()
	defer m.Unlock()

	m.BrineFlow = r.Measurand.Temperature4
	m.BrineReturn = r.Measurand.Temperature3
	return
}

func (m *Measurements) readMeter() (err error) {
	var r meterURLBody

	if err = read(m.config.meterURL, &r); err != nil {
		return
	}

	m.Lock()
	defer m.Unlock()

	m.Power = r.Measurand.P
	return
}

func read(url string, data interface{}) (err error) {
	done := make(chan bool, 1)
	go func() {
		// ensures that data is sent to the channel when the function is terminated
		defer func() {
			select {
			case done <- true:
			default:
			}
			close(done)
		}()

		debug.TraceLog.Printf("performing http get: %v\n", url)

		var resp *http.Response
		if resp, err = http.Get(url); err != nil {
			return
		}

		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()

		if err = json.Unmarshal(bodyBytes, data); err != nil {
			return
		}
	}()

	// wait for API Data
	select {
	case <-done:
	case <-time.After(httpRequestTimeout):
		err = errors.New("timeout during receive data")
	}

	if err != nil {
		debug.ErrorLog.Println(err)
		return
	}
	return
}
