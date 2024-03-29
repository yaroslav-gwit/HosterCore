package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"HosterCore/cmd"

	"github.com/gofiber/fiber/v2"
)

func handleHaRegistration(fiberContext *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: handleHaRegistration(): "+errorValue).Run()
		}
	}()

	tagCustomError = ""

	hosterNode := new(NodeInfoStruct)
	if err := fiberContext.BodyParser(hosterNode); err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusUnprocessableEntity)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	hosterHaNode := HosterHaNodeStruct{}
	hosterHaNode.NodeInfo = *hosterNode
	hosterHaNode.NodeInfo.Address = fiberContext.IP()
	hosterHaNode.LastPing = time.Now().Unix()

	modifyHostsDb(ModifyHostsDbStruct{addOrUpdate: true, data: hosterHaNode}, &hostsDbLock)

	return fiberContext.JSON(fiber.Map{"message": "done", "context": hosterHaNode})
}

func handleHaPing(fiberContext *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: handleHaPing(): "+errorValue).Run()
		}
	}()

	tagCustomError = ""
	hosterNode := new(NodeInfoStruct)
	if err := fiberContext.BodyParser(hosterNode); err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusUnprocessableEntity)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	hosterHaNode := HosterHaNodeStruct{}
	hosterHaNode.NodeInfo = *hosterNode
	hosterHaNode.NodeInfo.Address = fiberContext.IP()
	hosterHaNode.LastPing = time.Now().Unix()

	modifyHostsDb(ModifyHostsDbStruct{addOrUpdate: true, data: hosterHaNode}, &hostsDbLock)

	return fiberContext.JSON(fiber.Map{"message": "pong"})
}

func handleHaTerminate(fiberContext *fiber.Ctx) error {
	go func() {
		service := cmd.ApiProcessServiceInfo()
		_ = exec.Command("logger", "-t", "HOSTER_HA_WATCHDOG", "INFO: received a remote terminating call").Run()
		_ = exec.Command("kill", "-SIGTERM", strconv.Itoa(service.HaWatchDogPid)).Run()
	}()

	go func() {
		time.Sleep(1500 * time.Millisecond)
		_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "INFO: received a remote terminating call").Run()
		os.Exit(0)
	}()

	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleHaShareAllMembers(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(haHostsDb)
}

func handleHaMonitor(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

type HaVm struct {
	VmName         string `json:"vm_name"`
	Live           bool   `json:"live"`
	LatestSnapshot string `json:"latest_snapshot"`
	ParentHost     string `json:"parent_host"`
	CurrentHost    string `json:"current_host"`
}

func haVmsList() []HaVm {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: haVmsList(): "+errorValue).Run()
		}
	}()

	haVms := []HaVm{}
	vmList := cmd.GetAllVms()

	for _, vm := range vmList {
		vmFound := false
		for _, haVm := range haVms {
			if vm == haVm.VmName {
				vmFound = true
			}
		}
		if vmFound {
			continue
		}

		vmConfig := cmd.VmConfig(vm)
		if !cmd.IsVmInProduction(vmConfig.LiveStatus) {
			continue
		}

		haVmTemp := HaVm{}
		haVmTemp.VmName = vm
		haVmTemp.ParentHost = vmConfig.ParentHost
		haVmTemp.CurrentHost = cmd.GetHostName()
		haVmTemp.Live = cmd.VmLiveCheck(vm)
		snapshotList, _ := cmd.GetSnapshotInfo(vm, true)
		if len(snapshotList) > 0 {
			haVmTemp.LatestSnapshot = snapshotList[len(snapshotList)-1].Name
		}

		haVms = append(haVms, haVmTemp)
	}

	return haVms
}

func handleHaVmsList(fiberContext *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: handleHaVmsList(): "+errorValue).Run()
		}
	}()

	return fiberContext.JSON(haVmsList())
}
