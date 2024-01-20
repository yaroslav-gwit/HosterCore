package HosterVmUtils

import (
	"errors"
	"strconv"
)

func ValidateResName(resourceName string) error {
	vmNameMinLength := 5
	vmNameMaxLength := 22
	vmNameCantStartWith := "1234567890-_"
	vmNameValidChars := "qwertyuiopasdfghjklzxcvbnm-QWERTYUIOPASDFGHJKLZXCVBNM_1234567890"

	// Check if vmName uses valid characters
	for _, v := range resourceName {
		valid := false
		for _, vv := range vmNameValidChars {
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
		for _, vv := range vmNameCantStartWith {
			if v == vv {
				return errors.New("name cannot start with a number, an underscore or a hyphen")
			}
		}
	}
	// EOF Check if vmName starts with a valid character

	// Check vmName length
	if len(resourceName) < vmNameMinLength {
		return errors.New("name cannot contain less than " + strconv.Itoa(vmNameMinLength) + " characters")
	} else if len(resourceName) > vmNameMaxLength {
		return errors.New("name cannot contain more than " + strconv.Itoa(vmNameMaxLength) + " characters")
	}
	// EOF Check vmName length

	return nil
}
