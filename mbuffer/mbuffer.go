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
	speedLimitStr := os.Getenv("SPEED_LIMIT_MBS")
	if speedLimitStr != "" {
		speedLimit, err := strconv.Atoi(speedLimitStr)
		if err == nil && speedLimit > 0 {
			speedLimitMBPerSecond = speedLimit
		}
	}

	// Step 2: Set up buffer
	bufferSize := 1024 // Adjust the buffer size as per your requirements
	buffer := make([]byte, bufferSize)

	// Step 3: Initialize the limiter
	limiter := time.Tick(time.Second)

	for {
		// Step 3: Read from stdin respecting the speed limit
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
		dataSizeMB := float64(bytesRead) // (1024 * 1024)
		durationSeconds := elapsedTime.Seconds()

		// Step 4: Wait for the limiter to control the speed
		for range limiter {
			// Calculate the actual speed of data transfer
			actualSpeedMBPerSecond := dataSizeMB / durationSeconds

			// If the actual speed exceeds the limit, adjust the limiter to control the speed
			if actualSpeedMBPerSecond > float64(speedLimitMBPerSecond) {
				limiter = time.Tick(time.Second / time.Duration(speedLimitMBPerSecond))
			}
			break
		}

		// Step 5: Write to stdout
		_, err = os.Stdout.Write(buffer[:bytesRead])
		if err != nil {
			panic(err)
		}
	}

	// Step 6: Finish writing and exit
}
