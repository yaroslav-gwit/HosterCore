package main

import (
	"io"
	"os"
	"strconv"
	"time"
)

func main() {
	// Parse environment variable for speed limit
	speedLimitMBPerSecond := 100 // Set the default limit value, 100MB in our case
	speedLimitStr := os.Getenv("SPEED_LIMIT_MBS")
	if len(speedLimitStr) > 0 {
		speedLimit, err := strconv.Atoi(speedLimitStr)
		if err == nil && speedLimit > 0 {
			speedLimitMBPerSecond = speedLimit
		}
	}

	// Set up buffer
	bufferSize := 1024 * 1024 // Adjust the buffer size as per your requirements
	buffer := make([]byte, bufferSize)

	// Initialize the limiter
	limiter := time.Tick(time.Second)

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
		dataSizeMB := float64(bytesRead)
		durationSeconds := elapsedTime.Seconds()

		// Wait for the limiter to control the speed
		for range limiter {
			actualSpeedMBPerSecond := dataSizeMB / durationSeconds
			if actualSpeedMBPerSecond > float64(speedLimitMBPerSecond) {
				limiter = time.Tick(time.Second / time.Duration(speedLimitMBPerSecond) * 60 / 1024)
			}
			break
		}

		// Write to stdout
		_, err = os.Stdout.Write(buffer[0:bytesRead])
		if err != nil {
			panic(err)
		}
	}
	// Finish writing and exit
}
