package emojlog

func generateChanged(value string) string {
	initialValue := " 🔶 CHANGED: 🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}
