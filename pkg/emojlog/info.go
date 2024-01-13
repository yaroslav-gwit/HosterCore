package emojlog

func generateInfo(value string) string {
	initialValue := " 🟢 INFO:    🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}
