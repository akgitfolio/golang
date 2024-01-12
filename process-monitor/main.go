package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/olekukonko/tablewriter"
	"github.com/shirou/gopsutil/process"
)

func main() {
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		switch cmd {
		case "list":
			listProcesses()
		case "kill":
			if len(os.Args) != 3 {
				fmt.Println("Usage: ./process-monitor kill <pid>")
				return
			}
			pid, err := strconv.Atoi(os.Args[2])
			if err != nil {
				log.Fatalf("Invalid PID: %v", err)
			}
			killProcess(pid)
		case "signal":
			if len(os.Args) != 4 {
				fmt.Println("Usage: ./process-monitor signal <pid> <signal>")
				return
			}
			pid, err := strconv.Atoi(os.Args[2])
			if err != nil {
				log.Fatalf("Invalid PID: %v", err)
			}
			sig, err := strconv.Atoi(os.Args[3])
			if err != nil {
				log.Fatalf("Invalid signal: %v", err)
			}
			sendSignal(pid, syscall.Signal(sig))
		default:
			fmt.Println("Unknown command:", cmd)
			fmt.Println("Usage: ./process-monitor <list|kill|signal>")
		}
	} else {
		fmt.Println("Usage: ./process-monitor <list|kill|signal>")
	}
}

func listProcesses() {
	procs, err := process.Processes()
	if err != nil {
		log.Fatalf("Error retrieving processes: %v", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"PID", "Name", "CPU %", "Memory %"})
	for _, proc := range procs {
		name, _ := proc.Name()
		cpu, _ := proc.CPUPercent()
		mem, _ := proc.MemoryPercent()
		table.Append([]string{
			strconv.Itoa(int(proc.Pid)),
			name,
			fmt.Sprintf("%.2f", cpu),
			fmt.Sprintf("%.2f", mem),
		})
	}
	table.Render()
}

func killProcess(pid int) {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		log.Fatalf("Error finding process: %v", err)
	}
	err = proc.Kill()
	if err != nil {
		log.Fatalf("Error killing process: %v", err)
	}
	fmt.Printf("Process %d killed\n", pid)
}

func sendSignal(pid int, sig syscall.Signal) {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		log.Fatalf("Error finding process: %v", err)
	}
	err = proc.SendSignal(sig)
	if err != nil {
		log.Fatalf("Error sending signal to process: %v", err)
	}
	fmt.Printf("Signal %d sent to process %d\n", sig, pid)
}

func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range c {
			fmt.Printf("Caught signal %v: shutting down.\n", sig)
			os.Exit(0)
		}
	}()
}
