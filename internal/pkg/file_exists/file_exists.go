package FileExists

import (
	"errors"
	"os"
)

func CheckUsingOsStat(filePath string) bool {
	_, error := os.Stat(filePath)
	//return !os.IsNotExist(err)
	return !errors.Is(error, os.ErrNotExist)
}
