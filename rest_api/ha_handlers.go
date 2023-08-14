package main

import (
	"fmt"
	"os/exec"
	"time"

	"hoster/cmd"

	"github.com/gofiber/fiber/v2"
)

func handleHaManagerRegistration(fiberContext *fiber.Ctx) error {
	tagCustomError = ""

	hosterNode := new(NodeStruct)
	if err := fiberContext.BodyParser(hosterNode); err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusUnprocessableEntity)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	hosterHaNode := HosterHaNodeStruct{}
	hosterHaNode.IsManager = false
	hosterHaNode.IsCandidate = false
	hosterHaNode.IsWorker = true
	hosterHaNode.NodeInfo = *hosterNode
	hosterHaNode.NodeInfo.Address = fiberContext.IP()
	hosterHaNode.LastPing = time.Now().Unix()

	haChannelAdd <- hosterHaNode
	return fiberContext.JSON(fiber.Map{"message": "done", "context": hosterHaNode})
}

func handleHaPing(fiberContext *fiber.Ctx) error {
	tagCustomError = ""
	hosterNode := new(NodeStruct)
	if err := fiberContext.BodyParser(hosterNode); err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusUnprocessableEntity)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	hosterHaNode := HosterHaNodeStruct{}
	hosterHaNode.NodeInfo = *hosterNode
	hosterHaNode.NodeInfo.Address = fiberContext.IP()
	hosterHaNode.LastPing = time.Now().Unix()

	haChannelAdd <- hosterHaNode
	return fiberContext.JSON(fiber.Map{"message": "pong"})
}

func handleHaTerminate(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleHaPromote(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleHaMonitor(fiberContext *fiber.Ctx) error {
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleHaShareWorkers(fiberContext *fiber.Ctx) error {
	workers := []HosterHaNodeStruct{}
	for _, v := range haHostsDb {
		if v.IsWorker {
			workers = append(workers, v)
		}
	}
	return fiberContext.JSON(workers)
}

func handleHaShareManager(fiberContext *fiber.Ctx) error {
	manager := HosterHaNodeStruct{}
	for _, v := range haHostsDb {
		if v.IsManager {
			manager = v
		}
	}
	return fiberContext.JSON(manager)
}

func handleHaShareCandidates(fiberContext *fiber.Ctx) error {
	candidates := []HosterHaNodeStruct{}
	for _, v := range haHostsDb {
		if v.IsCandidate {
			candidates = append(candidates, v)
		}
	}
	return fiberContext.JSON(candidates)
}

func handleHaShareAllMembers(fiberContext *fiber.Ctx) error {
	if iAmManager {
		return fiberContext.JSON(haHostsDb)
	} else {
		return fiberContext.JSON([]HosterHaNodeStruct{})
	}
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
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "haVmsList() Recovered from error: "+errorValue).Run()
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
		if !cmd.VmIsInProduction(vmConfig.LiveStatus) {
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
	return fiberContext.JSON(haVmsList())
}
