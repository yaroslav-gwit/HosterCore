package cmd

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"hoster/emojlog"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"time"

	"facette.io/natsort"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	imageCmd = &cobra.Command{
		Use:   "image",
		Short: "Image and template (.raw) related operations",
		Long:  `Image and template (.raw) related operations`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
)

var (
	imageOsType  string
	imageDataset string

	imageDownloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download an image from the public or private repo",
		Long:  `Download an image from the public or private repo`,
		Run: func(cmd *cobra.Command, args []string) {
			err := imageDownload(imageOsType)
			if err != nil {
				log.Fatal(err)
			}
			err = imageUnzip(imageDataset, imageOsType)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func imageUnzip(imageDataset string, imageOsType string) error {
	emojlog.PrintLogMessage("Initiating image 'unzip' process", emojlog.Info)

	// Host config read/parse
	hostConfig := HostConfig{}
	// JSON config file location
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	hostConfigFile := path.Dir(execPath) + "/config_files/host_config.json"
	// Read the JSON file
	data, err := os.ReadFile(hostConfigFile)
	if err != nil {
		return err
	}
	// Unmarshal the JSON data into a slice of Network structs
	err = json.Unmarshal(data, &hostConfig)
	if err != nil {
		return err
	}

	if !slices.Contains(hostConfig.ActiveDatasets, imageDataset) {
		return errors.New("dataset is not being used for VMs or doesn't exist")
	}

	_, err = os.Stat("/" + imageDataset + "/")
	if err != nil {
		return errors.New("dataset doesn't exist, or is not mounted")
	}
	_, err = os.Stat("/" + imageDataset + "/template-" + imageOsType)
	if err != nil {
		emojlog.PrintLogMessage("Created new image template dataset: "+imageDataset+"/template-"+imageOsType, emojlog.Debug)
		out, err := exec.Command("zfs", "create", imageDataset+"/template-"+imageOsType).CombinedOutput()
		if err != nil {
			return errors.New("could not run zfs create: " + string(out))
		}
	}

	_, diskErr := os.Stat("/" + imageDataset + "/template-" + imageOsType + "/disk0.img")
	if diskErr == nil {
		emojlog.PrintLogMessage("Removed old disk image here: /"+imageDataset+"/template-"+imageOsType+"/disk0.img", emojlog.Debug)
		_ = os.Remove("/" + imageDataset + "/template-" + imageOsType + "/disk0.img")
	}

	zipFileLocation := "/tmp/" + imageOsType + ".zip"
	r, err := zip.OpenReader(zipFileLocation)
	if err != nil {
		return err
	}
	defer r.Close()

	// Find the first file in the archive
	var file *zip.File
	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			file = f
			break
		}
	}
	if file == nil {
		return errors.New("no files found in archive")
	}

	// Create the progress bar
	bar := progressbar.NewOptions(
		int(file.FileInfo().Size()),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetDescription(" 📤 Unzipping the OS image || "+zipFileLocation+" || /"+imageDataset+"/template-"+imageOsType+"/disk0.img"),
	)

	// Open the file inside the archive
	rc, err := file.Open()
	if err != nil {
		fmt.Println("Error opening file in archive:", err)
		return err
	}
	defer rc.Close()

	// Create the output file
	// fw, err := os.Create("/tmp/disk0.img")
	fw, err := os.Create("/" + imageDataset + "/template-" + imageOsType + "/disk0.img")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return err
	}
	defer fw.Close()

	// Copy the file data and update the progress bar
	_, err = io.Copy(io.MultiWriter(fw, bar), rc)
	if err != nil {
		fmt.Println("Error copying file data:", err)
		return err
	}

	bar.Finish()
	time.Sleep(time.Millisecond * 250)
	fmt.Println()

	imageRemovalErr := os.Remove(zipFileLocation)
	if imageRemovalErr != nil {
		emojlog.PrintLogMessage("Removed previously downloaded archive: "+zipFileLocation, emojlog.Error)
		return err
	}
	emojlog.PrintLogMessage("Removed previously downloaded archive: "+zipFileLocation, emojlog.Debug)

	emojlog.PrintLogMessage("Process finished for: template-"+imageOsType, emojlog.Changed)
	return nil
}

func imageDownload(osType string) error {
	emojlog.PrintLogMessage("Initiating image download process for the OS/distribution: "+osType, emojlog.Info)
	// Host config read/parse
	hostConfig := HostConfig{}
	// JSON config file location
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	hostConfigFile := path.Dir(execPath) + "/config_files/host_config.json"
	// Read the JSON file
	data, err := os.ReadFile(hostConfigFile)
	if err != nil {
		return err
	}
	// Unmarshal the JSON data into a slice of Network structs
	err = json.Unmarshal(data, &hostConfig)
	if err != nil {
		return err
	}

	// Parse website response
	resp, err := http.Get(hostConfig.ImageServer)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var vmImageMap map[string][]map[string][]string
	err = json.Unmarshal(body, &vmImageMap)
	if err != nil {
		return err
	}
	var imageList []string
	for _, v := range vmImageMap["vm_images"] {
		for key, vv := range v {
			if key == osType {
				imageList = vv
				natsort.Sort(imageList)
			}
		}
	}
	if len(imageList) > 0 {
		vmImage := imageList[len(imageList)-1]
		vmImageFullLink := hostConfig.ImageServer + "images/" + vmImage
		req, err := http.NewRequest("GET", vmImageFullLink, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		f, err := os.OpenFile("/tmp/"+osType+".zip", os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()

		bar := progressbar.DefaultBytes(
			resp.ContentLength,
			" 📥 Downloading OS image   || "+vmImage+" || /tmp/"+osType+".zip",
		)
		io.Copy(io.MultiWriter(f, bar), resp.Body)
	} else {
		return errors.New("sorry, could not find the image")
	}

	time.Sleep(time.Millisecond * 250)
	emojlog.PrintLogMessage("Image was downloaded: /tmp/"+osType+".zip", emojlog.Changed)

	return nil
}
