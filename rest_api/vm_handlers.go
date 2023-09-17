package main

import (
	"HosterCore/cmd"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/gofiber/fiber/v2"
)

func handleVmStart(fiberContext *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: handleVmStart(): "+errorValue).Run()
		}
	}()

	type vmName struct {
		Name string `json:"name" xml:"name" form:"name"`
	}

	tagCustomError = ""
	vm := new(vmName)
	if err := fiberContext.BodyParser(vm); err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusUnprocessableEntity)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	if cmd.VmLiveCheck(vm.Name) {
		fiberContext.Status(fiber.StatusInternalServerError)
		return fiberContext.JSON(fiber.Map{"message": "vm is already live"})
	}

	vmConfig := cmd.VmConfig(vm.Name)
	if vmConfig.ParentHost != cmd.GetHostName() {
		fiberContext.Status(fiber.StatusInternalServerError)
		return fiberContext.JSON(fiber.Map{"message": "vm is a backup from another host"})
	}

	// Using NOHUP option in order to avoid killing the VMs process when API server stops
	execPath, err := os.Executable()
	if err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusInternalServerError)
		return fiberContext.JSON(fiber.Map{"message": "failed to start the process"})
	}
	execFile := path.Dir(execPath) + "/hoster"
	// Execute vm start from the terminal using nohup
	cmd := exec.Command("nohup", execFile, "vm", "start", vm.Name)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err = cmd.Start()
	if err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusInternalServerError)
		return fiberContext.JSON(fiber.Map{"message": "failed to start the process"})
	}
	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Println(err.Error())
		}
	}()

	fiberContext.Status(fiber.StatusOK)
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleVmChangeParent(fiberContext *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: handleVmChangeParent(): "+errorValue).Run()
		}
	}()

	type vmName struct {
		Name string `json:"name" xml:"name" form:"name"`
	}

	tagCustomError = ""
	vm := new(vmName)
	if err := fiberContext.BodyParser(vm); err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusUnprocessableEntity)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	vmConfig := cmd.VmConfig(vm.Name)
	if vmConfig.ParentHost == cmd.GetHostName() {
		fiberContext.Status(fiber.StatusInternalServerError)
		return fiberContext.JSON(fiber.Map{"message": "vm is not a backup"})
	}

	err := cmd.ReplaceParent(vm.Name, cmd.GetHostName(), false)
	if err != nil {
		fiberContext.Status(fiber.StatusInternalServerError)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	fiberContext.Status(fiber.StatusOK)
	return fiberContext.JSON(fiber.Map{"message": "done"})
}

func handleVmCiReset(fiberContext *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: handleVmCiReset(): "+errorValue).Run()
		}
	}()

	type vmName struct {
		Name string `json:"name" xml:"name" form:"name"`
	}

	tagCustomError = ""
	vm := new(vmName)
	if err := fiberContext.BodyParser(vm); err != nil {
		tagCustomError = err.Error()
		fiberContext.Status(fiber.StatusUnprocessableEntity)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	vmConfig := cmd.VmConfig(vm.Name)
	if vmConfig.ParentHost == cmd.GetHostName() {
		fiberContext.Status(fiber.StatusInternalServerError)
		return fiberContext.JSON(fiber.Map{"error": "vm is not a backup"})
	}

	err := cmd.CiReset(vm.Name, vm.Name)
	if err != nil {
		fiberContext.Status(fiber.StatusInternalServerError)
		return fiberContext.JSON(fiber.Map{"error": err.Error()})
	}

	fiberContext.Status(fiber.StatusOK)
	return fiberContext.JSON(fiber.Map{"message": "done"})
}
