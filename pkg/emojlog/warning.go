package emojlog

func generateWarning(value string) string {
	initialValue := " 🔴 WARNING: 🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}
