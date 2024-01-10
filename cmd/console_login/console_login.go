package main

import (
	"HosterCore/internal/app/console_login"
	"fmt"
	"os"
)

func main() {
	err := console_login.New()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
