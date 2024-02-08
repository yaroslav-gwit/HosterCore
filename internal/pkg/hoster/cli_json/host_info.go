// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterCliJson

import (
	HosterHostUtils "HosterCore/internal/pkg/hoster/host/utils"
	"encoding/json"
	"fmt"
)

func GenerateHostInfoJson(pretty bool) error {
	info, err := HosterHostUtils.GetHostInfo()
	if err != nil {
		return nil
	}

	if pretty {
		j, err := json.MarshalIndent(info, "", JSON_INDENT)
		if err != nil {
			return err
		}
		fmt.Println(string(j))
	} else {
		j, err := json.Marshal(info)
		if err != nil {
			return err
		}
		fmt.Println(string(j))
	}

	return nil
}
