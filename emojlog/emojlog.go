package emojlog

import (
	"fmt"
	"time"
)

const Info string = "info"
const Changed string = "changed"
const Debug string = "debug"
const Warning string = "warning"
const Error string = "error"

// Prints out log messages to the screen, includes status emoji, time and value.
//
// > Example: `emojlog.PrintLogMessage("message", emojlog.Changed)`
//
// > Result `🔶 CHANGED: 🕔 2023-02-23 22:05:58: 📄 message`.
func PrintLogMessage(value string, msgType string) {
	var result string
	switch msgType {
	case Info:
		result = generateInfo(value)
	case Changed:
		result = generateChanged(value)
	case Debug:
		result = generateDebug(value)
	case Warning:
		result = generateWarning(value)
	case Error:
		result = generateError(value)
	default:
		result = ""
	}
	fmt.Println(result)
}

func generateInfo(value string) string {
	initialValue := " 🟢 INFO:    🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func generateChanged(value string) string {
	initialValue := " 🔶 CHANGED: 🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func generateDebug(value string) string {
	initialValue := " 🔷 DEBUG:   🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func generateWarning(value string) string {
	initialValue := " 🔴 WARNING: 🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func generateError(value string) string {
	initialValue := " 🚫 ERROR:   🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func generateTime() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
