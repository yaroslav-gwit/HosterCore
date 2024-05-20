package HosterHostUtils

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func UnlockEncryptedDataset(dataset string, password string) (e error) {
	// out, err := exec.Command("zfs", "load-key", dataset).CombinedOutput()
	// if err != nil {
	// 	e = fmt.Errorf("could not load the key for the dataset: %s", strings.TrimSpace(string(out)))
	// 	return
	// }

	// cmd := exec.Command("zfs", "mount", "-l", "tank/vm-encrypted")
	cmd := exec.Command("zfs", "load-key", dataset)

	// Create a pipe to the command's stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		e = fmt.Errorf("error creating stdin pipe: %s", err.Error())
		return
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	// Start the command
	if err := cmd.Start(); err != nil {
		e = fmt.Errorf("could not start the load-key command: %s", err.Error())
		return
	}
	// Write the password to stdin and close it
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, password+"\n")
	}()
	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		e = fmt.Errorf("could not load the encryption key: %s; %s", stderr.String(), err.Error())
		return
	}

	out, err := exec.Command("zfs", "mount", dataset).CombinedOutput()
	if err != nil {
		e = fmt.Errorf("could not mount dataset: %s", strings.TrimSpace(string(out)))
		return
	}

	return
}
