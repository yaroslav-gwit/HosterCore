package main

import (
	"HosterCore/cmd"
	"encoding/json"
	"fmt"
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

const LOG_SEPARATOR = " || "

var tagCustomError string
var restApiConfig cmd.RestApiConfig

func main() {
	defer func() {
		if r := recover(); r != nil {
			errorValue := fmt.Sprintf("%s", r)
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PANIC AVOIDED: main(): "+errorValue).Run()
		}
	}()

	// user := "admin"
	// password := "123456"
	// port := 3000

	// portEnv := os.Getenv("REST_API_PORT")
	// userEnv := os.Getenv("REST_API_USER")
	// passwordEnv := os.Getenv("REST_API_PASSWORD")

	var err error

	app := fiber.New(fiber.Config{DisableStartupMessage: true, Prefork: false})
	// app.Use(recover.New())
	app.Use(requestid.New())

	// Custom File Writer
	logFile, err := os.OpenFile("/var/log/hoster_rest_api.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// file, err := os.OpenFile("/var/run/hoster_rest_api.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer logFile.Close()

	tagCustomError = ""
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
		Format: "${custom_tag_time}" + LOG_SEPARATOR + "${ip}:${port}" + LOG_SEPARATOR + "${status}" + LOG_SEPARATOR + "${method}" + LOG_SEPARATOR + "${path}" + LOG_SEPARATOR + "${latency}" + LOG_SEPARATOR + "bytesSent: ${bytesSent}${custom_tag_err}\n",
		Output: logFile,
	}))

	authMap := make(map[string]string)
	for _, auth := range restApiConfig.HTTPAuth {
		authMap[auth.User] = auth.Password
	}
	app.Use(basicauth.New(basicauth.Config{
		Users: authMap}))

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
		err := cmd.VmStop(vm.Name, false, false)
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
		err := cmd.VmStop(vm.Name, true, false)
		if err != nil {
			tagCustomError = err.Error()
			fiberContext.Status(fiber.StatusBadRequest)
			return fiberContext.JSON(fiber.Map{"error": err.Error()})

		}
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "success"})
	})

	app.Post("/vm/stop-all", func(fiberContext *fiber.Ctx) error {
		go cmd.VmStopAll(false, false)
		fiberContext.Status(fiber.StatusOK)
		return fiberContext.JSON(fiber.Map{"message": "process started"})
	})

	app.Post("/vm/stop-all-force", func(fiberContext *fiber.Ctx) error {
		go cmd.VmStopAll(true, false)
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

	api := app.Group("/api")
	v1 := api.Group("/v1")

	v1.Post("/vm/start", handleVmStart)
	v1.Post("/vm/change-parent", handleVmChangeParent)
	v1.Post("/vm/cireset", handleVmCiReset)

	if restApiConfig.HaMode {
		ha := v1.Group("/ha")
		ha.Post("/register", handleHaRegistration)
		ha.Get("/all-members", handleHaShareAllMembers)
		ha.Post("/terminate", handleHaTerminate)
		ha.Post("/ping", handleHaPing)
		ha.Get("/monitor", handleHaMonitor)
		ha.Get("/vms", handleHaVmsList)
	}

	timesFailed := 0
	timesFailedMax := 3
	hosterRestLabel := "HOSTER_REST"
	if restApiConfig.HaMode {
		if restApiConfig.HaDebug {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "DEBUG: hoster_rest_api started in DEBUG mode").Run()
		} else {
			_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PROD: hoster_rest_api started in PRODUCTION mode").Run()
		}

		hosterRestLabel = "HOSTER_HA_REST"
		_ = exec.Command("logger", "-t", hosterRestLabel, "DEBUG: hoster_rest_api service start-up").Run()

		// Execute ha_watchdog using nohup and disown it
		if restApiConfig.HaDebug {
			os.Setenv("REST_API_HA_DEBUG", "true")
		}
		haWatchdogCmd := exec.Command("nohup", "/opt/hoster-core/ha_watchdog", "&")
		haWatchdogCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		err = haWatchdogCmd.Start()

		go func() {
			for {
				time.Sleep(time.Second * 4)

				out, err := exec.Command("pgrep", "ha_watchdog").CombinedOutput()
				if err != nil {
					_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "ERROR: ha_watchdog process is not running").Run()
					timesFailed += 1
				} else {
					_ = exec.Command("kill", "-SIGHUP", strings.TrimSpace(string(out))).Run()
					timesFailed = 0
				}

				if timesFailed >= timesFailedMax {
					if restApiConfig.HaDebug {
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "DEBUG: process will exit due to HA_WATCHDOG failure").Run()
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "DEBUG: the host system shall reboot soon").Run()
						os.Exit(1)
					} else {
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PROD: process will exit due to HA_WATCHDOG failure").Run()
						cmd.LockAllVms()
						_ = exec.Command("logger", "-t", "HOSTER_HA_REST", "PROD: the host system shall reboot soon").Run()
						_ = exec.Command("reboot").Run()
					}
				}
			}
		}()
	}

	signals := make(chan os.Signal, 1)
	// SIGTERM -> kill -s SIGTERM  ||  SIGINT -> CTRL+C
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for sig := range signals {
			if sig == syscall.SIGTERM || sig == syscall.SIGINT {
				if sig == syscall.SIGTERM {
					_ = exec.Command("logger", "-t", hosterRestLabel, "INFO: received SIGTERM, exiting").Start()
				}

				if sig == syscall.SIGINT {
					_ = exec.Command("logger", "-t", hosterRestLabel, "INFO: received SIGINT (CTRL+C // graceful HA stop), exiting").Start()

					if restApiConfig.HaMode {
						terminateOtherMembers()
					}
				}

				err := app.Shutdown()
				if err != nil {
					_ = exec.Command("logger", "-t", hosterRestLabel, "ERROR: could not shutdown the server gracefully: "+err.Error()).Run()
				}
			}
		}
	}()

	err = app.Listen(restApiConfig.Bind + ":" + strconv.Itoa(restApiConfig.Port))
	if err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}

	os.Exit(0)
}
