package osfreebsd

import (
	"fmt"
)

func BytesToHuman(bytes uint64) string {
	// SET TO KB
	var firstIteration = bytes / 1024
	var iterationTitle = "K"

	// SET TO MB
	if firstIteration > 1024 {
		firstIteration = firstIteration / 1024
		iterationTitle = "M"
	}

	// SET TO GB
	var firstIterationFloat = 0.0
	if firstIteration > 1024 {
		firstIterationFloat = float64(firstIteration) / 1024.0
		iterationTitle = "G"
	}

	// FORMAT THE OUTPUT
	var result string
	if firstIterationFloat > 0.0 {
		result = fmt.Sprintf("%.2f%s", firstIterationFloat, iterationTitle)
	} else {
		result = fmt.Sprintf("%d%s", firstIteration, iterationTitle)
	}

	return result
}
