package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

var (
	cpuUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "container_cpu_usage_seconds_total",
			Help: "Total CPU usage in seconds.",
		},
		[]string{"container_name"},
	)
	memUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "container_memory_usage_bytes",
			Help: "Memory usage in bytes.",
		},
		[]string{"container_name"},
	)
	netIO = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "container_network_io_bytes_total",
			Help: "Total network I/O in bytes.",
		},
		[]string{"container_name", "direction"},
	)
)

func init() {
	prometheus.MustRegister(cpuUsage)
	prometheus.MustRegister(memUsage)
	prometheus.MustRegister(netIO)
}

func recordMetrics() {
	for {
		// Simulate container name
		containerName := "example_container"

		// CPU usage
		cpuPercent, _ := cpu.Percent(time.Second, false)
		cpuUsage.WithLabelValues(containerName).Set(cpuPercent[0] / 100)

		// Memory usage
		v, _ := mem.VirtualMemory()
		memUsage.WithLabelValues(containerName).Set(float64(v.Used))

		// Network I/O
		ioCounters, _ := net.IOCounters(false)
		for _, io := range ioCounters {
			netIO.WithLabelValues(containerName, "rx").Set(float64(io.BytesRecv))
			netIO.WithLabelValues(containerName, "tx").Set(float64(io.BytesSent))
		}

		time.Sleep(15 * time.Second)
	}
}

func main() {
	go recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Addr: ":9323"}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on :9323: %v\n", err)
		}
	}()

	fmt.Println("Server is ready to handle requests at :9323")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	fmt.Println("Server is shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	fmt.Println("Server stopped")
}
