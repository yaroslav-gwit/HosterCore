package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/spf13/cobra"
)

var (
	apiServerPort     int
	apiServerUser     string
	apiServerPassword string

	apiCmd = &cobra.Command{
		Use:   "api-server",
		Short: "Start an API server",
		Long:  `Start an API server on port 3000 (default).`,
		Run: func(cmd *cobra.Command, args []string) {
			StartApiServer(apiServerPort, apiServerUser, apiServerPassword)
		},
	}
)

func StartApiServer(port int, user string, password string) {
	app := fiber.New(fiber.Config{DisableStartupMessage: true, Prefork: false})
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${locals:requestid} - ${ip}]:${port} ${status} - ${method} ${path} - Error: ${error}\n"}))

	app.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			user: password,
		},
	}))

	app.Get("/host/info", func(fiberContext *fiber.Ctx) error {
		result := jsonOutputHostInfo()
		jsonResult, _ := json.Marshal(result)
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	app.Get("/vm/list", func(fiberContext *fiber.Ctx) error {
		result := getAllVms()
		jsonResult, _ := json.Marshal(result)
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	type vmName struct {
		Name string `json:"name" xml:"name" form:"name"`
	}
	app.Post("/vm/info", func(fiberContext *fiber.Ctx) error {
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		result, err := getVmInfo(vm.Name)
		if err != nil {
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		jsonResult, err := json.Marshal(result)
		if err != nil {
			log.Println(err)
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	app.Post("/vm/destroy", func(fiberContext *fiber.Ctx) error {
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}

		err := vmDestroy(vm.Name)
		if err != nil {
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		err = generateNewDnsConfig()
		if err != nil {
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		err = reloadDnsService()
		if err != nil {
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}

		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	app.Post("/vm/start", func(fiberContext *fiber.Ctx) error {
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		// Using NOHUP option in order to avoid killing the VMs process when API server stops
		execPath, err := os.Executable()
		if err != nil {
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.SendString(`{ "message": "failed to start the process"}`)
		}
		execFile := path.Dir(execPath) + "/hoster"
		// Execute start all from the terminal using nohup
		cmd := exec.Command("nohup", execFile, "vm", "start", vm.Name)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		err = cmd.Start()
		if err != nil {
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
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	app.Post("/vm/start-all", func(fiberContext *fiber.Ctx) error {
		// Using NOHUP option in order to avoid killing all VMs when API server stops
		execPath, err := os.Executable()
		if err != nil {
			return fiberContext.SendString(`{ "message": "failed to start the process"}`)
		}
		execFile := path.Dir(execPath) + "/hoster"
		// Execute start all from the terminal using nohup
		cmd := exec.Command("nohup", execFile, "vm", "start-all", "&")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		err = cmd.Start()
		if err != nil {
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"message": "failed to start the process"})
		}
		go func() {
			err := cmd.Wait()
			if err != nil {
				log.Println(err)
			}
		}()

		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "process started"})
	})

	app.Post("/vm/stop", func(fiberContext *fiber.Ctx) error {
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		err := vmStop(vm.Name)
		if err != nil {
			fiberContext.Status(fiber.StatusBadRequest)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})

		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	app.Post("/vm/stop-all", func(fiberContext *fiber.Ctx) error {
		go vmStopAll()
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "process started"})
	})

	type vmSnap struct {
		VmName          string `json:"vm_name" xml:"name" form:"name"`
		SnapshotType    string `json:"s_type" xml:"s_type" form:"s_type"`
		SnapshotsToKeep int    `json:"s_to_keep" xml:"s_to_keep" form:"s_to_keep"`
	}
	app.Post("/vm/snapshot", func(fiberContext *fiber.Ctx) error {
		vmSnapVar := new(vmSnap)
		if err := fiberContext.BodyParser(vmSnapVar); err != nil {
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		go vmZfsSnapshot(vmSnapVar.VmName, vmSnapVar.SnapshotType, vmSnapVar.SnapshotsToKeep)
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "process started"})
	})

	type vmAllSnap struct {
		SnapshotType    string `json:"s_type" xml:"s_type" form:"s_type"`
		SnapshotsToKeep int    `json:"s_to_keep" xml:"s_to_keep" form:"s_to_keep"`
	}
	app.Post("/vm/snapshot-all", func(fiberContext *fiber.Ctx) error {
		vmSnapVar := new(vmAllSnap)
		if err := fiberContext.BodyParser(vmSnapVar); err != nil {
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		go func() {
			for _, vm := range getAllVms() {
				if vmLiveCheck(vm) {
					vmZfsSnapshot(vm, vmSnapVar.SnapshotType, vmSnapVar.SnapshotsToKeep)
				}
			}
		}()
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "process started"})
	})

	type diskExpand struct {
		VmName        string `json:"name" xml:"name" form:"name"`
		DiskImage     string `json:"disk_image" xml:"disk_image" form:"disk_image"`
		ExpansionSize int    `json:"expansion_size" xml:"expansion_size" form:"expansion_size"`
	}
	app.Post("/vm/disk-expand", func(fiberContext *fiber.Ctx) error {
		diskExpandVar := new(diskExpand)
		if err := fiberContext.BodyParser(diskExpandVar); err != nil {
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}

		err := diskExpandOffline(diskExpandVar.VmName, diskExpandVar.DiskImage, diskExpandVar.ExpansionSize)
		if err != nil {
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}

		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})

	})

	// This is required to make the VMs started using NOHUP to continue running normally
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch

		// Gracefully shut down the server
		log.Println("Shutting down the API Server gracefully...")
		err := app.Shutdown()
		if err != nil {
			log.Printf("Error shutting down server: %s\n", err)
		}
	}()

	fmt.Println("")
	fmt.Println(" Use these credentials to authenticate with the API:")
	fmt.Println(" Username:", user, "|| Password:", password)
	fmt.Println(" Address: http://0.0.0.0:" + strconv.Itoa(port) + "/")
	fmt.Println("")

	err := app.Listen("0.0.0.0:" + strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}

	// app.Listen("0.0.0.0:" + strconv.Itoa(port))
	os.Exit(0)
}
