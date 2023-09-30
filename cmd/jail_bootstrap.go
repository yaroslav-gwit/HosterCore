package cmd

import (
	"HosterCore/emojlog"
	"archive/tar"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/ulikunitz/xz"
)

var (
	jailBootstrapCmdOsRelease    string
	jailBootstrapCmdDataset      string
	jailBootstrapCmdExcludeLib32 bool

	jailBootstrapCmd = &cobra.Command{
		Use:   "bootstrap",
		Short: "Bootstrap a new Jail template",
		Long:  `Bootstrap a new Jail template`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			err := checkInitFile()
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
			// cmd.Help()
			err = bootstrapJailArchives(jailBootstrapCmdOsRelease, jailBootstrapCmdDataset, jailBootstrapCmdExcludeLib32)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

// Downloads a FreeBSD release specified, to be used as a Jail base,
// creates a new ZFS dataset (eg: zroot/vm-encrypted/jail-template-13.2-RELEASE),
// and extracts the downloaded files there.
//
// Returns an `error` if something goes wrong. Starts and reports the download progress otherwise.
func bootstrapJailArchives(release string, dataset string, excludeLib32 bool) error {
	// FreeBSD Mirror to get the archives from
	// https://download.freebsd.org/releases/amd64/

	if len(dataset) < 1 {
		datasets, err := getZfsDatasetInfo()
		if err != nil {
			return err
		}
		dataset = datasets[0].Name
	}

	if !doesDatasetExist(dataset) {
		return fmt.Errorf("sorry, the dataset specified doesn't exist")
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
		emojlog.PrintLogMessage(fmt.Sprintf("%s has been downloaded (%s)", archiveName, ByteConversion(int(bytesWritten))), emojlog.Changed)
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
			emojlog.PrintLogMessage(fmt.Sprintf("%s has been downloaded (%s)", archiveName, ByteConversion(int(bytesWritten))), emojlog.Changed)
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

	err = createNestedZfsDataset(dataset, "jail-template-"+release)
	if err != nil {
		return err
	}

	jailTemplateFolder := fmt.Sprintf("/%s/jail-template-%s", dataset, release)
	jailTemplateRootPath := fmt.Sprintf("%s/root_folder", jailTemplateFolder)

	_ = os.RemoveAll(jailTemplateRootPath)
	err = os.Mkdir(jailTemplateRootPath, 0755)
	if err != nil {
		return err
	}

	err = extractTxz(baseFileLocation, jailTemplateRootPath)
	if err != nil {
		return err
	}
	if !excludeLib32 {
		err = extractTxz(lib32FileLocation, jailTemplateRootPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// Creates a new (nested) ZFS dataset, which is usually used as a template for VMs and Jails.
//
// Returns an error, if the operation was not successful.
func createNestedZfsDataset(parentDs string, nestedDs string) error {
	datasets, err := getZfsDatasetInfo()
	if err != nil {
		return err
	}

	dsFound := false
	for _, v := range datasets {
		if v.Name == parentDs {
			dsFound = true
		}
	}
	if !dsFound {
		return fmt.Errorf("parent ZFS dataset could not be found: %s", parentDs)
	}

	if doesDatasetExist(parentDs + "/" + nestedDs) {
		return fmt.Errorf("sorry, the dataset already exists: %s/%s", parentDs, nestedDs)
	}

	out, err := exec.Command("zfs", "create", parentDs+"/"+nestedDs).CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not create a new dataset: %s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	return nil
}

// Checks if the ZFS dataset exists, and returns true or false.
//
// Takes in a full dataset path (ZFS path, not a mount path)
func doesDatasetExist(dataset string) bool {
	out, err := exec.Command("zfs", "list", "-Hp").CombinedOutput()
	if err != nil {
		return false
	}

	reSplitAtSpace := regexp.MustCompile(`\+s`)
	for _, v := range strings.Split(string(out), "\n") {
		if len(v) < 1 {
			continue
		}
		v = reSplitAtSpace.Split(v, -1)[0]
		if v == dataset {
			return true
		}
	}

	return false
}

func extractTxz(archivePath, rootFolder string) error {
	xzFile, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer xzFile.Close()

	stat, _ := xzFile.Stat()
	bar := progressbar.DefaultBytes(stat.Size(), "Extracting")

	xzReader, err := xz.NewReader(xzFile)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(xzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if !strings.HasPrefix(header.Name, "/") {
			return fmt.Errorf("invalid archive path: %s", header.Name)
		}
		extractedFilePath := fmt.Sprintf("%s%s", rootFolder, header.Name)

		// Extract and apply file permissions from the header
		permissions := os.FileMode(header.Mode)
		if header.Typeflag == tar.TypeDir {
			// If it's a directory, use MkdirAll with the extracted permissions
			if err := os.MkdirAll(extractedFilePath, permissions); err != nil {
				return err
			}
		} else {
			// If it's a regular file, create the parent directory with extracted permissions
			// and create the file itself with the same permissions
			parentDir := filepath.Dir(extractedFilePath)
			if err := os.MkdirAll(parentDir, permissions); err != nil {
				return err
			}

			extractedFile, err := os.Create(extractedFilePath)
			if err != nil {
				return err
			}
			defer extractedFile.Close()

			if _, err := io.Copy(extractedFile, tarReader); err != nil {
				return err
			}
		}

		bar.Add(int(header.Size))
	}

	bar.Finish()
	return nil
}
