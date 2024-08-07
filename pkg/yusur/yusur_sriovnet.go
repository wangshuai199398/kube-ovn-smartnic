package yusur

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	NetSysDir = "/sys/class/net"
	PciSysDir = "/sys/bus/pci/devices"

	netdevPhysSwitchID = "phys_switch_id"
	netdevPhysPortName = "phys_port_name"

	HwAddr        = "hw"
	YusurSmartNic = "smart-nic"
	PlatName      = "plat_name"
)

var virtFnRe = regexp.MustCompile(`virtfn(\d+)`)

// IsYusurSmartNic check Yusur Nic is not smart nic
func IsYusurSmartNic(pciAddress string) bool {
	platFile := filepath.Join(PciSysDir, pciAddress, HwAddr, PlatName)

	platName, err := os.ReadFile(platFile)
	if err != nil {
		return false
	}

	yusurSmartNic := strings.TrimSpace(string(platName))
	if strings.HasSuffix(yusurSmartNic, YusurSmartNic) {
		return true
	}

	return false
}

// GetPfPciFromVfPci retrieves the parent PF PCI address of the provided VF PCI address in D:B:D.f format
func GetPfPciFromVfPci(vfPciAddress string) (string, error) {
	pfPath := filepath.Join(PciSysDir, vfPciAddress, "physfn")
	pciDevDir, err := os.Readlink(pfPath)
	if err != nil {
		return "", fmt.Errorf("failed to read physfn link, provided address may not be a VF. %v", err)
	}

	pf := path.Base(pciDevDir)
	if pf == "" {
		return pf, fmt.Errorf("could not find PF PCI Address")
	}
	return pf, err
}

// GetPfIndexByPciAddress gets a VF PCI address and
// returns the correlate PF index.
func GetYsk2PfIndexByPciAddress(pfPci string) (int, error) {
	pfIndex, err := strconv.Atoi(string((pfPci[len(pfPci)-1])))
	if err != nil {
		return -1, fmt.Errorf("failed to get pfPci of device %s %w", pfPci, err)
	}

	return pfIndex, nil
}

// GetYsk2VfIndexByPciAddress gets a VF PCI address and
// returns the correlate VF index.
func GetYsk2VfIndexByPciAddress(vfPciAddress string) (int, error) {
	vfPath := filepath.Join(PciSysDir, vfPciAddress, "physfn", "virtfn*")
	matches, err := filepath.Glob(vfPath)
	if err != nil {
		return -1, err
	}
	for _, match := range matches {
		tmp, err := os.Readlink(match)
		if err != nil {
			continue
		}
		if strings.Contains(tmp, vfPciAddress) {
			result := virtFnRe.FindStringSubmatch(match)
			vfIndex, err := strconv.Atoi(result[1])
			if err != nil {
				continue
			}
			return vfIndex, nil
		}
	}
	return -1, fmt.Errorf("vf index for %s not found", vfPciAddress)
}

func GetYsk2VfRepresentor(pfIndex int, vfIndex int) string {
	vfr := fmt.Sprintf("pf%dvf%drep", pfIndex, vfIndex)
	return vfr
}
