package HosterVmUtils

import (
	"HosterCore/internal/pkg/byteconversion"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type VmTemplate struct {
	Size       uint64 `json:"size"`       // Size in bytes
	SizeHuman  string `json:"size_human"` // Human-readable size (e.g. 5.0G)
	Name       string `json:"name"`       // Dataset name/path (e.g. tank/vm-encrypted/template-ubuntu2404)
	ShortName  string `json:"short_name"` // Short name, template- is trimmed (e.g. ubuntu2404)
	Mountpoint string `json:"mountpoint"` // Mountpoint (e.g. /tank/vm-encrypted/template-ubuntu2404)
}

func GetTemplates(ds string) (r []VmTemplate, e error) {
	// Clean up the dataset name
	ds = strings.TrimSpace(ds)
	ds = strings.TrimSuffix(ds, "/")
	ds = strings.TrimPrefix(ds, "/")

	// Example output:
	// [0 dataset]                                  [1 free]	    [2 used]	    [3 refer]	[4 mountpoint]
	// tank/vm-encrypted	                        143357046784	1774603857920	417792	    /tank/vm-encrypted
	// tank/vm-encrypted/elixirTry03	            10023571456	    1774603857920	7111569408	/tank/vm-encrypted/elixirTry03
	// tank/vm-encrypted/jail-template-14.2-RELEASE	605532160	    1774603857920	599846912	/tank/vm-encrypted/jail-template-14.2-RELEASE
	// tank/vm-encrypted/prometheus-hzima-0102	    1787916288	    1774603857920	1484595200	/tank/vm-encrypted/prometheus-hzima-0102
	// tank/vm-encrypted/template-debian12	        5373919232	    1774603857920	5373837312	/tank/vm-encrypted/template-debian12
	// tank/vm-encrypted/template-rockylinux8	    5373837312	    1774603857920	5373837312	/tank/vm-encrypted/template-rockylinux8
	// tank/vm-encrypted/template-rockylinux9	    5373919232	    1774603857920	5373837312	/tank/vm-encrypted/template-rockylinux9
	// tank/vm-encrypted/template-ubuntu2404	    5373837312	    1774603857920	5373837312	/tank/vm-encrypted/template-ubuntu2404
	out, err := exec.Command("zfs", "list", "-r", "-H", "-p", ds).CombinedOutput()
	if err != nil {
		e = fmt.Errorf("error getting VM template list: %s; %s", strings.TrimSpace(string(out)), err.Error())
		return
	}

	for _, v := range strings.Split(string(out), "\n") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		// Skip the dataset itself
		if v == ds {
			continue
		}
		// Split the line into fields
		parts := strings.Fields(v)
		if len(parts) < 5 {
			continue
		}

		templateName := parts[0]
		templateShortName := strings.TrimPrefix(templateName, ds+"/")

		if strings.Contains(templateShortName, "/") {
			// e = fmt.Errorf("template name contains a slash: %s", templateShortName)
			// return
			continue // Skip datasets that are not templates
		}

		if !strings.HasPrefix(templateShortName, "template-") {
			continue
		}
		templateShortName = strings.TrimPrefix(templateName, "template-")

		temp := VmTemplate{
			Name:       templateName,
			ShortName:  templateShortName,
			Mountpoint: parts[4],
		}
		temp.Size, _ = strconv.ParseUint(parts[1], 10, 64)
		temp.SizeHuman = byteconversion.BytesToHuman(temp.Size)

		r = append(r, temp)
	}

	return r, e
}
