package emojlog

func generateDebug(value string) string {
	initialValue := " 🔷 DEBUG:   🕔 " + generateTime() + ": 📄 "
	return initialValue + value
}
