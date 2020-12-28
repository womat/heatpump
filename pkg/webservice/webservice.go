package webservice

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"heatpump/global"
)

func init() {
	InitWebService()
}

func InitWebService() (err error) {
	for pattern, f := range map[string]func(http.ResponseWriter, *http.Request){
		"version":     httpGetVersion,
		"currentdata": httpReadCurrentData,
	} {
		if set, ok := global.Config.Webserver.Webservices[pattern]; ok && set {
			http.HandleFunc("/"+pattern, f)
		}
	}

	port := ":" + strconv.Itoa(global.Config.Webserver.Port)
	go http.ListenAndServe(port, nil)
	return
}

// httpGetVersion prints the SW Version
func httpGetVersion(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(global.VERSION)); err != nil {
		errorLog.Println(err)
		return
	}
}

// httpReadCurrentData supplies the data of al meters
func httpReadCurrentData(w http.ResponseWriter, r *http.Request) {
	const connectionString = "http://raspberrypi:4000/currentdata"

	var err error
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

		debugLog.Printf("performing http get: %v\n", connectionString)

		var resp *http.Response
		if resp, err = http.Get(connectionString); err != nil {
			return
		}

		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()

		// Convert response body to result struct
		type body struct {
			Timestamp time.Time
			Runtime   int
			Data      struct {
				Temperature1, Temperature2, Temperature3, Temperature4 float64
				Out1, Out2                                             bool
				RotationSpeed                                          float64
			}
		}

		var bodyStruct body
		if err = json.Unmarshal(bodyBytes, &bodyStruct); err != nil {
			return
		}
		traceLog.Printf("api response: %+v\n", bodyStruct)
		global.Measurements.Lock()
		defer global.Measurements.Unlock()

		global.Measurements.TimeStamp = time.Now()
		global.Measurements.BrineFlow = bodyStruct.Data.Temperature4
		global.Measurements.BrineReturn = bodyStruct.Data.Temperature3
	}()

	// wait for API Data
	select {
	case <-done:
	case <-time.After(time.Second * 10):
		err = errors.New("timeout during receive data")
	}

	if err != nil {
		errorLog.Println(err)
		return
	}

	var j []byte
	func() {
		global.Measurements.RLock()
		defer global.Measurements.RUnlock()
		j, err = json.MarshalIndent(&global.Measurements, "", "  ")
	}()

	if err != nil {
		errorLog.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if _, err = w.Write(j); err != nil {
		errorLog.Println(err)
		return
	}
}
