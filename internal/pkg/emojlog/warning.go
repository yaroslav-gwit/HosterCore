package emojlog

import "fmt"

func generateWarning(value string) string {
	initialValue := " 🔴 WARNING: 🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func PrintWarningMessage(value string) {
	message := generateWarning(value)
	fmt.Println(message)
}
