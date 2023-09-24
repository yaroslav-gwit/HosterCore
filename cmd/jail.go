package cmd

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	jailCmd = &cobra.Command{
		Use:   "jail",
		Short: "Jail related operations",
		Long:  `Jail related operations: deploy, stop, start, destroy, etc`,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				log.Fatal(err.Error())
			}
			cmd.Help()
		},
	}
)

type LiveJailStruct struct {
	ID         int
	Name       string
	Path       string
	Running    bool
	Ip4address string
	Ip6address string
}

func getRunningJails() ([]LiveJailStruct, error) {
	reSpaceSplit := regexp.MustCompile(`\s+`)
	jails := []LiveJailStruct{}

	out, err := exec.Command("jls", "-h", "jid", "name", "path", "dying", "ip4.addr", "ip6.addr").CombinedOutput()
	// Example output (in case we need to compare it if anything changes on the FreeBSD side in the future)
	// jid  name   path        dying   ip4.addr   ip6.addr
	// [0]  [1]     [2]         [3]       [4]      [5]
	// 1  example /root/jail   false  10.0.105.50   -
	// 2  twelve  /root/12_4   false  10.0.105.51   -
	// 3  twelve1 /root/12_4_1 false  10.0.105.52   -
	// 4  twelve2 /root/12_4_2 false  10.0.105.53   -
	// 5  twelve3 /root/12_4_3 false  10.0.105.54   -

	if err != nil {
		errorValue := "output: " + string(out) + " " + err.Error()
		return []LiveJailStruct{}, errors.New(errorValue)
	}

	for i, v := range strings.Split(string(out), "\n") {
		// Skip the header
		if i == 0 {
			continue
		}
		// Skip empty lines
		if len(v) < 1 {
			continue
		}

		tempList := reSpaceSplit.Split(strings.TrimSpace(v), -1)
		// In case we need to check the split output in the future
		// fmt.Println(tempList)

		tempStruct := LiveJailStruct{}

		jailId, err := strconv.Atoi(tempList[0])
		if err != nil {
			return []LiveJailStruct{}, err
		}

		tempStruct.ID = jailId
		tempStruct.Name = tempList[1]
		tempStruct.Path = tempList[2]

		tempStruct.Running, err = strconv.ParseBool(tempList[3])
		if err != nil {
			return []LiveJailStruct{}, err
		}

		tempStruct.Ip4address = tempList[4]
		tempStruct.Ip6address = tempList[5]

		jails = append(jails, tempStruct)
	}

	return jails, nil
}

type JailConfigFileStruct struct {
	CPULimitPercent  int      `json:"cpu_limit_percent"`
	RAMLimit         string   `json:"ram_limit"`
	Production       bool     `json:"production"`
	StartupScript    string   `json:"startup_script"`
	ShutdownScript   string   `json:"shutdown_script"`
	ConsoleOutputLog string   `json:"console_output_log"`
	ConfigFileAppend string   `json:"config_file_append"`
	StartAfter       string   `json:"start_after,omitempty"`
	StartupDelay     int      `json:"startup_delay,omitempty"`
	IPAddress        string   `json:"ip_address"`
	Network          string   `json:"network"`
	DNSServers       []string `json:"dns_servers"`
	Timezone         string   `json:"timezone"`
	Parent           string   `json:"parent"`
	Description      string   `json:"description"`

	// Not a part of JSON config
	JailName     string
	JailHostname string
	JailRootPath string
	JailFolder   string
	Netmask      string
	Running      bool
	Backup       bool
	CpuLimitReal int
}

func getAllJailsList() ([]string, error) {
	zfsDatasets, err := getZfsDatasetInfo()
	if err != nil {
		return []string{}, err
	}

	jails := []string{}
	for _, v := range zfsDatasets {
		mountPointDirWalk, err := os.ReadDir(v.MountPoint)
		if err != nil {
			return []string{}, err
		}

		for _, directory := range mountPointDirWalk {
			if directory.IsDir() {
				configFile := v.MountPoint + "/" + directory.Name() + "/jail_config.json"
				if FileExists(configFile) {
					jails = append(jails, directory.Name())
				}
			}
		}
	}

	return jails, nil
}

func checkJailOnline(jailConfig JailConfigFileStruct) (jailOnline bool, jailError error) {
	liveJails, err := getRunningJails()
	if err != nil {
		jailError = err
		return
	}

	for _, v := range liveJails {
		if v.Path == jailConfig.JailRootPath {
			jailOnline = true
			return
		}
	}

	return
}

func getJailReleaseInfo(jailConfig JailConfigFileStruct) (jailRelease string, jailError error) {
	jailOsReleaseFile, err := os.ReadFile(jailConfig.JailRootPath + "/etc/os-release")
	if err != nil {
		jailError = err
		return
	}

	reMatchVersion := regexp.MustCompile(`VERSION=`)
	reMatchQuotes := regexp.MustCompile(`"`)

	for _, v := range strings.Split(string(jailOsReleaseFile), "\n") {
		if reMatchVersion.MatchString(v) {
			v = reMatchVersion.ReplaceAllString(v, "")
			v = reMatchQuotes.ReplaceAllString(v, "")
			jailRelease = v
			return
		}
	}

	return
}
