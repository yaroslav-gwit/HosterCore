// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package HosterHostUtils

import (
	FreeBSDKill "HosterCore/internal/pkg/freebsd/kill"
	FreeBSDPgrep "HosterCore/internal/pkg/freebsd/pgrep"
	"errors"
	"regexp"
)

func ReloadDns() error {
	svcInfo, err := dnsServiceInfo()
	if err != nil {
		reMatchExit1 := regexp.MustCompile(`exit status 1`)
		if reMatchExit1.MatchString(err.Error()) {
			return errors.New("DNS server is not running")
		} else {
			return err
		}
	}

	err = FreeBSDKill.KillProcess(FreeBSDKill.KillSignalHUP, svcInfo.Pid)
	if err != nil {
		return err
	}

	// emojlog.PrintLogMessage("DNS server config has been reloaded", emojlog.Changed)
	return nil
}

type DnsServiceInfo struct {
	Pid     int
	Running bool
}

func dnsServiceInfo() (r DnsServiceInfo, e error) {
	pids, err := FreeBSDPgrep.Pgrep("dns_server")
	if err != nil {
		reMatchExit1 := regexp.MustCompile(`exit status 1`)
		if reMatchExit1.MatchString(err.Error()) {
			e = errors.New("DNS server is not running")
		} else {
			e = err
		}
		return
	}

	if len(pids) < 1 {
		e = errors.New("DNS server is not running")
		return
	}

	reMatch := regexp.MustCompile(`.*dns_server$`)
	reMatchOld := regexp.MustCompile(`.*dns_server\s+&`)
	reMatchSkipLogProcess := regexp.MustCompile(`.*tail*`)

	for _, v := range pids {
		if reMatchSkipLogProcess.MatchString(v.ProcessCmd) {
			continue
		}

		if reMatch.MatchString(v.ProcessCmd) || reMatchOld.MatchString(v.ProcessCmd) {
			r.Pid = v.ProcessId
			r.Running = true
			return
		}
	}

	return
}
