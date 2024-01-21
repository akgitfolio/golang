package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/shirou/gopsutil/process"
)

func main() {
	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer termui.Close()

	// Create UI widgets
	cpuGauge := widgets.NewGauge()
	cpuGauge.Title = "CPU Usage"
	cpuGauge.Percent = 0
	cpuGauge.SetRect(0, 0, 50, 3)
	cpuGauge.BarColor = termui.ColorRed
	cpuGauge.LabelStyle = termui.NewStyle(termui.ColorWhite)

	memGauge := widgets.NewGauge()
	memGauge.Title = "Memory Usage"
	memGauge.Percent = 0
	memGauge.SetRect(0, 4, 50, 7)
	memGauge.BarColor = termui.ColorGreen
	memGauge.LabelStyle = termui.NewStyle(termui.ColorWhite)

	diskGauge := widgets.NewGauge()
	diskGauge.Title = "Disk Usage"
	diskGauge.Percent = 0
	diskGauge.SetRect(0, 8, 50, 11)
	diskGauge.BarColor = termui.ColorYellow
	diskGauge.LabelStyle = termui.NewStyle(termui.ColorWhite)

	netTable := widgets.NewTable()
	netTable.Title = "Network Traffic"
	netTable.Rows = [][]string{
		{"Interface", "Bytes Sent", "Bytes Received"},
	}
	netTable.SetRect(0, 12, 50, 20)
	netTable.TextStyle = termui.NewStyle(termui.ColorWhite)
	netTable.RowSeparator = true

	procTable := widgets.NewTable()
	procTable.Title = "Running Processes"
	procTable.Rows = [][]string{
		{"PID", "Name", "CPU (%)", "Memory (%)"},
	}
	procTable.SetRect(0, 21, 100, 40)
	procTable.TextStyle = termui.NewStyle(termui.ColorWhite)
	procTable.RowSeparator = true

	// Render the UI
	termui.Render(cpuGauge, memGauge, diskGauge, netTable, procTable)

	uiEvents := termui.PollEvents()
	ticker := time.NewTicker(time.Second).C

	for {
		select {
		case <-ticker:
			// Update CPU usage
			cpuPercent, _ := cpu.Percent(0, false)
			if len(cpuPercent) > 0 {
				cpuGauge.Percent = int(cpuPercent[0])
			}

			// Update memory usage
			memInfo, _ := mem.VirtualMemory()
			memGauge.Percent = int(memInfo.UsedPercent)

			// Update disk usage
			diskInfo, _ := disk.Usage("/")
			diskGauge.Percent = int(diskInfo.UsedPercent)

			// Update network traffic
			netInfo, _ := net.IOCounters(true)
			netTable.Rows = [][]string{
				{"Interface", "Bytes Sent", "Bytes Received"},
			}
			for _, netStat := range netInfo {
				netTable.Rows = append(netTable.Rows, []string{
					netStat.Name,
					fmt.Sprintf("%v", netStat.BytesSent),
					fmt.Sprintf("%v", netStat.BytesRecv),
				})
			}

			// Update running processes
			processes, _ := process.Processes()
			procTable.Rows = [][]string{
				{"PID", "Name", "CPU (%)", "Memory (%)"},
			}
			for _, proc := range processes {
				name, _ := proc.Name()
				cpuPercent, _ := proc.CPUPercent()
				memPercent, _ := proc.MemoryPercent()
				procTable.Rows = append(procTable.Rows, []string{
					fmt.Sprintf("%d", proc.Pid),
					name,
					fmt.Sprintf("%.2f", cpuPercent),
					fmt.Sprintf("%.2f", memPercent),
				})
			}

			termui.Render(cpuGauge, memGauge, diskGauge, netTable, procTable)

		case e := <-uiEvents:
			if e.Type == termui.KeyboardEvent {
				switch e.ID {
				case "q", "<C-c>":
					return
				}
			}
		}
	}
}
