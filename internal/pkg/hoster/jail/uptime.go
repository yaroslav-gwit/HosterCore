package HosterJail

import (
	timeconversion "HosterCore/internal/pkg/time_conversion"
	"os"
)

var stateFilePrefix = "/var/run/hoster_jail_state_"

// Function creates a state file for any given Jail name, that can be used later as an uptime indicator.
//
// It also removes the state file if it already exists, to avoid reporting a bad uptime to the end user.
func CreateUptimeStateFile(jailName string) error {
	_ = RemoveUptimeStateFile(jailName)
	_, err := os.Create(stateFilePrefix + jailName)
	if err != nil {
		return err
	}

	return nil
}

// Function removes a state file for any given Jail name, that can be used later as an uptime indicator.
func RemoveUptimeStateFile(jailName string) error {
	err := os.Remove(stateFilePrefix + jailName)
	if err != nil {
		return err
	}

	return nil
}

// Function returns a Jail uptime in this format: "8d 13h 55m 54s" or "0s".
//
// It uses the `mtime` of the Jail state file which was created/updated by the functions above.
func GetUptimeHuman(jailName string) (jailUptime string) {
	fileStat, err := os.Stat(stateFilePrefix + jailName)
	if err != nil {
		jailUptime = "0s"
		return
	}

	rawUptime := fileStat.ModTime().Unix()
	jailUptime = timeconversion.UnixTimeToUptime(rawUptime)
	return
}

// Function returns a Jail uptime using the unix time format: 3878912.
//
// It uses the `mtime` of the Jail state file which was created/updated by the functions above.
func GetUptimeRaw(jailName string) (jailUptime int64) {
	fileStat, err := os.Stat(stateFilePrefix + jailName)
	if err != nil {
		jailUptime = 0
		return
	}

	jailUptime = fileStat.ModTime().Unix()
	return
}
