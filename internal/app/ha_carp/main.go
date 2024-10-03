package main

import (
	"fmt"
	"os"
)

func main() {
	out, err := parseIfconfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(out)
}
