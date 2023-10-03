package cmd

import (
	"HosterCore/emojlog"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
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

	dsExists, err := doesDatasetExist(dataset)
	if err != nil {
		return err
	}
	if !dsExists {
		return fmt.Errorf("sorry, the dataset specified doesn't exist: %s", dataset)
	}

	err = createNestedZfsDataset(dataset, "jail-template-"+release)
	if err != nil {
		return err
	}

	if len(release) < 1 {
		release, err = getFreeBsdRelease()
		if err != nil {
			return err
		}
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

	jailTemplateFolder := fmt.Sprintf("/%s/jail-template-%s", dataset, release)
	jailTemplateRootPath := fmt.Sprintf("%s/root_folder", jailTemplateFolder)

	_ = os.RemoveAll(jailTemplateRootPath)
	err = os.Mkdir(jailTemplateRootPath, 0755)
	if err != nil {
		return err
	}

	emojlog.PrintLogMessage(fmt.Sprintf("Extracting %s to %s", baseFileLocation, jailTemplateRootPath), emojlog.Debug)
	err = extractTxz(baseFileLocation, jailTemplateRootPath)
	if err != nil {
		return err
	}
	emojlog.PrintLogMessage(fmt.Sprintf("%s has been extracted", baseFileLocation), emojlog.Changed)

	if !excludeLib32 {
		emojlog.PrintLogMessage(fmt.Sprintf("Extracting %s to %s", lib32FileLocation, jailTemplateRootPath), emojlog.Debug)
		err = extractTxz(lib32FileLocation, jailTemplateRootPath)
		if err != nil {
			return err
		}
		emojlog.PrintLogMessage(fmt.Sprintf("%s has been extracted", lib32FileLocation), emojlog.Changed)
	}

	emojlog.PrintLogMessage(fmt.Sprintf("A new Jail template has been bootstrapped (%s) at %s", release, jailTemplateFolder), emojlog.Info)
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

	dsExists, err := doesDatasetExist(parentDs + "/" + nestedDs)
	if err != nil {
		return err
	}
	if dsExists {
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
func doesDatasetExist(dataset string) (bool, error) {
	out, err := exec.Command("zfs", "list", "-Hp").CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("could not run `zfs list`: %s; %s", strings.TrimSpace(string(out)), err.Error())
	}

	reSplitAtSpace := regexp.MustCompile(`\s+`)
	for _, v := range strings.Split(string(out), "\n") {
		if len(v) < 1 {
			continue
		}
		v = reSplitAtSpace.Split(v, -1)[0]
		if dataset == v {
			return true, nil
		}
	}

	return false, nil
}

func extractTxz(archivePath, rootFolder string) error {
	out, err := exec.Command("tar", "-xf", archivePath, "-C", rootFolder, "--unlink").CombinedOutput()
	if err != nil {
		return fmt.Errorf("could not extract the archive %s: %s; %s", archivePath, strings.TrimSpace(string(out)), err.Error())
	}

	return nil
}