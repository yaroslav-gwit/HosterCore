package cmd

import (
	"HosterCore/emojlog"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
				log.Fatal(err.Error())
			}
			// cmd.Help()
			err = downloadJailArchives(jailBootstrapCmdOsRelease, jailBootstrapCmdDataset, jailBootstrapCmdExcludeLib32)
			if err != nil {
				emojlog.PrintLogMessage(err.Error(), emojlog.Error)
				os.Exit(1)
			}
		},
	}
)

// Downloads a FreeBSD release specified, to be used as a Jail base.
//
// Returns an `error` if something goes wrong. Starts and reports the download progress otherwise.
func downloadJailArchives(release string, dataset string, excludeLib32 bool) error {
	// FreeBSD Mirror to get the archives from
	// https://download.freebsd.org/releases/amd64/

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
	fsFileLocation := fmt.Sprintf("/tmp/%s", archiveName)

	res, err = http.Get(fmt.Sprintf("%s/base.txz", requestUrl))
	if err != nil {
		return err
	}

	_ = os.Remove(fsFileLocation)
	f, err := os.OpenFile(fsFileLocation, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	bar := progressbar.DefaultBytes(
		res.ContentLength,
		fmt.Sprintf(" 📥 Downloading an archive || %s || %s", archiveName, fsFileLocation),
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
	if excludeLib32 {
		// skip lib32 download
		_ = 0
	} else {
		archiveName = "lib32.txz"
		fsFileLocation = fmt.Sprintf("/tmp/%s", archiveName)

		res, err = http.Get(fmt.Sprintf("%s/base.txz", requestUrl))
		if err != nil {
			return err
		}

		_ = os.Remove(fsFileLocation)
		f, err = os.OpenFile(fsFileLocation, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}

		bar = progressbar.DefaultBytes(
			res.ContentLength,
			fmt.Sprintf(" 📥 Downloading an archive || %s || %s", archiveName, fsFileLocation),
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

	return nil
}
