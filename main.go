package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ANSI escape codes for text formatting
const (
	red   = "\033[91m"
	green = "\033[92m"
	bold  = "\033[1m"
	end   = "\033[0m"
)

func getServiceInfo(protocol, port string) string {
	file, err := os.Open("all.csv")
	if err != nil {
		log.Fatal("Error loading services:", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error reading all.csv:", err)
	}

	for _, record := range records {
		if strings.ToLower(record[0]) == strings.ToLower(protocol) && record[1] == port {
			return record[2]
		}
	}

	return "Unknown Service"
}

func scanPort(wg *sync.WaitGroup, protocol, host string, port int, logFile *os.File) {
	defer wg.Done()

	target := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial(protocol, target)
	if err != nil {
		return // Port is closed
	}
	defer conn.Close()

	service := getServiceInfo(protocol, strconv.Itoa(port))
	result := fmt.Sprintf("%s>> Port %d\t%s\t%s%s ", green, port, protocol, service, end)
	fmt.Println(result)
	logResult(logFile, result)
}

func logResult(logFile *os.File, result string) {
	if logFile != nil {
		log.SetOutput(logFile)
		log.Println(result)
		logFile.Sync() // Ensure the log file is written immediately
	}
}

func simplePortScanner(host string, startPort, endPort int, protocols []string, logFile *os.File) {
	var wg sync.WaitGroup

	for _, protocol := range protocols {
		for port := startPort; port <= endPort; port++ {
			wg.Add(1)
			go scanPort(&wg, protocol, host, port, logFile)
		}
	}

	wg.Wait()
}

func main() {
	for {
		targetHost := "example.com"
		startPort := 1
		endPort := 1024

		fmt.Printf("%s----------------------------------------%s\n", red, end)
		fmt.Printf("%s|                GOSCAN		       |%s\n", red, end)
		fmt.Printf("%s----------------------------------------%s\n", red, end)
		fmt.Printf("%s >> Written by Xnrrrrrr%s\n\n\n\n", red, end)

		fmt.Printf("%sEnter the target host (default: %s):%s ", bold, targetHost, end)
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input := scanner.Text()
			if input != "" {
				targetHost = input
			}
		}

		protocols := []string{"tcp", "udp"}

		fmt.Printf("%sDo you want to scan for TCP, UDP, or both? (tcp/udp/both - default: both):%s ", bold, end)
		if scanner.Scan() {
			protocolInput := strings.ToLower(scanner.Text())
			if protocolInput == "exit" {
				break
			}

			if protocolInput == "tcp" {
				protocols = []string{"tcp"}
			} else if protocolInput == "udp" {
				protocols = []string{"udp"}
			}

			logFileName := "logging.txt"
			logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal("Error opening log file:", err)
			}
			defer logFile.Close()

			// If the file doesn't exist, create it with the header
			fileStat, err := logFile.Stat()
			if err != nil || fileStat.Size() == 0 {
				header := "Scan Time: " + time.Now().Format("2006-01-02 15:04:05") + "\n"
				logResult(logFile, header)
			}

			fmt.Printf("\033[1;32m-------------------------------------------------------\033[0m\n")
			fmt.Printf("\033[1;32m|  PORT        OPEN     SERVICE                       |\033[0m\n")
			fmt.Printf("\033[1;32m-------------------------------------------------------\033[0m\n")

			fmt.Printf("\n%sScanning ports on %s from %d to %d using %s%s\n", bold, targetHost, startPort, endPort, protocols, end)
			simplePortScanner(targetHost, startPort, endPort, protocols, logFile)

			if protocolInput != "both" {
				fmt.Printf("%sDo you want to scan for the other protocol? (yes/no - default: no):%s ", bold, end)
				if scanner.Scan() {
					otherProtocolInput := strings.ToLower(scanner.Text())
					if otherProtocolInput == "yes" {
						otherProtocol := "udp"
						if protocolInput == "udp" {
							otherProtocol = "tcp"
						}

						protocols = []string{otherProtocol}
						logFileName = "logging.txt"
						logFile, err = os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
						if err != nil {
							log.Fatal("Error opening log file:", err)
						}
						defer logFile.Close()

						fmt.Printf("\n%sScanning ports on %s from %d to %d using %s%s\n", bold, targetHost, startPort, endPort, protocols, end)
						simplePortScanner(targetHost, startPort, endPort, protocols, logFile)
					}
				}
			}
		}
	}
}
