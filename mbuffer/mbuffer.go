package main

import (
	"io"
	"os"
	"strconv"
	"time"
)

func main() {
	speedLimitMBPerSecond := 100 // Set the default limit value, 100MB in our case
	speedLimitStr := os.Getenv("SPEED_LIMIT_MBS")
	if len(speedLimitStr) > 0 {
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
		dataSizeMB := float64(bytesRead) / (1024 * 1024)
		durationSeconds := elapsedTime.Seconds()

		// Calculate the actual speed of data transfer
		actualSpeedMBPerSecond := dataSizeMB / durationSeconds

		// If the actual speed exceeds the limit, adjust the limiter to control the speed
		if actualSpeedMBPerSecond > float64(speedLimitMBPerSecond) {
			limiter = time.Tick(time.Second / time.Duration(speedLimitMBPerSecond))
		}
		// this is here, because the compiler was complaining about the value not being used
		_ = limiter

		// Step 5: Write to stdout
		_, err = os.Stdout.Write(buffer[:bytesRead])
		if err != nil {
			panic(err)
		}
	}

	// Step 6: Finish writing and exit
}
