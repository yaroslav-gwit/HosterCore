// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterHostUtils

import (
	"math/rand"
	"time"
)

// Generate a random password given the length and character types
func GenerateRandomPassword(length int, caps bool, nums bool) string {
	// Define the character set for the password
	charset := "abcdefghijklmnopqrstuvwxyz"
	capS := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numS := "0123456789"
	if caps {
		charset = charset + capS
	}
	if nums {
		charset = charset + numS
	}

	rand.Seed(time.Now().UnixNano())
	result := ""
	iter := 0
	for {
		pwByte := charset[rand.Intn(len(charset))]
		result = result + string(pwByte)
		iter = iter + 1
		if iter > length {
			break
		}
	}

	return result
}
