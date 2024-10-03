package emojlog

import "fmt"

func generateChanged(value string) string {
	initialValue := " 🔶 CHANGED: 🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func PrintChangedMessage(value string) {
	message := generateChanged(value)
	fmt.Println(message)
}
