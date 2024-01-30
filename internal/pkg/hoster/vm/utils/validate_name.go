// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterVmUtils

import (
	"errors"
	"strconv"
)

// Runs some VM or Jail name validations, before they can be deployed or cloned.
//
// Returns an error, if one of the checks fails.
func ValidateResName(resourceName string) error {
	minLength := 5
	maxLength := 25
	cantStartWith := "1234567890-_"
	validChars := "qwertyuiopasdfghjklzxcvbnm-QWERTYUIOPASDFGHJKLZXCVBNM_1234567890"

	// Check if vmName uses valid characters
	for _, v := range resourceName {
		valid := false
		for _, vv := range validChars {
			if v == vv {
				valid = true
				break
			}
		}
		if !valid {
			return errors.New("name cannot contain '" + string(v) + "' character")
		}
	}
	// EOF Check if vmName uses valid characters

	// Check if vmName starts with a valid character
	for i, v := range resourceName {
		if i > 1 {
			break
		}
		for _, vv := range cantStartWith {
			if v == vv {
				return errors.New("name cannot start with a number, an underscore or a hyphen")
			}
		}
	}
	// EOF Check if vmName starts with a valid character

	// Check the name length
	if len(resourceName) < minLength {
		return errors.New("name cannot contain less than " + strconv.Itoa(minLength) + " characters")
	} else if len(resourceName) > maxLength {
		return errors.New("name cannot contain more than " + strconv.Itoa(maxLength) + " characters")
	}
	// EOF Check vmName length

	return nil
}
