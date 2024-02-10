package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type ShellyplugResponse struct {
	//Ble struct {
	//} `json:"ble"`
	//Cloud struct {
	//	Connected bool `json:"connected"`
	//} `json:"cloud"`
	//Mqtt struct {
	//	Connected bool `json:"connected"`
	//} `json:"mqtt"`
	//PlugsUI struct {
	//} `json:"plugs_ui"`
	Switch0 struct {
		//ID      int     `json:"id"`
		//Source  string  `json:"source"`
		Output  bool    `json:"output"`
		Apower  float64 `json:"apower"`
		Voltage float64 `json:"voltage"`
		Current float64 `json:"current"`
		Aenergy struct {
			Total float64 `json:"total"`
			//ByMinute []float64 `json:"by_minute"`
			//MinuteTs int       `json:"minute_ts"`
		} `json:"aenergy"`
		Temperature struct {
			TC float64 `json:"tC"`
			//TF float64 `json:"tF"`
		} `json:"temperature"`
	} `json:"switch:0"`
	Sys struct {
		Mac string `json:"mac"`
		//RestartRequired  bool   `json:"restart_required"`
		//Time             string `json:"time"`
		//Unixtime         int    `json:"unixtime"`
		//Uptime           int    `json:"uptime"`
		//RAMSize          int    `json:"ram_size"`
		//RAMFree          int    `json:"ram_free"`
		//FsSize           int    `json:"fs_size"`
		//FsFree           int    `json:"fs_free"`
		//CfgRev           int    `json:"cfg_rev"`
		//KvsRev           int    `json:"kvs_rev"`
		//ScheduleRev      int    `json:"schedule_rev"`
		//WebhookRev       int    `json:"webhook_rev"`
		AvailableUpdates struct {
			Stable struct {
				Version string `json:"version"`
			} `json:"stable"`
		} `json:"available_updates"`
		//ResetReason int `json:"reset_reason"`
	} `json:"sys"`
	//Wifi struct {
	//	StaIP  string `json:"sta_ip"`
	//	Status string `json:"status"`
	//	Ssid   string `json:"ssid"`
	//	Rssi   int    `json:"rssi"`
	//} `json:"wifi"`
	//Ws struct {
	//	Connected bool `json:"connected"`
	//} `json:"ws"`
}

func getMetrics(url string) *ShellyplugResponse {
	log.Println("Getting status using url " + url + "/rpc/Shelly.GetStatus")
	res, err := http.Get(url + "/rpc/Shelly.GetStatus")
	if err != nil {
		log.Println("Error getting Shelly Plug status")
		return nil
	}
	if res.StatusCode > 299 {
		log.Printf("Response failed with status code: %d", res.StatusCode)
		return nil
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	var result ShellyplugResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		log.Println("Can not unmarshal JSON")
		return nil
	}
	return &result
}

func recordMetrics() {
	go func() {
		url := os.Getenv("SHELLYPLUG_URL")
		if url == "" {
			log.Fatal("SHELLYPLUG_URL is empty")
		}
		for {
			result := getMetrics(url)
			if result == nil {
				continue
			}
			aPower.WithLabelValues(result.Sys.Mac).Set(result.Switch0.Apower)
			voltageShelly.WithLabelValues(result.Sys.Mac).Set(result.Switch0.Voltage)
			currentShelly.WithLabelValues(result.Sys.Mac).Set(result.Switch0.Current)
			aEnergyTotalShelly.WithLabelValues(result.Sys.Mac).Set(result.Switch0.Aenergy.Total)
			temperatureShelly.WithLabelValues(result.Sys.Mac).Set(result.Switch0.Temperature.TC)
			if result.Switch0.Output {
				outputShelly.WithLabelValues(result.Sys.Mac).Set(1)
			} else {
				outputShelly.WithLabelValues(result.Sys.Mac).Set(0)
			}
			availableUpdatesShelly.Reset()
			if result.Sys.AvailableUpdates.Stable.Version != "" {
				availableUpdatesShelly.WithLabelValues(result.Sys.Mac, result.Sys.AvailableUpdates.Stable.Version).Set(1)
			} else {
				availableUpdatesShelly.WithLabelValues(result.Sys.Mac, "current").Set(1)
			}
			lastUpdatedShelly.WithLabelValues(result.Sys.Mac).Set(float64(time.Now().Unix()))
			time.Sleep(60 * time.Second)
		}
	}()
}

var (
	aPower = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_apower",
		Help: "Instantaneous power in W",
	}, []string{"mac"})
	voltageShelly = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_voltage",
		Help: "Voltage in V",
	}, []string{"mac"})
	currentShelly = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_current",
		Help: "Current in A",
	}, []string{"mac"})
	aEnergyTotalShelly = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_aenergy_total",
		Help: "Total energy so far in Wh",
	}, []string{"mac"})
	temperatureShelly = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_temperature",
		Help: "Temperature of Shellyplug in °C",
	}, []string{"mac"})
	outputShelly = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_output",
		Help: "true if output channel is currently on, false otherwise",
	}, []string{"mac"})

	availableUpdatesShelly = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_available_updates_info",
		Help: "Information about available updates",
	}, []string{"mac", "version"})

	lastUpdatedShelly = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "shellyplug_last_updated",
		Help: "Last update of Shellyplug",
	}, []string{"mac"})
)

func main() {
	log.Println("Starting server")
	recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":2112", nil))
}
