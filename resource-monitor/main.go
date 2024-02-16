package main

import (
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type SystemResourceMonitor struct {
	cpuThreshold     float64
	memoryThreshold  float64
	networkThreshold float64

	cpuUsage     float64
	memoryUsage  float64
	networkUsage float64

	threadPool      *sync.WaitGroup
	gcTrigger       chan bool
	networkThrottle chan bool

	mu sync.Mutex
}

func NewSystemResourceMonitor(cpuThreshold, memoryThreshold, networkThreshold float64) *SystemResourceMonitor {
	return &SystemResourceMonitor{
		cpuThreshold:     cpuThreshold,
		memoryThreshold:  memoryThreshold,
		networkThreshold: networkThreshold,
		threadPool:       &sync.WaitGroup{},
		gcTrigger:        make(chan bool, 1),
		networkThrottle:  make(chan bool, 1),
	}
}

func (m *SystemResourceMonitor) Start() {
	go func() {
		var prevNetIO []net.IOCountersStat
		for {
			m.collectResourceUsage(&prevNetIO)

			if m.cpuUsage > m.cpuThreshold {
				m.scaleThreadPool(true)
			} else if m.cpuUsage < m.cpuThreshold/2 {
				m.scaleThreadPool(false)
			}

			if m.memoryUsage > m.memoryThreshold {
				m.gcTrigger <- true
			}

			if m.networkUsage > m.networkThreshold {
				m.networkThrottle <- true
			}

			time.Sleep(1 * time.Second)
		}
	}()

	go m.handleGC()
}

func (m *SystemResourceMonitor) collectResourceUsage(prevNetIO *[]net.IOCountersStat) {
	var err error

	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		fmt.Println("Error getting CPU usage:", err)
		return
	}
	m.mu.Lock()
	m.cpuUsage = cpuPercent[0]
	m.mu.Unlock()

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Error getting memory usage:", err)
		return
	}
	m.mu.Lock()
	m.memoryUsage = memInfo.UsedPercent
	m.mu.Unlock()

	netIO, err := net.IOCounters(true)
	if err != nil {
		fmt.Println("Error getting network usage:", err)
		return
	}
	if len(*prevNetIO) > 0 {
		m.mu.Lock()
		m.networkUsage = float64((netIO[0].BytesSent+netIO[0].BytesRecv)-((*prevNetIO)[0].BytesSent+(*prevNetIO)[0].BytesRecv)) / float64(time.Second)
		m.mu.Unlock()
	}
	*prevNetIO = netIO
}

func (m *SystemResourceMonitor) scaleThreadPool(increase bool) {
	if increase {
		for i := 0; i < runtime.NumCPU()/2; i++ {
			m.threadPool.Add(1)
			go func() {
				time.Sleep(100 * time.Millisecond)
				m.threadPool.Done()
			}()
		}
	} else {
		m.threadPool.Wait()
		runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	}
}

func (m *SystemResourceMonitor) handleGC() {
	for range m.gcTrigger {
		runtime.GC()
		fmt.Println("Garbage collection triggered")
	}
}

func (m *SystemResourceMonitor) HandleRequest(w http.ResponseWriter, r *http.Request) {
	select {
	case <-m.networkThrottle:
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(w, "Network is busy, please try again later.")
	default:
		fmt.Fprintf(w, "Request handled successfully.")
	}
}

func main() {
	monitor := NewSystemResourceMonitor(80, 90, 1000000)
	monitor.Start()

	http.HandleFunc("/", monitor.HandleRequest)
	fmt.Println("Server listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
