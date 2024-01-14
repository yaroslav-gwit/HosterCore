package emojlog

func generateError(value string) string {
	initialValue := " 🚫 ERROR:   🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}
