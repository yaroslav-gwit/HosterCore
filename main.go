//go:build freebsd
// +build freebsd

package main

// This file is only used as an entrypoint into the app
import (
	"HosterCore/cmd"
)

func main() {
	cmd.Execute()
}
