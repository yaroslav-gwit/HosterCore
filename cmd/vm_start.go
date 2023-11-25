package cmd

import (
	"HosterCore/emojlog"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	vmStartCmdRestoreVmState bool
	vmStartCmdWaitForVnc     bool
	vmStartCmdDebug          bool

	vmStartCmd = &cobra.Command{
		Use:   "start [vmName]",
		Short: "Start a particular VM using it's name",
		Long:  `Start a particular VM using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}

			err = VmStart(args[0], vmStartCmdRestoreVmState, vmStartCmdWaitForVnc, vmStartCmdDebug)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

// Starts the VM using BhyveCTL and vm_supervisor_service
func VmStart(vmName string, restoreVmState bool, waitForVnc bool, startDebug bool) error {
	allVms := getAllVms()
	if !slices.Contains(allVms, vmName) {
		return errors.New("VM is not found on this system")
	} else if VmLiveCheck(vmName) {
		return errors.New("VM is already up-and-running")
	}

	emojlog.PrintLogMessage("Starting the VM: "+vmName, emojlog.Info)

	// Generate bhyve start command
	bhyveCommand := generateBhyveStartCommand(vmName, restoreVmState, waitForVnc, startDebug)
	// Set env vars to send to "vm_supervisor"
	os.Setenv("VM_START", bhyveCommand)
	os.Setenv("VM_NAME", vmName)
	os.Setenv("LOG_FILE", getVmFolder(vmName)+"/vm_supervisor.log")
	// Get location of the "hoster" executable, as "vm_supervisor" executable is in the same directory
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	// Start VM supervisor process
	if !startDebug {
		execFile := path.Dir(execPath) + "/vm_supervisor_service"
		cmd := exec.Command("nohup", execFile, "for", vmName, "&")
		err = cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		go func() {
			err := cmd.Wait()
			if err != nil {
				log.Println(err)
			}
		}()
	}

	emojlog.PrintLogMessage("VM started: "+vmName, emojlog.Changed)
	return nil
}

func generateBhyveStartCommand(vmName string, restoreVmState bool, waitForVnc bool, debugStart bool) string {
	vmConfigVar := vmConfig(vmName)

	var availableTaps []string
	for _, v := range vmConfigVar.Networks {
		availableTap := findAvailableTapInterface()
		availableTaps = append(availableTaps, availableTap)
		emojlog.PrintLogMessage("Creating the TAP interface: "+availableTap, emojlog.Debug)
		createTapInterface := "ifconfig " + availableTap + " create"
		parts := strings.Fields(createTapInterface)
		emojlog.PrintLogMessage("Executing: "+createTapInterface, emojlog.Debug)
		if !debugStart {
			exec.Command(parts[0], parts[1:]...).Run()
		}

		bridgeTapInterface := "ifconfig vm-" + v.NetworkBridge + " addm " + availableTap
		parts = strings.Fields(bridgeTapInterface)
		emojlog.PrintLogMessage("Executing: "+bridgeTapInterface, emojlog.Debug)
		if !debugStart {
			exec.Command(parts[0], parts[1:]...).Run()
		}

		upBridgeInterface := "ifconfig vm-" + v.NetworkBridge + " up"
		parts = strings.Fields(upBridgeInterface)
		emojlog.PrintLogMessage("Executing: "+upBridgeInterface, emojlog.Debug)
		if !debugStart {
			exec.Command(parts[0], parts[1:]...).Run()
		}

		tapDescription := "\"" + availableTap + " " + vmName + " interface -> " + v.NetworkBridge + "\""
		emojlog.PrintLogMessage(fmt.Sprintf("Executing: ifconfig %s description %s", availableTap, tapDescription), emojlog.Debug)
		if !debugStart {
			exec.Command("ifconfig", availableTap, "description", tapDescription).Run()
		}
	}

	// bhyveFinalCommand := "bhyve -HAw -s 0:0,hostbridge -s 31,lpc "
	var bhyveFinalCommand string
	if len(vmConfigVar.Passthru) > 0 {
		bhyveFinalCommand = "bhyve -S -HAuw -s 0:0,hostbridge -s 31,lpc "
	} else {
		bhyveFinalCommand = "bhyve -HAuw -s 0:0,hostbridge -s 31,lpc "
	}

	bhyvePci1 := 2
	bhyvePci2 := 0

	var networkFinal string
	var networkAdaptorType string
	if len(vmConfigVar.Networks) > 1 {
		for i, v := range vmConfigVar.Networks {
			networkAdaptorType = "," + v.NetworkAdaptorType + ","
			if i == 0 {
				networkFinal = "-s " + strconv.Itoa(bhyvePci1) + ":" + strconv.Itoa(bhyvePci2) + networkAdaptorType + availableTaps[i] + ",mac=" + v.NetworkMac
			} else {
				// bhyvePci2 = bhyvePci2 + 1
				bhyvePci1 += 1
				networkFinal = networkFinal + " -s " + strconv.Itoa(bhyvePci1) + ":" + strconv.Itoa(bhyvePci2) + networkAdaptorType + availableTaps[i] + ",mac=" + v.NetworkMac
			}
		}
	} else {
		networkAdaptorType = "," + vmConfigVar.Networks[0].NetworkAdaptorType + ","
		networkFinal = "-s " + strconv.Itoa(bhyvePci1) + ":" + strconv.Itoa(bhyvePci2) + networkAdaptorType + availableTaps[0] + ",mac=" + vmConfigVar.Networks[0].NetworkMac
	}

	bhyveFinalCommand = bhyveFinalCommand + networkFinal
	// fmt.Println(bhyveFinalCommand)

	bhyvePci := bhyvePci1 + 1
	var diskFinal string
	var genericDiskText string
	var diskImageLocation string
	if len(vmConfigVar.Disks) > 1 {
		for i, v := range vmConfigVar.Disks {
			if v.DiskLocation == "internal" {
				diskImageLocation = getVmFolder(vmName) + "/" + v.DiskImage
			} else {
				diskImageLocation = v.DiskImage
			}
			genericDiskText = ":0," + v.DiskType + ","
			if i == 0 {
				diskFinal = " -s " + strconv.Itoa(bhyvePci) + genericDiskText + diskImageLocation
			} else {
				bhyvePci = bhyvePci + 1
				diskFinal = diskFinal + " -s " + strconv.Itoa(bhyvePci) + genericDiskText + diskImageLocation
			}
		}
	} else {
		if vmConfigVar.Disks[0].DiskLocation == "internal" {
			diskImageLocation = getVmFolder(vmName) + "/" + vmConfigVar.Disks[0].DiskImage
		} else {
			diskImageLocation = vmConfigVar.Disks[0].DiskImage
		}
		genericDiskText = ":0," + vmConfigVar.Disks[0].DiskType + ","
		diskFinal = " -s " + strconv.Itoa(bhyvePci) + genericDiskText + diskImageLocation
	}

	bhyveFinalCommand = bhyveFinalCommand + diskFinal
	// fmt.Println(bhyveFinalCommand)

	cpuAndRam := " -c sockets=" + vmConfigVar.CPUSockets + ",cores=" + vmConfigVar.CPUCores + " -m " + vmConfigVar.Memory
	bhyveFinalCommand = bhyveFinalCommand + cpuAndRam
	// fmt.Println(bhyveFinalCommand)

	// VNC options
	bhyvePci = bhyvePci + 1
	bhyvePci2 = 0
	vncCommand := " -s " + strconv.Itoa(bhyvePci) + ":" + strconv.Itoa(bhyvePci2) + ",fbuf,tcp=0.0.0.0:" + vmConfigVar.VncPort + ",w=800,h=600,password=" + vmConfigVar.VncPassword
	// Set the VGA mode if found in the config file
	if len(vmConfigVar.VGA) > 0 {
		if vmConfigVar.VGA == "io" || vmConfigVar.VGA == "on" || vmConfigVar.VGA == "off" {
			_ = 0
		} else {
			emojlog.PrintLogMessage("vga config option may only be set to 'on', 'off', or 'io', expect VM boot failures", emojlog.Warning)
		}
		vncCommand = vncCommand + ",vga=" + vmConfigVar.VGA
	}
	// Wait for the VNC connection before booting the VM
	if waitForVnc {
		vncCommand = vncCommand + ",wait"
	}
	// EOF VNC options
	bhyveFinalCommand = bhyveFinalCommand + vncCommand
	// fmt.Println(bhyveFinalCommand)

	// Passthru section
	passthruString := ""
	if len(vmConfigVar.Passthru) > 0 {
		// PCI slot added for each card iteration, no need to add one before the loop
		bhyvePci = bhyvePci + 1
		// bhyvePci2 = 0
		// for _, v := range vmConfigVar.Passthru {
		// 	bhyvePci = bhyvePci + 1
		// 	bhyveFinalCommand = bhyveFinalCommand + " -s " + strconv.Itoa(bhyvePci) + ",passthru," + v
		// }
		passthruString, bhyvePci = passthruPciSplitter(bhyvePci, vmConfigVar.Passthru)
		bhyveFinalCommand = bhyveFinalCommand + " " + passthruString
	}
	// EOF Passthru section

	// XHCI and LOADER section
	bhyvePci = bhyvePci + 1
	bhyvePci2 = 0

	var xhciCommand string
	if vmConfigVar.DisableXHCI {
		xhciCommand = ""
	} else {
		xhciCommand = " -s " + strconv.Itoa(bhyvePci) + ":" + strconv.Itoa(bhyvePci2) + ",xhci,tablet"
	}

	var loaderCommand string
	if vmConfigVar.Loader == "bios" {
		loaderCommand = xhciCommand + " -l com1,/dev/nmdm-" + vmName + "-1A -l bootrom,/usr/local/share/uefi-firmware/BHYVE_UEFI_CSM.fd"
	} else if vmConfigVar.Loader == "uefi" {
		loaderCommand = xhciCommand + " -l com1,/dev/nmdm-" + vmName + "-1A -l bootrom,/usr/local/share/uefi-firmware/BHYVE_UEFI.fd"
	} else {
		emojlog.PrintLogMessage("make sure your loader is set to 'bios' or 'uefi', expect VM boot failures", emojlog.Warning)
	}
	// EOF XHCI and LOADER section

	if len(vmConfigVar.UUID) > 0 {
		loaderCommand = loaderCommand + " -U " + vmConfigVar.UUID
	}

	if restoreVmState {
		vmFolder := getVmFolder(vmName)
		loaderCommand = loaderCommand + " -r " + vmFolder + "/vm_state"
	} else {
		loaderCommand = loaderCommand + " -u " + vmName
	}

	bhyveFinalCommand = bhyveFinalCommand + loaderCommand
	emojlog.PrintLogMessage("Executing bhyve command: "+bhyveFinalCommand, emojlog.Debug)
	return bhyveFinalCommand
}

func findAvailableTapInterface() string {
	cmd := exec.Command("ifconfig")
	stdout, stderr := cmd.Output()
	if stderr != nil {
		log.Fatal("ifconfig exited with an error " + stderr.Error())
	}

	reMatchTap, _ := regexp.Compile(`^tap`)

	var tapList []int
	var trimmedTap string
	for _, v := range strings.Split(string(stdout), "\n") {
		trimmedTap = strings.Trim(v, "")
		if reMatchTap.MatchString(trimmedTap) {
			for _, vv := range strings.Split(trimmedTap, ":") {
				if reMatchTap.MatchString(vv) {
					vv = strings.Replace(vv, "tap", "", 1)
					vvInt, err := strconv.Atoi(vv)
					if err != nil {
						log.Fatal("Could not convert tap int: " + err.Error())
					}
					tapList = append(tapList, vvInt)
				}
			}
		}
	}

	nextFreeTap := 0
	for {
		if slices.Contains(tapList, nextFreeTap) {
			nextFreeTap = nextFreeTap + 1
		} else {
			return "tap" + strconv.Itoa(nextFreeTap)
		}
	}
}

// Takes in a list of PCI devices like so: [ "4/0/0", "43/0/1", "43/0/12" ]
//
// And returns the correct mappings for the Bhyve passthru, like so:
// -s 6:0,passthru,4/0/0 -s 7:1,passthru,43/0/1 -s 7:12,passthru,43/0/12
func passthruPciSplitter(startWithId int, devices []string) (pciDevs string, latestPciId int) {
	group0 := ""
	group1 := ""
	iter := 0
	skipThisIterationList := []int{}

	for i, v := range devices {
		if len(strings.Split(v, "/")) < 3 {
			emojlog.PrintLogMessage("This PCI device would not be added to passthru: "+v+", because it uses the incorrect format", emojlog.Error)
			continue
		}
		if slices.Contains(skipThisIterationList, i) {
			continue
		}

		group0 = strings.Split(v, "/")[0]
		group1 = strings.Split(v, "/")[1]
		iter = i

		if strings.HasPrefix(v, "-") {
			vNoPrefix := strings.TrimPrefix(v, "-")
			pciDevs = pciDevs + " -s " + strconv.Itoa(startWithId) + ":0,passthru," + vNoPrefix
		} else {
			pciDevs = pciDevs + " -s " + strconv.Itoa(startWithId) + ":" + strings.Split(v, "/")[2] + ",passthru," + v
		}
		for ii, vv := range devices {
			if ii == iter || strings.HasPrefix(vv, "-") || strings.HasPrefix(v, "-") {
				continue
			}
			if slices.Contains(skipThisIterationList, ii) {
				continue
			}
			if len(strings.Split(vv, "/")) < 3 {
				emojlog.PrintLogMessage("This PCI device would not be added to passthru: "+vv+", because it uses the incorrect format", emojlog.Error)
				continue
			}

			vvSpit := strings.Split(vv, "/")
			if strings.TrimSpace(vvSpit[0]) == group0 && strings.TrimSpace(vvSpit[1]) == group1 {
				skipThisIterationList = append(skipThisIterationList, ii)
				pciDevs = pciDevs + " -s " + strconv.Itoa(startWithId) + ":" + strings.Split(vv, "/")[2] + ",passthru," + vv
			}
		}
		startWithId += 1
	}

	pciDevs = strings.TrimSpace(pciDevs)
	if iter > 0 {
		latestPciId = startWithId - 1
	}
	return
}
