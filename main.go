package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL           = "http://srv.msk01.gigacorp.local/_stats"
	pollInterval        = 5 * time.Second
	loadAvgThreshold    = 30.0
	memUsageThreshold   = 0.8
	diskUsageThreshold  = 0.9
	netUsageThreshold   = 0.9
	maxConsecutiveFails = 3
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get(serverURL)
		if err != nil {
			errorCount++
			handleError(errorCount)
			time.Sleep(pollInterval)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			errorCount++
			handleError(errorCount)
			resp.Body.Close()
			time.Sleep(pollInterval)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			handleError(errorCount)
			time.Sleep(pollInterval)
			continue
		}

		data := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(data) != 7 {
			errorCount++
			handleError(errorCount)
			time.Sleep(pollInterval)
			continue
		}

		values := make([]float64, 7)
		valid := true
		for i, v := range data {
			values[i], err = strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil {
				valid = false
				break
			}
		}

		if !valid {
			errorCount++
			handleError(errorCount)
			time.Sleep(pollInterval)
			continue
		}

		// Сброс счётчика ошибок при успешном получении данных
		errorCount = 0

		loadAvg := values[0]
		memTotal := values[1]
		memUsed := values[2]
		diskTotal := values[3]
		diskUsed := values[4]
		netTotal := values[5]
		netUsed := values[6]

		// Проверка порогов
		if loadAvg > loadAvgThreshold {
			fmt.Printf("Load Average is too high: %.2f\n", loadAvg)
		}

		memUsage := memUsed / memTotal
		if memUsage > memUsageThreshold {
			fmt.Printf("Memory usage too high: %.2f%%\n", memUsage*100)
		}

		diskUsage := diskUsed / diskTotal
		if diskUsage > diskUsageThreshold {
			freeMb := (diskTotal - diskUsed) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeMb)
		}

		netUsage := netUsed / netTotal
		if netUsage > netUsageThreshold {
			freeMbit := (netTotal - netUsed) * 8 / (1024 * 1024)
			fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", freeMbit)
		}

		time.Sleep(pollInterval)
	}
}

func handleError(errorCount int) {
	if errorCount >= maxConsecutiveFails {
		fmt.Println("Unable to fetch server statistic.")
	}
}
