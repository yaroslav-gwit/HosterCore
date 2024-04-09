// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterCliJson

import (
	SchedulerClient "HosterCore/internal/app/scheduler/client"
	"encoding/json"
	"fmt"
)

func GenerateSchedulerJobInfo(jobID string, pretty bool) error {
	resp, err := SchedulerClient.GetJobInfo(jobID)
	if err != nil {
		return err
	}

	var out []byte
	if pretty {
		out, err = json.MarshalIndent(resp, "", "   ")
		if err != nil {
			return err
		}
	} else {
		out, err = json.Marshal(resp)
		if err != nil {
			return err
		}
	}

	fmt.Println(string(out))
	return nil
}

func GenerateSchedulerJson(pretty bool) error {
	resp, err := SchedulerClient.GetJobList()
	if err != nil {
		return err
	}

	var out []byte
	if pretty {
		out, err = json.MarshalIndent(resp, "", "   ")
		if err != nil {
			return err
		}
	} else {
		out, err = json.Marshal(resp)
		if err != nil {
			return err
		}
	}

	fmt.Println(string(out))
	return nil
}
