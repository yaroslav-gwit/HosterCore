package main

import (
	"io"
	"os"
	"strconv"
	"time"
)

func main() {
	// Step 1: Parse environment variable for speed limit
	speedLimitMBPerSecond := 100 // Default value
	if speedLimitStr := os.Getenv("SPEED_LIMIT_MBS"); speedLimitStr != "" {
		speedLimit, err := strconv.Atoi(speedLimitStr)
		if err == nil && speedLimit > 0 {
			speedLimitMBPerSecond = speedLimit
		}
	}

	// Step 2: Set up buffer
	bufferSize := 1024 // Adjust the buffer size as per your requirements
	buffer := make([]byte, bufferSize)

	for {
		// Step 4: Read from stdin respecting the speed limit
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
		durationSeconds := elapsedTime.Seconds()

		// Calculate the actual speed of data transfer
		actualSpeedMBPerSecond := dataSizeMB / durationSeconds

		// If the actual speed exceeds the limit, wait to control the speed
		if actualSpeedMBPerSecond > float64(speedLimitMBPerSecond) {
			sleepTime := time.Duration((dataSizeMB / float64(speedLimitMBPerSecond)) * float64(time.Second))
			time.Sleep(sleepTime)
		}

		// Step 5: Write to stdout
		_, err = os.Stdout.Write(buffer[:bytesRead])
		if err != nil {
			panic(err)
		}
	}

	// Step 6: Finish writing and exit
}
