package main

import (
	"io"
	"os"
	"strconv"
	"time"
)

func main() {
	// Parse environment variable for speed limit
	speedLimitMBPerSecond := 100 // Default value
	if speedLimitStr := os.Getenv("SPEED_LIMIT_MB_PER_SECOND"); speedLimitStr != "" {
		speedLimit, err := strconv.Atoi(speedLimitStr)
		if err == nil && speedLimit > 0 {
			speedLimitMBPerSecond = speedLimit
		}
	}

	// Set up buffer
	bufferSize := 1024
	buffer := make([]byte, bufferSize)

	for {
		// Read from stdin respecting the speed limit
		startTime := time.Now()
		bytesRead, err := os.Stdin.Read(buffer)

		if err != nil {
			if err != io.EOF {
				panic(err)
			}
			break // Exit loop on EOF
		}

		// Calculate the time taken to read the data
		elapsedTime := time.Since(startTime)
		dataSizeMB := float64(bytesRead) / (1024 * 1024)

		// Calculate the desired time to read the data based on the speed limit
		desiredTimeSeconds := dataSizeMB / float64(speedLimitMBPerSecond)

		// Sleep if the actual read time is less than the desired time
		if elapsedTime.Seconds() < desiredTimeSeconds {
			sleepDuration := time.Duration(desiredTimeSeconds*float64(time.Second)) - elapsedTime
			// compensate the 40% we are missing
			sleepDuration = sleepDuration - sleepDuration*40/100
			time.Sleep(sleepDuration)
		}

		// Write to stdout
		_, err = os.Stdout.Write(buffer[:bytesRead])
		if err != nil {
			panic(err)
		}
	}
}
