package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rodaine/table"
)

var (
	timeout  time.Duration = 1 * time.Second
	protocol string        = "tcp"
)

type ScanResult struct {
	IP   string
	Port int
	Open bool
	Err  error
}

func main() {
	network := flag.String("network", "192.168.1.0/24", "Network range to scan in CIDR notation")
	flag.Parse()

	ips, err := generateIPList(*network)
	if err != nil {
		log.Fatalf("Failed to generate IP list: %v", err)
	}

	var wg sync.WaitGroup
	results := make(chan ScanResult, len(ips)*1024)

	for _, ip := range ips {
		for port := 1; port <= 1024; port++ {
			wg.Add(1)
			go func(ip string, port int) {
				defer wg.Done()
				result := ScanResult{IP: ip, Port: port}
				conn, err := net.DialTimeout(protocol, fmt.Sprintf("%s:%d", ip, port), timeout)
				if err != nil {
					result.Open = false
					result.Err = err
				} else {
					result.Open = true
					conn.Close()
				}
				results <- result
			}(ip, port)
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	processResults(results)
}

func generateIPList(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	return ips[1 : len(ips)-1], nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func processResults(results chan ScanResult) {
	openPorts := make(map[string][]int)
	for result := range results {
		if result.Open {
			openPorts[result.IP] = append(openPorts[result.IP], result.Port)
		}
	}

	// Sort IPs
	ips := make([]string, 0, len(openPorts))
	for ip := range openPorts {
		ips = append(ips, ip)
	}
	sort.Strings(ips)

	// Display results in a table format
	tbl := table.New("IP Address", "Open Ports")
	for _, ip := range ips {
		ports := openPorts[ip]
		sort.Ints(ports)
		tbl.AddRow(ip, strings.Trim(strings.Replace(fmt.Sprint(ports), " ", ", ", -1), "[]"))
	}
	tbl.Print()
}
