// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterCliJson

import (
	HosterJailUtils "HosterCore/internal/pkg/hoster/jail/utils"
	"encoding/json"
	"fmt"
)

func GenerateJailsJson(unixStyleTable bool) error {
	list, err := HosterJailUtils.ListAllExtendedTable()
	if err != nil {
		return err
	}

	j, err := json.MarshalIndent(list, "", JSON_INDENT)
	if err != nil {
		return err
	}
	fmt.Println(string(j))

	return nil
}
