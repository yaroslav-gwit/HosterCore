package cmd

import (
	"HosterCore/internal/pkg/emojlog"
	FreeBSDOsInfo "HosterCore/internal/pkg/freebsd/info"
	HosterNetwork "HosterCore/internal/pkg/hoster/network"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize system services and kernel modules required by Hoster Core",
		Long:  `Initialize system services and kernel modules required by Hoster Core`,
		Run: func(cmd *cobra.Command, args []string) {
			err := createInitFile()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				err = nil
			}

			err = loadMissingModules()
			if err != nil {
				emojlog.PrintLogMessage("Could not load kernel modules: "+err.Error(), emojlog.Error)
				err = nil
			}
			err = applySysctls()
			if err != nil {
				emojlog.PrintLogMessage("Could not apply sysctls: "+err.Error(), emojlog.Error)
				err = nil
			}
			err = loadNetworkConfig()
			if err != nil {
				emojlog.PrintLogMessage("Could not load network config: "+err.Error(), emojlog.Error)
				err = nil
			}

			// Stop and disable Local Unbound to avoid conflicts with our own DNS server
			stopAndDisableUnbound()

			err = startDnsServer()
			if err != nil {
				emojlog.PrintLogMessage("Could not start the internal DNS server: "+err.Error(), emojlog.Error)
				err = nil
			} else {
				emojlog.PrintLogMessage("Started the internal DNS server", emojlog.Changed)
			}

			err = applyPfSettings()
			if err != nil {
				emojlog.PrintLogMessage("Could not reload pf: "+err.Error(), emojlog.Error)
				err = nil
			}

			err = startSchedulerService()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				err = nil
			}

			// Start Node_Exporter_Custom
			err = startNodeExporter()
			if err != nil {
				emojlog.PrintLogMessage("Could not start NodeExporterCustom service: "+err.Error(), emojlog.Error)
				err = nil
			} else {
				emojlog.PrintLogMessage("NodeExporterCustom service has been started", emojlog.Changed)
			}

			// Remind user to mount the ZFS volumes
			emojlog.PrintLogMessage("Please don't forget to mount any encrypted ZFS volumes", emojlog.Info)
		},
	}
)

// kldload vmm
// kldload nmdm
// kldload if_bridge
// kldload if_tuntap
// kldload if_tap
// kldload pf
// kldload pflog
// sysctl net.link.tap.up_on_open=1

func loadMissingModules() error {
	moduleList, err := returnMissingModules()
	if err != nil {
		return err
	}
	for _, v := range moduleList {
		stdout, stderr := exec.Command("kldload", v).CombinedOutput()
		if stderr != nil {
			return errors.New("error running kldstat: " + string(stdout) + " " + stderr.Error())
		}
		emojlog.PrintLogMessage("Loaded kernel module: "+v, emojlog.Changed)
	}
	return nil
}

// Returns a list of kernel modules that are not yet loaded, or an error
func returnMissingModules() ([]string, error) {
	stdout, stderr := exec.Command("kldstat", "-v").CombinedOutput()
	if stderr != nil {
		return []string{}, errors.New("error running kldstat: " + string(stdout) + " " + stderr.Error())
	}
	reMatchKo := regexp.MustCompile(`\.ko`)
	reSplitSpace := regexp.MustCompile(`\s+`)

	var loadedModules []string
	kernelModuleList := []string{"vmm", "nmdm", "if_bridge", "if_vxlan", "if_epair", "if_tap", "if_tuntap", "pf", "pflog"}

	// Load CPU temperature kernel module
	cpuInfo, err := FreeBSDOsInfo.GetCpuInfo()
	if err != nil {
		return []string{}, err
	}
	reMatchIntelCpu := regexp.MustCompile(`.*[Ii]ntel.*`)
	if reMatchIntelCpu.MatchString(cpuInfo.Model) {
		kernelModuleList = append(kernelModuleList, "coretemp")
	} else {
		kernelModuleList = append(kernelModuleList, "amdtemp")
	}
	// EOF Load CPU temperature kernel module

	for _, v := range strings.Split(string(stdout), "\n") {
		if reMatchKo.MatchString(v) {
			for _, vv := range kernelModuleList {
				reMatchModule := regexp.MustCompile(vv + `\.ko`)
				reMatchModuleFinal := regexp.MustCompile(vv + `\.ko$`)
				if reMatchModule.MatchString(v) {
					tempList := reSplitSpace.Split(v, -1)
					for _, vvv := range tempList {
						if reMatchModuleFinal.MatchString(vvv) {
							vvv = reMatchKo.ReplaceAllString(vvv, "")
							loadedModules = append(loadedModules, strings.TrimSpace(vvv))
						}
					}
				}
			}
		}
	}

	kernelModuleListNoKo := []string{"if_tuntap", "if_tap"}
	for _, v := range strings.Split(string(stdout), "\n") {
		for _, vv := range kernelModuleListNoKo {
			reMatchModule := regexp.MustCompile(vv)
			if reMatchModule.MatchString(v) {
				tempList := reSplitSpace.Split(v, -1)
				for _, vvv := range tempList {
					if reMatchModule.MatchString(vvv) {
						loadedModules = append(loadedModules, strings.TrimSpace(vvv))
					}
				}
			}
		}
	}

	stdout, stderr = exec.Command("sysctl", "net.link.tap.up_on_open").CombinedOutput()
	if stderr != nil {
		return []string{}, errors.New("error running sysctl: " + string(stdout) + " " + stderr.Error())
	}
	for _, v := range loadedModules {
		emojlog.PrintLogMessage("Module is already active: "+v, emojlog.Debug)
	}

	var result []string
	for _, v := range kernelModuleList {
		if !slices.Contains(loadedModules, v) {
			result = append(result, v)
		}
	}

	for _, v := range kernelModuleListNoKo {
		if !slices.Contains(loadedModules, v) {
			result = append(result, v)
		}
	}

	return result, nil
}

func applySysctls() error {
	stdout, stderr := exec.Command("sysctl", "net.link.tap.up_on_open").CombinedOutput()
	if stderr != nil {
		return errors.New("error running sysctl: " + string(stdout) + " " + stderr.Error())
	}

	reSplitSpace := regexp.MustCompile(`\s+`)

	tapUpOnOpen := false
	for _, v := range reSplitSpace.Split(string(stdout), -1) {
		if v == "1" {
			tapUpOnOpen = true
			emojlog.PrintLogMessage("Sysctl is already active: net.link.tap.up_on_open", emojlog.Debug)
		}
	}

	if !tapUpOnOpen {
		err := exec.Command("sysctl", "net.link.tap.up_on_open=1").Run()
		if err != nil {
			return errors.New("error running sysctl: " + err.Error())
		}
		emojlog.PrintLogMessage("Applied Sysctl setting: sysctl net.link.tap.up_on_open=1", emojlog.Changed)

	}

	return nil
}

func loadNetworkConfig() error {
	networkInfoVar, err := HosterNetwork.GetNetworkConfig()
	if err != nil {
		return err
	}

	stdout, stderr := exec.Command("ifconfig").CombinedOutput()
	if stderr != nil {
		return errors.New("error running ifconfig: " + string(stdout) + " " + stderr.Error())
	}

	// reSplitSpace := regexp.MustCompile(`\s+`)
	reMatchVmInterface := regexp.MustCompile(`vm-.*:`)
	reReplacePrefix := regexp.MustCompile(`vm-`)
	reReplaceSuffix := regexp.MustCompile(`:.*`)
	var loadedInterfaceList []string
	for _, v := range strings.Split(string(stdout), "\n") {
		if reMatchVmInterface.MatchString(v) {
			v = reReplacePrefix.ReplaceAllLiteralString(v, "")
			v = reReplaceSuffix.ReplaceAllLiteralString(v, "")
			loadedInterfaceList = append(loadedInterfaceList, v)
		}
	}

	for _, v := range networkInfoVar {
		if slices.Contains(loadedInterfaceList, v.NetworkName) {
			emojlog.PrintLogMessage("Interface is up-to-date: vm-"+v.NetworkName, emojlog.Debug)
		} else {
			stdout, stderr := exec.Command("ifconfig", "bridge", "create", "name", "vm-"+v.NetworkName).CombinedOutput()
			if stderr != nil {
				return errors.New("error running ifconfig bridge create: " + string(stdout) + " " + stderr.Error())
			}
			emojlog.PrintLogMessage("Created a network bridge for VM use: vm-"+v.NetworkName, emojlog.Changed)

			if v.BridgeInterface != "None" {
				stdout, stderr := exec.Command("ifconfig", "vm-"+v.NetworkName, "addm", v.BridgeInterface).CombinedOutput()
				if stderr != nil {
					return errors.New("error running ifconfig bridge create: " + string(stdout) + " " + stderr.Error())
				}
				emojlog.PrintLogMessage("Bridged external interface with our VM network: "+v.BridgeInterface, emojlog.Changed)
			}

			if v.ApplyBridgeAddr {
				subnet := strings.Split(v.Subnet, "/")[1]
				stdout, stderr := exec.Command("ifconfig", "vm-"+v.NetworkName, "inet", v.Gateway+"/"+subnet).CombinedOutput()
				if stderr != nil {
					return errors.New("error running ifconfig bridge create: " + string(stdout) + " " + stderr.Error())
				}
				emojlog.PrintLogMessage("Added IP address for vm-"+v.NetworkName+" - "+v.Gateway+"/"+subnet, emojlog.Changed)
			}
		}
	}

	return nil
}

func applyPfSettings() error {
	stdout, stderr := exec.Command("pfctl", "-f", "/etc/pf.conf").CombinedOutput()
	if stderr != nil {
		return errors.New("error running pfctl: " + string(stdout) + " " + stderr.Error())
	}
	emojlog.PrintLogMessage("pf Settings have been applied: pfctl -f /etc/pf.conf", emojlog.Changed)
	return nil
}

func stopAndDisableUnbound() {
	exec.Command("service", "local_unbound", "stop").Run()
	exec.Command("service", "local_unbound", "disable").Run()
}

func createInitFile() error {
	stdout, stderr := exec.Command("touch", "/var/run/hoster_init").CombinedOutput()
	if stderr != nil {
		return errors.New("error creating hoster_init file: " + string(stdout) + " " + stderr.Error())
	}

	stdout, stderr = exec.Command("chmod", "0600", "/var/run/hoster_init").CombinedOutput()
	if stderr != nil {
		return errors.New("error setting permissions for hoster_init file: " + string(stdout) + " " + stderr.Error())
	}

	emojlog.PrintLogMessage("State file /var/run/hoster_init has been created", emojlog.Changed)
	return nil
}

func checkInitFile() {
	_, stderr := exec.Command("ls", "/var/run/hoster_init").CombinedOutput()
	if stderr != nil {
		// Add documentation for this error output in the online docs at some point
		emojlog.PrintLogMessage("Please, execute `hoster init` to start using this utility", emojlog.Error)
		os.Exit(1)
		// return errors.New("hoster process state file is missing")
	}

	if syscall.Geteuid() != 0 {
		emojlog.PrintLogMessage("`hoster` must be executed using elevated privileges (via root or sudo/doas)", emojlog.Error)
		os.Exit(1)
	}
	// return nil
}

func FileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	//return !os.IsNotExist(err)
	return !errors.Is(error, os.ErrNotExist)
}
