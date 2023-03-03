package cmd

import (
	"errors"
	"fmt"
	"hoster/emojlog"
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
	vmStartCmd = &cobra.Command{
		Use:   "start [vmName]",
		Short: "Start a particular VM using it's name",
		Long:  `Start a particular VM using it's name`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			err := vmStart(args[0])
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func vmStart(vmName string) error {
	allVms := getAllVms()
	if !slices.Contains(allVms, vmName) {
		return errors.New("VM is not found on this system")
	} else if vmLiveCheck(vmName) {
		return errors.New("VM is already up-and-running")
	}

	emojlog.PrintLogMessage("Starting the VM: "+vmName, emojlog.Info)

	// Generate bhyve start command
	bhyveCommand := generateBhyveStartCommand(vmName)
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

	emojlog.PrintLogMessage("VM started: "+vmName, emojlog.Changed)
	return nil
}

func generateBhyveStartCommand(vmName string) string {
	vmConfigVar := vmConfig(vmName)

	var availableTaps []string
	for _, v := range vmConfigVar.Networks {
		availableTap := findAvailableTapInterface()
		availableTaps = append(availableTaps, availableTap)
		emojlog.PrintLogMessage("Creating the TAP interface: "+availableTap, emojlog.Debug)
		createTapInterface := "ifconfig " + availableTap + " create"
		parts := strings.Fields(createTapInterface)
		emojlog.PrintLogMessage("Executing: "+createTapInterface, emojlog.Debug)
		exec.Command(parts[0], parts[1:]...).Run()

		bridgeTapInterface := "ifconfig vm-" + v.NetworkBridge + " addm " + availableTap
		parts = strings.Fields(bridgeTapInterface)
		emojlog.PrintLogMessage("Executing: "+bridgeTapInterface, emojlog.Debug)
		exec.Command(parts[0], parts[1:]...).Run()

		upBridgeInterface := "ifconfig vm-" + v.NetworkBridge + " up"
		parts = strings.Fields(upBridgeInterface)
		emojlog.PrintLogMessage("Executing: "+upBridgeInterface, emojlog.Debug)
		exec.Command(parts[0], parts[1:]...).Run()

		setTapDescription1 := "ifconfig"
		setTapDescription2 := availableTap
		setTapDescription3 := "description"
		setTapDescription4 := "\"" + availableTap + " " + vmName + " interface -> " + v.NetworkBridge + "\""
		setTapDescription := fmt.Sprint(setTapDescription1 + " " + setTapDescription2 + " " + setTapDescription3 + " " + setTapDescription4)
		emojlog.PrintLogMessage("Executing: "+setTapDescription, emojlog.Debug)
		exec.Command(setTapDescription1, setTapDescription2, setTapDescription3, setTapDescription4).Run()
	}

	bhyveFinalCommand := "bhyve -HAw -s 0:0,hostbridge -s 31,lpc "
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
				bhyvePci2 = bhyvePci2 + 1
				networkFinal = networkFinal + " -s " + strconv.Itoa(bhyvePci1) + ":" + strconv.Itoa(bhyvePci2) + networkAdaptorType + availableTaps[i] + ",mac=" + v.NetworkMac
			}
		}
	} else {
		networkAdaptorType = "," + vmConfigVar.Networks[0].NetworkAdaptorType + ","
		networkFinal = "-s " + strconv.Itoa(bhyvePci1) + ":" + strconv.Itoa(bhyvePci2) + networkAdaptorType + availableTaps[0] + ",mac=" + vmConfigVar.Networks[0].NetworkMac
	}

	bhyveFinalCommand = bhyveFinalCommand + networkFinal
	// fmt.Println(bhyveFinalCommand)

	bhyvePci := 3
	var diskFinal string
	var genericDiskText string
	var diskImageLocation string
	if len(vmConfigVar.Disks) > 1 {
		for i, v := range vmConfigVar.Disks {
			diskImageLocation = getVmFolder(vmName) + "/" + v.DiskImage
			genericDiskText = ":0," + v.DiskType + ","
			if i == 0 {
				diskFinal = " -s " + strconv.Itoa(bhyvePci) + genericDiskText + diskImageLocation
			} else {
				bhyvePci = bhyvePci + 1
				diskFinal = diskFinal + " -s " + strconv.Itoa(bhyvePci) + genericDiskText + diskImageLocation
			}
		}
	} else {
		diskImageLocation = getVmFolder(vmName) + "/" + vmConfigVar.Disks[0].DiskImage
		genericDiskText = ":0," + vmConfigVar.Disks[0].DiskType + ","
		diskFinal = " -s " + strconv.Itoa(bhyvePci) + genericDiskText + diskImageLocation
	}

	bhyveFinalCommand = bhyveFinalCommand + diskFinal
	// fmt.Println(bhyveFinalCommand)

	cpuAndRam := " -c sockets=" + vmConfigVar.CPUSockets + ",cores=" + vmConfigVar.CPUCores + " -m " + vmConfigVar.Memory
	bhyveFinalCommand = bhyveFinalCommand + cpuAndRam
	// fmt.Println(bhyveFinalCommand)

	bhyvePci = bhyvePci + 1
	vncCommand := " -s " + strconv.Itoa(bhyvePci) + ":" + strconv.Itoa(bhyvePci2) + ",fbuf,tcp=0.0.0.0:" + vmConfigVar.VncPort + ",w=1280,h=1024,password=" + vmConfigVar.VncPassword
	bhyveFinalCommand = bhyveFinalCommand + vncCommand
	// fmt.Println(bhyveFinalCommand)

	bhyvePci = bhyvePci + 1
	var loaderCommand string
	if vmConfigVar.Loader == "bios" {
		loaderCommand = " -s " + strconv.Itoa(bhyvePci) + ":" + strconv.Itoa(bhyvePci2) + ",xhci,tablet -l com1,/dev/nmdm-" + vmName + "-1A -l bootrom,/usr/local/share/uefi-firmware/BHYVE_UEFI_CSM.fd -u " + vmName
	} else if vmConfigVar.Loader == "uefi" {
		loaderCommand = " -s " + strconv.Itoa(bhyvePci) + ":" + strconv.Itoa(bhyvePci2) + ",xhci,tablet -l com1,/dev/nmdm-" + vmName + "-1A -l bootrom,/usr/local/share/uefi-firmware/BHYVE_UEFI.fd -u " + vmName
	} else {
		log.Fatal("Please make sure your loader is set to 'bios' or 'uefi'")
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
