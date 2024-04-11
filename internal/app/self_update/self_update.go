package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/schollz/progressbar/v3"
)

type Release struct {
	Assets []Asset `json:"assets"`
}

type Asset struct {
	BrowserDownloadURL string `json:"browser_download_url"`
}

var version = "" // automatically set during the build process

func main() {
	// Print the version and exit
	args := os.Args
	if len(args) > 1 {
		res := os.Args[1]
		if res == "version" || res == "v" || res == "--version" || res == "-v" {
			fmt.Println(version)
			return
		}
	}

	// Make a request to the GitHub API
	resp, err := http.Get("https://api.github.com/repos/yaroslav-gwit/HosterCore/releases/latest")
	if err != nil {
		fmt.Println("Error making request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Decode the response into a Release struct
	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		os.Exit(1)
	}

	// Print the browser_download_url values for each asset
	var hosterBinLink string
	var vmSupervisorBinLink string
	reMatchHoster := regexp.MustCompile(`/hoster`)
	reMatchVmSupervisor := regexp.MustCompile(`/vm_supervisor_service`)
	for _, asset := range release.Assets {
		if reMatchHoster.MatchString(asset.BrowserDownloadURL) {
			hosterBinLink = asset.BrowserDownloadURL
		} else if reMatchVmSupervisor.MatchString(asset.BrowserDownloadURL) {
			vmSupervisorBinLink = asset.BrowserDownloadURL
		}
	}

	// Download new version of hoster
	req, err := http.NewRequest("GET", hosterBinLink, nil)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		os.Exit(1)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	f, err := os.OpenFile("/opt/hoster-core/hoster", os.O_CREATE|os.O_WRONLY, 0750)
	if err != nil {
		fmt.Println("Could not replace hoster:", err)
		os.Exit(1)
	}
	defer f.Close()

	hosterBar := progressbar.DefaultBytes(
		resp.ContentLength,
		"📥 Downloading new 'hoster' binary: ",
	)
	io.Copy(io.MultiWriter(f, hosterBar), resp.Body)

	// Download new version of
	req, err = http.NewRequest("GET", vmSupervisorBinLink, nil)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		os.Exit(1)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error decoding response:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	f, err = os.OpenFile("/opt/hoster-core/vm_supervisor_service", os.O_CREATE|os.O_WRONLY, 0750)
	if err != nil {
		fmt.Println("Could not replace vm_supervisor_service:", err)
		os.Exit(1)
	}
	defer f.Close()

	vmSupervisorBar := progressbar.DefaultBytes(
		resp.ContentLength,
		"📥 Downloading new 'vm_supervisor' binary: ",
	)
	io.Copy(io.MultiWriter(f, vmSupervisorBar), resp.Body)
}
