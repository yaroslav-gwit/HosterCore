package emojlog

import "fmt"

func generateInfo(value string) string {
	initialValue := " 🟢 INFO:    🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func PrintInfoMessage(value string) {
	message := generateInfo(value)
	fmt.Println(message)
}
