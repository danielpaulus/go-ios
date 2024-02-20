package ncm

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	deviceCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "device_count",
		Help: "How many iOS devices are connected",
	})
	networkReceiveBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "network_receive_bytes",
		Help: "Counter metric for received bytes on the virtual TAP device",
	}, []string{"device", "serial"})
	networkSendBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "network_send_bytes",
		Help: "Counter metric for bytes sent to the virtual TAP device",
	}, []string{"device", "serial"})
	usbSendBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "usb_send_bytes",
		Help: "Counter metric for send bytes on the usb endpoint",
	}, []string{"serial"})
	usbReceiveBytes = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "usb_receive_bytes",
		Help: "Counter metric for received bytes on the USB endpoint",
	}, []string{"serial"})
)

func StartPrometheus(port int) {
	prometheus.MustRegister(deviceCount)
	prometheus.MustRegister(networkReceiveBytes)
	prometheus.MustRegister(networkSendBytes)
	prometheus.MustRegister(usbSendBytes)
	prometheus.MustRegister(usbReceiveBytes)

	// Expose metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Start HTTP server
	slog.Info("prometheus metrics up", "port", port, "endpoint", "/metrics")
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		slog.Error("failed starting metrics endpoint", "err", err)
	}
}
