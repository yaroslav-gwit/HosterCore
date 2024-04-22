// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterJail

import (
	"HosterCore/internal/pkg/byteconversion"
	FreeBSDOsInfo "HosterCore/internal/pkg/freebsd/info"
	HosterHost "HosterCore/internal/pkg/hoster/host"
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// Downloads a FreeBSD release specified, to be used as a Jail base,
// creates a new ZFS dataset (eg: zroot/vm-encrypted/jail-template-13.2-RELEASE),
// and extracts the downloaded files there.
//
// Returns an `error` if something goes wrong. Starts and reports the download progress otherwise.
func BootstrapOfficial(release string, dataset string, excludeLib32 bool) error {
	// If the logger was already set, ignore this
	if !log.ConfigSet {
		log.SetFileLocation(HosterJailUtils.JAIL_AUDIT_LOG_LOCATION)
	}

	// FreeBSD Mirror to get the archives from
	// https://download.freebsd.org/releases/amd64/
	var err error

	if len(release) < 1 {
		release, err = FreeBSDOsInfo.GetMajorReleaseVersion()
		if err != nil {
			return err
		}
	}

	hostConf, err := HosterHost.GetHostConfig()
	if err != nil {
		return err
	}

	if len(dataset) < 1 {
		dataset = hostConf.ActiveZfsDatasets[0]
	} else {
		if !slices.Contains(hostConf.ActiveZfsDatasets, dataset) {
			return errors.New("dataset you wanted to use doesn't exist: " + dataset)
		}
	}

	// Create a new, nested ZFS dataset for our Jail template
	out, err := exec.Command("zfs", "create", fmt.Sprintf("%s/jail-template-%s", dataset, release)).CombinedOutput()
	if err != nil {
		return fmt.Errorf("error: %s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	// Check if the release exists block
	requestUrl := fmt.Sprintf("https://download.freebsd.org/releases/amd64/%s", release)
	res, err := http.Get(requestUrl)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode == 200 && strings.Contains(string(body), "base.txz") {
		_ = 0
	} else {
		return fmt.Errorf("release was not found on the FreeBSD website (here): %s", requestUrl)
	}

	err = res.Body.Close()
	if err != nil {
		return err
	}
	// EOF Check if the release exists block

	// TO DO: create a separate, generic archive download function to deduplicate the code below
	// Download base.txz
	archiveName := "base.txz"
	baseFileLocation := fmt.Sprintf("/tmp/%s/%s", release, archiveName)

	res, err = http.Get(fmt.Sprintf("%s/%s", requestUrl, archiveName))
	if err != nil {
		return err
	}

	_ = os.Mkdir(fmt.Sprintf("/tmp/%s", release), 0600)
	_ = os.Remove(baseFileLocation)
	f, err := os.OpenFile(baseFileLocation, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	bar := progressbar.DefaultBytes(
		res.ContentLength,
		fmt.Sprintf(" 📥 Downloading an archive || %s || %s", archiveName, baseFileLocation),
	)

	bytesWritten, err := io.Copy(io.MultiWriter(f, bar), res.Body)
	if err != nil {
		return err
	} else {
		message := fmt.Sprintf("%s has been downloaded (%s)", archiveName, byteconversion.BytesToHuman(uint64(bytesWritten)))
		log.Info(message)
	}

	err = res.Body.Close()
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}
	// EOF Download base.txz

	// Download lib32.txz
	var lib32FileLocation string
	if excludeLib32 {
		// skip lib32 download
		_ = 0
	} else {
		archiveName = "lib32.txz"
		lib32FileLocation = fmt.Sprintf("/tmp/%s/%s", release, archiveName)

		res, err = http.Get(fmt.Sprintf("%s/%s", requestUrl, archiveName))
		if err != nil {
			return err
		}

		_ = os.Remove(lib32FileLocation)
		f, err = os.OpenFile(lib32FileLocation, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}

		bar = progressbar.DefaultBytes(
			res.ContentLength,
			fmt.Sprintf(" 📥 Downloading an archive || %s || %s", archiveName, lib32FileLocation),
		)

		bytesWritten, err = io.Copy(io.MultiWriter(f, bar), res.Body)
		if err != nil {
			return err
		} else {
			message := fmt.Sprintf("%s has been downloaded (%s)", archiveName, byteconversion.BytesToHuman(uint64(bytesWritten)))
			log.Info(message)
		}

		err = res.Body.Close()
		if err != nil {
			return err
		}

		err = f.Close()
		if err != nil {
			return err
		}
	}
	// EOF Download lib32.txz

	jailTemplateFolder := fmt.Sprintf("/%s/jail-template-%s", dataset, release)
	jailTemplateRootPath := fmt.Sprintf("%s/root_folder", jailTemplateFolder)

	_ = os.RemoveAll(jailTemplateRootPath)
	err = os.Mkdir(jailTemplateRootPath, 0755)
	if err != nil {
		return err
	}

	log.Info(fmt.Sprintf("Extracting %s to %s", baseFileLocation, jailTemplateRootPath))
	err = extractTxz(baseFileLocation, jailTemplateRootPath)
	if err != nil {
		return err
	}
	log.Info(fmt.Sprintf("%s has been extracted", baseFileLocation))

	if !excludeLib32 {
		log.Info(fmt.Sprintf("Extracting %s to %s", lib32FileLocation, jailTemplateRootPath))
		err = extractTxz(lib32FileLocation, jailTemplateRootPath)
		if err != nil {
			return err
		}
		log.Info(fmt.Sprintf("%s has been extracted", lib32FileLocation))
	}

	log.Info(fmt.Sprintf("A new Jail template has been bootstrapped (%s) at %s", release, jailTemplateFolder))
	return nil
}

// Checks if the ZFS dataset exists, and returns true or false.
// Takes in a full dataset path (ZFS path, not a mount path)
// TBD to integrate
//
// func doesDatasetExist(dataset string) (bool, error) {
// 	out, err := exec.Command("zfs", "list", "-Hp").CombinedOutput()
// 	if err != nil {
// 		return false, fmt.Errorf("could not run `zfs list`: %s; %s", strings.TrimSpace(string(out)), err.Error())
// 	}
// 	reSplitAtSpace := regexp.MustCompile(`\s+`)
// 	for _, v := range strings.Split(string(out), "\n") {
// 		if len(v) < 1 {
// 			continue
// 		}
// 		v = reSplitAtSpace.Split(v, -1)[0]
// 		if dataset == v {
// 			return true, nil
// 		}
// 	}
// 	return false, nil
// }

func extractTxz(archivePath, rootFolder string) error {
	out, err := exec.Command("tar", "-xf", archivePath, "-C", rootFolder, "--unlink").CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not extract the archive %s: %s; %s", archivePath, strings.TrimSpace(string(out)), err.Error())
	}

	return nil
}
