package cmd

import (
	"HosterCore/emojlog"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

func downloadJailArchives(release string, dataset string, excludeLib32 bool) error {
	// FreeBSD Mirror to get the archives from
	// https://download.freebsd.org/releases/amd64/
	requestUrl := fmt.Sprintf("https://download.freebsd.org/releases/amd64/%s", release)
	res, err := http.Get(requestUrl)
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Body: %s", string(body))
	fmt.Println()
	fmt.Printf("Response code: %v", res.Status)
	fmt.Println()

	err = res.Body.Close()
	if err != nil {
		return err
	}

	return nil
}

// func configReleaseString(release string) (releaseStringCorrect bool) {
// 	correctReleases

// 	return
// }
