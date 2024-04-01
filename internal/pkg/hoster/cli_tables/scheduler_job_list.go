// Copyright 2024 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterTables

import (
	SchedulerClient "HosterCore/internal/app/scheduler/client"
	"fmt"
	"os"

	"github.com/aquasecurity/table"
)

func GenerateJobsTable(unix bool) error {
	jobs, err := SchedulerClient.GetJobList()
	if err != nil {
		return err
	}

	var t = table.New(os.Stdout)
	t.SetAlignment(
		table.AlignRight,  // ID number
		table.AlignLeft,   // Job ULID
		table.AlignCenter, // Job Status
		table.AlignCenter, // Job Type
		table.AlignCenter, // Resource Name
		table.AlignCenter, // Resource Type
		table.AlignCenter, // Replication:  snapshots transferred / total snapshots to transfer
		table.AlignLeft,   // Replication: bytes transferred / total amount of bytes
	)

	if unix {
		t.SetDividers(table.Dividers{
			ALL: " ",
			NES: " ",
			NSW: " ",
			NEW: " ",
			ESW: " ",
			NE:  " ",
			NW:  " ",
			SW:  " ",
			ES:  " ",
			EW:  " ",
			NS:  " ",
		})
		t.SetRowLines(false)
		t.SetBorderTop(false)
		t.SetBorderBottom(false)
	} else {
		t.SetHeaders("Scheduler Jobs")
		t.SetHeaderColSpans(0, 8)

		t.AddHeaders(
			"#",
			"Job ULID",
			"Job Status",
			"Job Type",
			"Resource Name",
			"Resource Type",
			"Repl: Snapshots",
			"Repl: Bytes",
		)

		t.SetLineStyle(table.StyleBrightCyan)
		t.SetDividers(table.UnicodeRoundedDividers)
		t.SetHeaderStyle(table.StyleBold)
	}

	for i, v := range jobs {
		jobStatus := ""
		if v.JobDone {
			jobStatus = "Done"
		} else if v.JobFailed {
			jobStatus = "Error"
		} else if v.JobInProgress {
			jobStatus = "In progress"
		} else {
			jobStatus = "Scheduled"
		}

		resName := ""
		if len(v.Snapshot.ResName) > 0 {
			resName = v.Snapshot.ResName
		}
		if len(v.Replication.ResName) > 0 {
			resName = v.Replication.ResName
		}

		if v.Replication.ProgressDoneSnaps == v.Replication.ProgressTotalSnaps {
			v.Replication.ProgressBytesDone = v.Replication.ProgressBytesTotal
		}

		t.AddRow(
			fmt.Sprintf("%d", i+1),
			v.JobId,
			jobStatus,
			v.JobType,
			resName,
			"res type will be here",
			fmt.Sprintf("%d/%d", v.Replication.ProgressDoneSnaps, v.Replication.ProgressTotalSnaps),
			fmt.Sprintf("%d/%d", v.Replication.ProgressBytesDone, v.Replication.ProgressBytesTotal),
		)
	}

	t.Render()
	return nil
}
