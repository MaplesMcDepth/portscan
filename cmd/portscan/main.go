package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var commonPorts = []int{
	21, 22, 23, 25, 53, 80, 110, 143, 443, 465, 587, 993, 995,
	3000, 3306, 3389, 5432, 5900, 8000, 8080, 8443, 9000, 9200,
}

func usage() {
	fmt.Print(`portscan — Fast TCP port scanner

Usage: portscan [options] <host>

Options:
  -p string    Ports to scan (e.g. "80,443" or "1-1000") (default "common")
  -t int       Timeout in ms (default 1000)
  -c int       Concurrency (default 100)
  -j           JSON output
  -v           Verbose (show closed ports too)
  -open        Show only open ports

Examples:
  portscan localhost                  # Scan common ports
  portscan -p 1-1000 192.168.1.1      # Scan range
  portscan -p 22,80,443 example.com   # Scan specific ports
  portscan -j -p 1-100 localhost      # JSON output
`)
}

type Result struct {
	Port    int    `json:"port"`
	Open    bool   `json:"open"`
	Service string `json:"service,omitempty"`
}

func main() {
	var (
		portsFlag = flag.String("p", "common", "Ports: common, all, or range (e.g. 1-1000, 80,443)")
		timeout   = flag.Int("t", 1000, "Timeout in milliseconds")
		workers   = flag.Int("c", 100, "Concurrency")
		jsonOut   = flag.Bool("j", false, "JSON output")
		verbose   = flag.Bool("v", false, "Verbose")
		onlyOpen  = flag.Bool("open", false, "Only show open ports")
	)
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	host := flag.Arg(0)

	ports := parsePorts(*portsFlag)
	if len(ports) == 0 {
		fmt.Fprintln(os.Stderr, "No ports to scan")
		os.Exit(1)
	}

	results := scan(host, ports, time.Duration(*timeout)*time.Millisecond, *workers)

	if *jsonOut {
		printJSON(results)
	} else {
		printTable(results, *verbose || *onlyOpen)
	}
}

func parsePorts(flag string) []int {
	switch flag {
	case "common":
		return commonPorts
	case "all":
		var ports []int
		for i := 1; i <= 65535; i++ {
			ports = append(ports, i)
		}
		return ports
	}

	var ports []int
	parts := strings.Split(flag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) == 2 {
				start, _ := strconv.Atoi(rangeParts[0])
				end, _ := strconv.Atoi(rangeParts[1])
				for i := start; i <= end; i++ {
					if i > 0 && i <= 65535 {
						ports = append(ports, i)
					}
				}
			}
		} else {
			port, err := strconv.Atoi(part)
			if err == nil && port > 0 && port <= 65535 {
				ports = append(ports, port)
			}
		}
	}
	return ports
}

func scan(host string, ports []int, timeout time.Duration, workers int) []Result {
	var wg sync.WaitGroup
	portChan := make(chan int, workers)
	resultsChan := make(chan Result, len(ports))

	// Workers
	for i := 0; i < workers; i++ {
		go func() {
			for port := range portChan {
				open := checkPort(host, port, timeout)
				resultsChan <- Result{
					Port:    port,
					Open:    open,
					Service: guessService(port),
				}
				wg.Done()
			}
		}()
	}

	// Send work
	for _, port := range ports {
		wg.Add(1)
		portChan <- port
	}

	close(portChan)
	wg.Wait()
	close(resultsChan)

	// Collect results
	var results []Result
	for r := range resultsChan {
		results = append(results, r)
	}

	// Sort by port
	sort.Slice(results, func(i, j int) bool {
		return results[i].Port < results[j].Port
	})

	return results
}

func checkPort(host string, port int, timeout time.Duration) bool {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func guessService(port int) string {
	services := map[int]string{
		21: "ftp", 22: "ssh", 23: "telnet", 25: "smtp",
		53: "dns", 80: "http", 110: "pop3", 143: "imap",
		443: "https", 465: "smtps", 587: "submission",
		993: "imaps", 995: "pop3s", 3000: "http-alt",
		3306: "mysql", 3389: "rdp", 5432: "postgresql",
		5900: "vnc", 8000: "http-alt", 8080: "http-proxy",
		8443: "https-alt", 9000: "http-alt", 9200: "elasticsearch",
	}
	if s, ok := services[port]; ok {
		return s
	}
	return ""
}

func printTable(results []Result, onlyOpen bool) {
	fmt.Printf("%-6s %-8s %s\n", "PORT", "STATE", "SERVICE")
	fmt.Println(strings.Repeat("-", 30))
	openCount := 0
	for _, r := range results {
		if !r.Open && onlyOpen {
			continue
		}
		state := "closed"
		if r.Open {
			state = "open"
			openCount++
		}
		fmt.Printf("%-6d %-8s %s\n", r.Port, state, r.Service)
	}
	fmt.Printf("\n%d/%d ports open\n", openCount, len(results))
}

func printJSON(results []Result) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(results)
}
