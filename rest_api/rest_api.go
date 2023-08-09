package main

import (
	"encoding/json"
	"hoster/cmd"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// Light yellow-ish coloured log separator
const LOG_SEPARATOR = " || "

func main() {
	user := "admin"
	password := "123456"
	port := 3000
	haMode := false

	portEnv := os.Getenv("REST_API_PORT")
	userEnv := os.Getenv("REST_API_USER")
	passwordEnv := os.Getenv("REST_API_PASSWORD")
	haModeEnv := os.Getenv("REST_API_HA_MODE")

	var err error
	if len(portEnv) > 0 {
		port, err = strconv.Atoi(portEnv)
		if err != nil {
			log.Fatal("please make sure port is an integer!")
		}
	}
	if len(userEnv) > 0 {
		user = userEnv
	}
	if len(passwordEnv) > 0 {
		password = passwordEnv
	}
	if len(haModeEnv) > 0 {
		haMode = true
	}

	app := fiber.New(fiber.Config{DisableStartupMessage: true, Prefork: false})
	app.Use(requestid.New())

	// Custom File Writer
	file, err := os.OpenFile("/var/run/hoster_rest_api.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()

	tagCustomError := ""
	app.Use(logger.New(logger.Config{
		CustomTags: map[string]logger.LogFunc{
			"custom_tag_err": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(
					func() string {
						if len(tagCustomError) > 0 {
							return LOG_SEPARATOR + "Error: " + tagCustomError
						}
						return ""
					}(),
				)
			},
			"custom_tag_time": func(output logger.Buffer, c *fiber.Ctx, data *logger.Data, extraParam string) (int, error) {
				return output.WriteString(
					func() string {
						return time.Now().Format("2006-01-02_15-04-05")
					}(),
				)
			},
		},
		Format: "Date: ${custom_tag_time}" + LOG_SEPARATOR + "Source: ${ip}:${port}" + LOG_SEPARATOR + "Status: ${status}" + LOG_SEPARATOR + "Method: ${method}" + LOG_SEPARATOR + "Path: ${path}" + LOG_SEPARATOR + "ExecTime: ${latency}" + LOG_SEPARATOR + "BytesSent: ${bytesSent}${custom_tag_err}\n",
		Output: file,
	}))

	app.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			user: password,
		},
	}))

	app.Get("/host/info", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		result := cmd.JsonOutputHostInfo()
		jsonResult, err := json.Marshal(result)
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	app.Get("/vm/list", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		result := cmd.GetAllVms()
		jsonResult, err := json.Marshal(result)
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	app.Get("/vm/info-all", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		allVms := cmd.GetAllVms()
		result := []cmd.VmInfoStruct{}
		for _, v := range allVms {
			tempRes, _ := cmd.GetVmInfo(v, true)
			result = append(result, tempRes)
		}
		jsonResult, err := json.Marshal(result)
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	type vmName struct {
		Name string `json:"name" xml:"name" form:"name"`
	}
	app.Post("/vm/info", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		result, err := cmd.GetVmInfo(vm.Name, false)
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		jsonResult, err := json.Marshal(result)
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.SendString(string(jsonResult))
	})

	app.Post("/vm/destroy", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}

		err := cmd.VmDestroy(vm.Name)
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}

		err = cmd.ReloadDnsServer()
		if err != nil {
			log.Fatal("Could not reload the DNS server: " + err.Error())
		}

		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	app.Post("/vm/start", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		// Using NOHUP option in order to avoid killing the VMs process when API server stops
		execPath, err := os.Executable()
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.SendString(`{ "message": "failed to start the process"}`)
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
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	app.Post("/vm/start-all", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		// Using NOHUP option in order to avoid killing all VMs when API server stops
		execPath, err := os.Executable()
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"message": "failed to start the process"})
		}
		execFile := path.Dir(execPath) + "/hoster"
		// Execute start all from the terminal using nohup
		cmd := exec.Command("nohup", execFile, "vm", "start-all", "&")
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
				log.Println(err)
			}
		}()

		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "process started"})
	})

	app.Post("/vm/stop", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		err := cmd.VmStop(vm.Name, false)
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusBadRequest)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})

		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	app.Post("/vm/stop-force", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		err := cmd.VmStop(vm.Name, true)
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusBadRequest)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})

		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	app.Post("/vm/stop-all", func(fiberContext *fiber.Ctx) error {
		go cmd.VmStopAll(false)
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "process started"})
	})

	app.Post("/vm/stop-all-force", func(fiberContext *fiber.Ctx) error {
		go cmd.VmStopAll(true)
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "process started"})
	})

	type vmSnap struct {
		VmName          string `json:"vm_name" xml:"name" form:"name"`
		SnapshotName    string `json:"s_name" xml:"s_name" form:"s_name"`
		SnapshotType    string `json:"s_type" xml:"s_type" form:"s_type"`
		SnapshotsToKeep int    `json:"s_to_keep" xml:"s_to_keep" form:"s_to_keep"`
	}

	app.Post("/vm/snapshot/new", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		vmSnapVar := new(vmSnap)
		if err := fiberContext.BodyParser(vmSnapVar); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		go cmd.VmZfsSnapshot(vmSnapVar.VmName, vmSnapVar.SnapshotType, vmSnapVar.SnapshotsToKeep)
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	app.Post("/vm/snapshot/destroy", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		vmSnapVar := new(vmSnap)
		if err := fiberContext.BodyParser(vmSnapVar); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		err = cmd.ZfsSnapshotDestroy(vmSnapVar.VmName, vmSnapVar.SnapshotName)
		if err := fiberContext.BodyParser(vmSnapVar); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	app.Post("/vm/snapshot/rollback", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		vmSnapVar := new(vmSnap)
		if err := fiberContext.BodyParser(vmSnapVar); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		err = cmd.ZfsSnapshotRollback(vmSnapVar.VmName, vmSnapVar.SnapshotName, false, false)
		if err := fiberContext.BodyParser(vmSnapVar); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	type vmAllSnap struct {
		SnapshotType    string `json:"s_type" xml:"s_type" form:"s_type"`
		SnapshotsToKeep int    `json:"s_to_keep" xml:"s_to_keep" form:"s_to_keep"`
	}
	app.Post("/vm/snapshot/all", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		vmSnapVar := new(vmAllSnap)
		if err := fiberContext.BodyParser(vmSnapVar); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		go func() {
			for _, vm := range cmd.GetAllVms() {
				if cmd.VmLiveCheck(vm) {
					cmd.VmZfsSnapshot(vm, vmSnapVar.SnapshotType, vmSnapVar.SnapshotsToKeep)
				}
			}
		}()
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	app.Post("/vm/snapshot-list", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		vm := new(vmName)
		if err := fiberContext.BodyParser(vm); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		snapInfo, err := cmd.GetSnapshotInfo(vm.Name, false)
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusBadRequest)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(snapInfo)
	})

	type diskExpand struct {
		VmName        string `json:"name" xml:"name" form:"name"`
		DiskImage     string `json:"disk_image" xml:"disk_image" form:"disk_image"`
		ExpansionSize int    `json:"expansion_size" xml:"expansion_size" form:"expansion_size"`
	}
	app.Post("/vm/disk-expand", func(fiberContext *fiber.Ctx) error {
		tagCustomError = ""
		diskExpandVar := new(diskExpand)
		if err := fiberContext.BodyParser(diskExpandVar); err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusUnprocessableEntity)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}

		err := cmd.DiskExpandOffline(diskExpandVar.VmName, diskExpandVar.DiskImage, diskExpandVar.ExpansionSize)
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusInternalServerError)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})
		}

		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})

	})

	timesFailed := 0
	timesFailedMax := 5
	if haMode {
		exec.Command("nohup", "/opt/hoster-core/ha_watchdog", "&").Run()
		go func() {
			for {
				out, err := exec.Command("pgrep", "ha_watchdog").CombinedOutput()
				if err != nil {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "process is not running: HA_WATCHDOG").Run()
					timesFailed += 1
				} else {
					pid := strings.TrimSpace(string(out))
					out1, err1 := exec.Command("kill", "-SIGHUP", pid).CombinedOutput()
					if err1 != nil {
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", string(out1)).Run()
					}
					timesFailed = 0
				}

				if timesFailed >= timesFailedMax {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "process exited due to HA_WATCHDOG failure").Run()
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "the host system shall reboot soon").Run()
					os.Exit(1)
				}

				time.Sleep(time.Second * 4)
			}
		}()
	}

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

	err = app.Listen("0.0.0.0:" + strconv.Itoa(port))
	if err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}
	os.Exit(0)
}
