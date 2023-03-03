package cmd

import (
	"errors"
	"hoster/emojlog"
	"os/exec"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize system services and kernel modules required by Hoster Core",
		Long:  `Initialize system services and kernel modules required by Hoster Core`,
		Run: func(cmd *cobra.Command, args []string) {
			err := loadMissingModules()
			if err != nil {
				emojlog.PrintLogMessage("Could not load kernel modules: "+err.Error(), emojlog.Warning)
			}
			err = applySysctls()
			if err != nil {
				emojlog.PrintLogMessage("Could not apply sysctls: "+err.Error(), emojlog.Warning)
			}
			err = loadNetworkConfig()
			if err != nil {
				emojlog.PrintLogMessage("Could not load network config: "+err.Error(), emojlog.Warning)
			}
			err = applyPfSettings()
			if err != nil {
				emojlog.PrintLogMessage("Could not reload pf: "+err.Error(), emojlog.Warning)
			}

			// Load Unbound settings (monkey patch for fresh installations, or user initiated config changes)
			restartUnbound()

			// Try to start Nebula if it's config file exists
			_, err = readNebulaClusterConfig()
			if err == nil {
				err := startNebulaService()
				if err != nil {
					emojlog.PrintLogMessage("Could not start Nebula service: "+err.Error(), emojlog.Warning)
				}
			}

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
	kernelModuleList := []string{"vmm", "nmdm", "if_bridge", "pf", "pflog"}
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
			return errors.New("error running sysctl: " + string(stdout) + " " + stderr.Error())
		}
		emojlog.PrintLogMessage("Applied Sysctl setting: sysctl net.link.tap.up_on_open=1", emojlog.Changed)

	}

	return nil
}

func loadNetworkConfig() error {
	networkInfoVar, err := networkInfo()
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
		if slices.Contains(loadedInterfaceList, v.Name) {
			emojlog.PrintLogMessage("Interface is up-to-date: vm-"+v.Name, emojlog.Debug)
		} else {
			stdout, stderr := exec.Command("ifconfig", "bridge", "create", "name", "vm-"+v.Name).CombinedOutput()
			if stderr != nil {
				return errors.New("error running ifconfig bridge create: " + string(stdout) + " " + stderr.Error())
			}
			emojlog.PrintLogMessage("Created a network bridge for VM use: vm-"+v.Name, emojlog.Changed)

			if v.BridgeInterface != "None" {
				stdout, stderr := exec.Command("ifconfig", "vm-"+v.Name, "addm", v.BridgeInterface).CombinedOutput()
				if stderr != nil {
					return errors.New("error running ifconfig bridge create: " + string(stdout) + " " + stderr.Error())
				}
				emojlog.PrintLogMessage("Bridged external interface with our VM network: "+v.BridgeInterface, emojlog.Changed)
			}

			if v.ApplyBridgeAddr {
				subnet := strings.Split(v.Subnet, "/")[1]
				stdout, stderr := exec.Command("ifconfig", "vm-"+v.Name, "inet", v.Gateway+"/"+subnet).CombinedOutput()
				if stderr != nil {
					return errors.New("error running ifconfig bridge create: " + string(stdout) + " " + stderr.Error())
				}
				emojlog.PrintLogMessage("Added IP address for vm-"+v.Name+" - "+v.Gateway+"/"+subnet, emojlog.Changed)
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

func restartUnbound() {
	err := generateNewDnsConfig()
	if err != nil {
		emojlog.PrintLogMessage("Could not generate new Unbound config: "+err.Error(), emojlog.Error)
	}

	unboundOut, unboundErr := exec.Command("service", "local_unbound", "restart").CombinedOutput()
	if unboundErr != nil {
		emojlog.PrintLogMessage("Could not restart Unbound: "+string(unboundOut), emojlog.Error)
	}
}
