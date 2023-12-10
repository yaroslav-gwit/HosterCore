package main

import (
	"HosterCore/utils/encryption"
	"HosterCore/utils/host"
	"fmt"
)

func main() {
	// password := "123456"
	// // hash, _ := HashPassword(password) // ignore error for the sake of simplicity
	// hash := "$2a$10$snRyHH4QrR9HCCV0/3LN/e1fwEt/Z47lfl1qFhvriZScs4JeIob/S"

	// fmt.Println("Password:", password)
	// fmt.Println("Hash:    ", hash)

	// match := CheckPasswordHash(password, hash)
	// fmt.Println("Match:   ", match)

	hostConfig, err := host.GetHostConfig()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("hostConfig:   ", hostConfig)

	pin := "123456"
	pin_hash := hostConfig.ConsolePanelPin
	match := encryption.CheckPasswordHash(pin, pin_hash)
	fmt.Println("Match:   ", match)
}
