package diskutils

import (
	"fmt"
	"os/exec"
	"strings"

	"k8s.io/klog/v2"
)

var (
	ExecCommand CommandExecutor
)

func init() {
	ExecCommand = CmdExec{}
}

const (
	// StateSuspended is a possible value of BlockDevice.State
	StateSuspended = "suspended"

	// dasdSubsystem is the Linux subsystem name reported by lsblk for DASD devices on IBM Z.
	dasdSubsystem = "dasd"
)

type CommandExecutor interface {
	Execute(name string, args ...string) Command
}

type CmdExec struct {
}

func (c CmdExec) Execute(name string, args ...string) Command {
	//nolint:gosec
	// G204: false positive - Execute is only called with hardcoded
	// "lsblk" command, name is never derived from user input
	return exec.Command(name, args...)
}

type Command interface {
	CombinedOutput() ([]byte, error)
}

// BlockDevice is the a block device as output by lsblk.
type BlockDeviceList struct {
	BlockDevices []BlockDevice `json:"blockdevices"`
}

type BlockDevice struct {
	Name       string        `json:"name"`
	Type       string        `json:"type"`
	Size       int64         `json:"size"`
	PathByID   string        `json:"id-link,omitempty"`
	Model      string        `json:"model,omitempty"`
	Vendor     string        `json:"vendor,omitempty"`
	ReadOnly   bool          `json:"RO,omitempty"`
	Removable  bool          `json:"RM,omitempty"`
	State      string        `json:"state,omitempty"`
	KName      string        `json:"kname"`
	FSType     string        `json:"fstype,omitempty"`
	PartLabel  string        `json:"partlabel,omitempty"`
	Path       string        `json:"path,omitempty"`
	WWN        string        `json:"wwn,omitempty"`
	Subsystems string        `json:"subsystems,omitempty"`
	Children   []BlockDevice `json:"children,omitempty"`
	Mountpoint string        `json:"mountpoint,omitempty"`
}

// IsDASD returns true when the block device is an IBM Z DASD (Direct
// Access Storage Device).  DASD disks are identified by their name prefix
// ("dasd") or, more robustly, by the "dasd" token present in the
// colon-separated subsystems field reported by lsblk (e.g.
// "block:dasd:ccw").  They never carry a WWN, so the standard WWN-based
// acceptance path must be bypassed for them.
func (b *BlockDevice) IsDASD() bool {
	if strings.HasPrefix(b.Name, "dasd") {
		return true
	}
	for _, sub := range strings.Split(b.Subsystems, ":") {
		if sub == dasdSubsystem {
			return true
		}
	}
	return false
}

func (b *BlockDevice) BiosPartition() bool {
	if b.Children != nil {
		for idx := range b.Children {
			if strings.Contains(
				strings.ToLower(b.Children[idx].PartLabel),
				strings.ToLower("bios")) ||
				strings.Contains(
					strings.ToLower(b.Children[idx].PartLabel),
					strings.ToLower("boot")) {
				return true
			}
			continue
		}
	}
	return strings.Contains(strings.ToLower(b.PartLabel), strings.ToLower("bios")) ||
		strings.Contains(strings.ToLower(b.PartLabel), strings.ToLower("boot"))
}

// GetDevPath for block device (/dev/sdx)
func (b *BlockDevice) GetDevPath() (path string, err error) {
	if b.FSType == "mpath_member" {
		// correct mpaths always have a single children
		if len(b.Children) != 0 {
			return fmt.Sprintf("/dev/%s", b.Children[0].KName), nil
		}
		return "", fmt.Errorf("no multipath members found %s", b.KName)
	}
	return b.Path, nil
}

// GetPathByID check on BlockDevice
func (b *BlockDevice) GetPathByID() (string, error) {
	if b.FSType == "mpath_member" {
		// correct mpaths always have a single children
		if len(b.Children) != 0 {
			return fmt.Sprintf("/dev/disk/by-id/%s", b.Children[0].PathByID), nil
		}
		return "", fmt.Errorf("no multipath members found %s", b.KName)
	}

	if b.PathByID != "" {
		return fmt.Sprintf("/dev/disk/by-id/%s", b.PathByID), nil
	}
	return "", fmt.Errorf("disk has no persistent ID")
}

// GetBlockDevices using the lsblk command
func GetBlockDevices() ([]byte, error) {
	args := []string{"--json", "-O", "-b"}
	cmd := ExecCommand.Execute("lsblk", args...)
	klog.Infof("Executing command: %#v", cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return []byte{}, fmt.Errorf("failed to run command: %s", err)
	}
	return output, err
}
