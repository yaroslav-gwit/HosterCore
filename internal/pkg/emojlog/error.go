package emojlog

import "fmt"

func generateError(value string) string {
	initialValue := " 🚫 ERROR:   🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}

func PrintErrorMessage(value string) {
	message := generateError(value)
	fmt.Println(message)
}
