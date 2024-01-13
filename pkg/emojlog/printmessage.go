package emojlog

import (
	"fmt"
)

// Prints out log messages to the terminal. Includes status emoji, time and value.
//
// > Example: `emojlog.PrintLogMessage("your log message", emojlog.LEVEL_CHANGED)`
//
// > Result: `🔶 CHANGED: 🕔 2023-02-23 22:05:58: 📄 message`.
func PrintLogMessage(value string, msgType string) {
	var result string

	if msgType == Info || msgType == LEVEL_INFO {
		result = generateInfo(value)
	} else if msgType == Changed || msgType == LEVEL_CHANGED {
		result = generateChanged(value)
	} else if msgType == Debug || msgType == LEVEL_DEBUG {
		result = generateDebug(value)
	} else if msgType == Warning || msgType == LEVEL_WARNING {
		result = generateWarning(value)
	} else if msgType == Error || msgType == LEVEL_ERROR {
		result = generateError(value)
	}

	fmt.Println(result)
}

// Prints out log messages to the terminal. Includes status emoji, time and value.
//
// > Example: `emojlog.PrintLogMessage(emojlog.LEVEL_CHANGED, "your log message")`
//
// > Result: `🔶 CHANGED: 🕔 2023-02-23 22:05:58: 📄 message`.
func PrintLogLine(msgType string, value string) {
	var result string

	if msgType == Info || msgType == LEVEL_INFO {
		result = generateInfo(value)
	} else if msgType == Changed || msgType == LEVEL_CHANGED {
		result = generateChanged(value)
	} else if msgType == Debug || msgType == LEVEL_DEBUG {
		result = generateDebug(value)
	} else if msgType == Warning || msgType == LEVEL_WARNING {
		result = generateWarning(value)
	} else if msgType == Error || msgType == LEVEL_ERROR {
		result = generateError(value)
	}

	fmt.Println(result)
}
