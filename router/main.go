package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Device struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Metrics struct {
	devices prometheus.Gauge
	info    *prometheus.GaugeVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		devices: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: "router",
			Name:      "devices",
			Help:      "Number of devices",
		}),
		info: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "router",
			Name:      "version",
			Help:      "Version of the router",
		}, []string{"version"}),
	}
	reg.MustRegister(m.devices, m.info)
	return m
}

var devices []Device
var version string

func init() {
	devices = []Device{
		{ID: 1, Name: "Device 1"},
		{ID: 2, Name: "Device 2"},
		{ID: 3, Name: "Device 3"},
	}
	version = "1.0"
}

func main() {
	reg := prometheus.NewRegistry()
	metrics := NewMetrics(reg)
	metrics.devices.Set(float64(len(devices)))
	metrics.info.With(prometheus.Labels{"version": version}).Set(1)

	mainMux := http.NewServeMux()
	rdh := RegisterDevicesHandler{metrics: metrics}
	mainMux.Handle("/devices", rdh)

	promMux := http.NewServeMux()
	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	promMux.Handle("/metrics", promHandler)

	go func() {
		log.Fatal(http.ListenAndServe(":8080", mainMux))
	}()

	go func() {
		log.Fatal(http.ListenAndServe(":8081", promMux))
	}()

	select {}
}

type RegisterDevicesHandler struct {
	metrics *Metrics
}

func (rdh RegisterDevicesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getDevices(w, r)
	case "POST":
		createDevice(w, r, rdh.metrics)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getDevices(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(devices)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(b)
}

func createDevice(w http.ResponseWriter, r *http.Request, m *Metrics) {
	var d Device

	err := json.NewDecoder(r.Body).Decode(&d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	devices = append(devices, d)
	m.devices.Set(float64(len(devices)))

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Device Created!"))
}
