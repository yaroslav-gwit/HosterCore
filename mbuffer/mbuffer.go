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
	if len(speedLimitStr) > 0 {
		speedLimit, err := strconv.Atoi(speedLimitStr)
		if err == nil && speedLimit > 0 {
			speedLimitMBPerSecond = speedLimit
		}
	}

	// Step 2: Set up buffer and adjust initial buffer size
	bufferSize := 1024 // Initial buffer size
	buffer := make([]byte, bufferSize)

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

		// If the actual speed exceeds the limit, adjust the buffer size
		if actualSpeedMBPerSecond > float64(speedLimitMBPerSecond) {
			newBufferSize := int(float64(bufferSize) * (float64(speedLimitMBPerSecond) / actualSpeedMBPerSecond))
			if newBufferSize > cap(buffer) {
				newBuffer := make([]byte, newBufferSize)
				copy(newBuffer, buffer)
				buffer = newBuffer
			}
			bufferSize = newBufferSize
		}

		// Step 4: Write to stdout
		_, err = os.Stdout.Write(buffer[:bytesRead])
		if err != nil {
			panic(err)
		}
	}

	// Step 5: Finish writing and exit
}
