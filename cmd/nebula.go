package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"hoster/emojlog"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	nebulaCmd = &cobra.Command{
		Use:   "nebula",
		Short: "Nebula network service manager",
		Long:  `Nebula network service manager`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

var (
	nebulaInitCmd = &cobra.Command{
		Use:   "init",
		Short: "Bootstrap Nebula on this node",
		Long:  `Bootstrap Nebula on this node (requires valid Nebula JSON config file)`,
		Run: func(cmd *cobra.Command, args []string) {
			err := nebulaBootstrap()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

var (
	nebulaShowLogCmd = &cobra.Command{
		Use:   "show-log",
		Short: "Use `tail -f` to display Nebula's live log",
		Long:  `Use "tail -f" to display Nebula's live log`,
		Run: func(cmd *cobra.Command, args []string) {
			err := tailNebulaLogFile()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

var (
	nebulaServiceStart  bool
	nebulaServiceStop   bool
	nebulaServiceReload bool

	nebulaServiceCmd = &cobra.Command{
		Use:   "service",
		Short: "Start, stop, or reload Nebula process",
		Long:  `Start, stop, or reload Nebula process`,
		Run: func(cmd *cobra.Command, args []string) {
			if nebulaServiceReload {
				err := reloadNebulaService()
				if err != nil {
					log.Fatal(err)
				}
			} else if nebulaServiceStart {
				err := startNebulaService()
				if err != nil {
					log.Fatal(err)
				}
			} else if nebulaServiceStop {
				err := stopNebulaService()
				if err != nil {
					log.Fatal(err)
				}
			} else {
				cmd.Help()
			}
		},
	}
)

var (
	nebulaUpdateBinary bool
	nebulaUpdateConfig bool

	nebulaUpdateCmd = &cobra.Command{
		Use:   "update",
		Short: "Download the latest changes from Nebula Control Plane API server",
		Long:  `Download the latest changes from Nebula Control Plane API server`,
		Run: func(cmd *cobra.Command, args []string) {
			if nebulaUpdateConfig {
				err := downloadNebulaConfig()
				if err != nil {
					log.Fatal(err)
				}
				err = downloadNebulaCerts()
				if err != nil {
					log.Fatal(err)
				}
			} else if nebulaUpdateBinary {
				err := downloadNebulaBin()
				if err != nil {
					log.Fatal(err)
				}
			} else {
				cmd.Help()
			}
		},
	}
)

// const nebulaServiceFolder = "/opt/nebula_new/"
const nebulaServiceFolder = "/opt/nebula/"

func startNebulaService() error {
	reMatchLocation := regexp.MustCompile(`.*` + nebulaServiceFolder + `nebula.*`)
	reMatchSpace := regexp.MustCompile(`\s+`)
	pgrepOut, _ := exec.Command("pgrep", "-lf", "nebula").CombinedOutput()

	nebulaPid := ""
	for _, v := range strings.Split(string(pgrepOut), "\n") {
		if reMatchLocation.MatchString(v) {
			nebulaPid = reMatchSpace.Split(v, -1)[0]
		}
	}

	if len(nebulaPid) > 0 {
		return errors.New("service process for Nebula is already running")
	}

	const nebulaStartSh = "(( nohup " + nebulaServiceFolder + "nebula -config " + nebulaServiceFolder + "config.yml 1>>" + nebulaServiceFolder + "log.txt 2>&1 )&)"
	const nebulaStartShLocation = "/tmp/nebula.sh"
	// Open nebulaStartShLocation for writing
	nebulaStartShFile, err := os.Create(nebulaStartShLocation)
	if err != nil {
		return err
	}
	defer nebulaStartShFile.Close()
	// Create a new writer
	writer := bufio.NewWriter(nebulaStartShFile)
	// Write a string to the file
	_, err = writer.WriteString(nebulaStartSh)
	if err != nil {
		return err
	}
	// Flush the writer to ensure all data has been written to the file
	err = writer.Flush()
	if err != nil {
		return err
	}
	err = os.Chmod(nebulaStartShLocation, os.FileMode(0600))
	if err != nil {
		return errors.New("error changing permissions: " + err.Error())
	}

	nebulaStartErr := exec.Command("sh", nebulaStartShLocation).Start()
	if err != nil {
		return nebulaStartErr
	}
	emojlog.PrintLogMessage("Started new Nebula process", emojlog.Debug)

	return nil
}

func stopNebulaService() error {
	reMatchLocation := regexp.MustCompile(`.*` + nebulaServiceFolder + `nebula.*`)
	reMatchSpace := regexp.MustCompile(`\s+`)
	pgrepOut, _ := exec.Command("pgrep", "-lf", "nebula").CombinedOutput()

	nebulaPid := ""
	for _, v := range strings.Split(string(pgrepOut), "\n") {
		if reMatchLocation.MatchString(v) {
			nebulaPid = reMatchSpace.Split(v, -1)[0]
		}
	}

	if len(nebulaPid) < 1 {
		emojlog.PrintLogMessage("Nebula service is already dead: ", emojlog.Error)
		return errors.New("service is already dead")
	}

	killOut, err := exec.Command("kill", "-SIGTERM", nebulaPid).CombinedOutput()
	if err != nil {
		return errors.New(string(killOut))
	}
	emojlog.PrintLogMessage("Stopped Nebula service using it's pid: "+nebulaPid, emojlog.Debug)

	return nil
}

func reloadNebulaService() error {
	reMatchLocation := regexp.MustCompile(`.*` + nebulaServiceFolder + `nebula.*`)
	reMatchSpace := regexp.MustCompile(`\s+`)
	pgrepOut, _ := exec.Command("pgrep", "-lf", "nebula").CombinedOutput()

	nebulaPid := ""
	for _, v := range strings.Split(string(pgrepOut), "\n") {
		if reMatchLocation.MatchString(v) {
			nebulaPid = reMatchSpace.Split(v, -1)[0]
		}
	}

	if len(nebulaPid) > 0 {
		const nebulaStartSh = "(( nohup " + nebulaServiceFolder + "nebula -config " + nebulaServiceFolder + "config.yml 1>>" + nebulaServiceFolder + "log.txt 2>&1 )&)"
		const nebulaStartShLocation = "/tmp/nebula.sh"
		// Open nebulaStartShLocation for writing
		nebulaStartShFile, err := os.Create(nebulaStartShLocation)
		if err != nil {
			return err
		}
		defer nebulaStartShFile.Close()
		// Create a new writer
		writer := bufio.NewWriter(nebulaStartShFile)
		// Write a string to the file
		_, err = writer.WriteString(nebulaStartSh)
		if err != nil {
			return err
		}
		// Flush the writer to ensure all data has been written to the file
		err = writer.Flush()
		if err != nil {
			return err
		}
		err = os.Chmod(nebulaStartShLocation, os.FileMode(0600))
		if err != nil {
			return errors.New("error changing permissions: " + err.Error())
		}

		killOut, err := exec.Command("kill", "-SIGTERM", nebulaPid).CombinedOutput()
		if err != nil {
			return errors.New(string(killOut))
		}
		emojlog.PrintLogMessage("Stopped Nebula service using it's pid: "+nebulaPid, emojlog.Debug)
		nebulaStartErr := exec.Command("sh", nebulaStartShLocation).Start()
		if err != nil {
			return nebulaStartErr
		}
		emojlog.PrintLogMessage("Started new Nebula process", emojlog.Debug)
	} else {
		emojlog.PrintLogMessage("Service is not running", emojlog.Warning)
	}

	return nil
}

func tailNebulaLogFile() error {
	tailCmd := exec.Command("tail", "-35", "-f", nebulaServiceFolder+"log.txt")
	tailCmd.Stdin = os.Stdin
	tailCmd.Stdout = os.Stdout
	tailCmd.Stderr = os.Stderr

	err := tailCmd.Run()
	if err != nil {
		return err
	}

	return nil
}

type NebulaClusterConfig struct {
	ClusterName   string `json:"cluster_name"`
	ClusterID     string `json:"cluster_id"`
	HostID        string `json:"host_id"`
	NatPunch      string `json:"nat_punch"`
	ListenAddress string `json:"listen_address"`
	ListenPort    string `json:"listen_port"`
	MTU           string `json:"mtu"`
	UseRelays     string `json:"use_relays"`
	APIServer     string `json:"api_server"`
}

func readNebulaClusterConfig() (NebulaClusterConfig, error) {
	execPath, err := os.Executable()
	if err != nil {
		return NebulaClusterConfig{}, err
	}
	nebulaClusterConfigFile := path.Dir(execPath) + "/config_files/nebula_config.json"
	// read the json file from disk
	data, err := os.ReadFile(nebulaClusterConfigFile)
	if err != nil {
		return NebulaClusterConfig{}, err
	}

	// unmarshal the json data into a Config struct
	var config NebulaClusterConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return NebulaClusterConfig{}, err
	}

	return config, nil
}

func downloadNebulaConfig() error {
	c, err := readNebulaClusterConfig()
	if err != nil {
		return err
	}

	emojlog.PrintLogMessage("Initiating Nebula config file download", emojlog.Debug)
	req, err := http.NewRequest("GET", "https://"+c.APIServer+"/get_config?cluster_name="+c.ClusterName+"&cluster_id="+c.ClusterID+"&host_name="+GetHostName()+"&host_id="+c.HostID+"&nat_punch="+c.NatPunch+"&listen_host="+c.ListenAddress+"&listen_port="+c.ListenPort+"&mtu="+c.MTU+"&use_relays="+c.UseRelays, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	reMatch := regexp.MustCompile(`.*/opt/nebula/ca.crt.*`)
	configFileValid := false
	for _, v := range strings.Split(string(body), "\n") {
		if reMatch.MatchString(v) {
			configFileValid = true
			break
		}
	}
	if !configFileValid {
		return errors.New("please check your settings: " + string(body))
	}

	// Open nebulaStartShLocation for writing
	nebulaConfFile, err := os.Create(nebulaServiceFolder + "config.yml")
	if err != nil {
		return err
	}
	defer nebulaConfFile.Close()
	// Create a new writer
	writer := bufio.NewWriter(nebulaConfFile)
	// Write a string to the file
	_, err = writer.WriteString(string(body))
	if err != nil {
		return err
	}
	// Flush the writer to ensure all data has been written to the file
	err = writer.Flush()
	if err != nil {
		return err
	}
	err = os.Chmod(nebulaServiceFolder+"config.yml", os.FileMode(0600))
	if err != nil {
		return errors.New("error changing permissions: " + err.Error())
	}

	emojlog.PrintLogMessage("Nebula config file download is now done", emojlog.Info)
	return nil
}

func downloadNebulaCerts() error {
	c, err := readNebulaClusterConfig()
	if err != nil {
		return err
	}

	emojlog.PrintLogMessage("Initiating Nebula certificates download", emojlog.Debug)
	req, err := http.NewRequest("GET", "https://"+c.APIServer+"/get_certs?cluster_name="+c.ClusterName+"&cluster_id="+c.ClusterID+"&host_name="+GetHostName()+"&host_id="+c.HostID, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	reMatch := regexp.MustCompile(`.*BEGIN NEBULA CERTIFICATE.*`)
	certShValid := false
	for _, v := range strings.Split(string(body), "\n") {
		if reMatch.MatchString(v) {
			certShValid = true
			break
		}
	}
	if !certShValid {
		return errors.New("please check your settings: " + string(body))
	}

	// Open nebulaStartShLocation for writing
	nebulaCertsFile, err := os.Create("/tmp/nebula_certs.sh")
	if err != nil {
		return err
	}
	defer nebulaCertsFile.Close()
	// Create a new writer
	writer := bufio.NewWriter(nebulaCertsFile)
	// Write a string to the file
	_, err = writer.WriteString(string(body))
	if err != nil {
		return err
	}
	// Flush the writer to ensure all data has been written to the file
	err = writer.Flush()
	if err != nil {
		return err
	}
	err = exec.Command("sh", "/tmp/nebula_certs.sh").Run()
	if err != nil {
		return err
	}

	_ = os.Remove("/tmp/nebula_certs.sh")

	emojlog.PrintLogMessage("Nebula certificates download is now done", emojlog.Info)
	return nil
}

func downloadNebulaBin() error {
	c, err := readNebulaClusterConfig()
	if err != nil {
		return err
	}

	emojlog.PrintLogMessage("Initiating latest compatible Nebula binary download", emojlog.Debug)
	req, err := http.NewRequest("GET", "https://"+c.APIServer+"/get_bins?os=freebsd&arch=amd64&nebula=true&service=false", nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_ = os.Remove(nebulaServiceFolder + "nebula")
	f, _ := os.OpenFile(nebulaServiceFolder+"nebula", os.O_CREATE|os.O_WRONLY, 0700)
	defer f.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		" 📥 Downloading new Nebula binary || "+nebulaServiceFolder+"nebula",
	)
	io.Copy(io.MultiWriter(f, bar), resp.Body)

	time.Sleep(time.Millisecond * 250)
	emojlog.PrintLogMessage("Latest compatible Nebula binary download is now done", emojlog.Info)
	return nil
}

func nebulaBootstrap() error {
	err := exec.Command("mkdir", "-p", nebulaServiceFolder).Run()
	if err != nil {
		return err
	}
	err = exec.Command("chmod", "700", nebulaServiceFolder).Run()
	if err != nil {
		return err
	}

	err = downloadNebulaBin()
	if err != nil {
		return err
	}

	err = downloadNebulaConfig()
	if err != nil {
		return err
	}

	err = downloadNebulaCerts()
	if err != nil {
		return err
	}

	err = startNebulaService()
	if err != nil {
		return err
	}

	return nil
}
