package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/red-hat-storage/odf-operator/services/devicefinder/diskutils"
	"github.com/red-hat-storage/odf-operator/services/devicefinder/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	udevEventPeriod = 5 * time.Second
	probeInterval   = 5 * time.Minute
)

// readFileFunc is the function used to read a file. It is a package-level
// variable so that tests can substitute a fake implementation without touching
// the real filesystem.
var readFileFunc = os.ReadFile

// getDASDUID reads the DASD UID from /sys/block/<name>/device/uid and returns
// a normalised identifier suitable for use as a Kubernetes object name.
//
// Two UID formats are produced by the IBM Z firmware:
//   - Normal   (4 tokens): IBM.75000000092461.e900.10
//   - Extended (5 tokens): IBM.750000000FMZ21.db80.34.0000000000021f7400000000000000
//
// For extended UIDs the trailing token (the 32-hex-character LUN extension) is
// removed so that both forms collapse to the same 4-token base.
// UID is converted to lowercase and dots are replaced with dashes
// to comply with Kubernetes DNS-label naming rules.
//
// An error is returned when the sysfs file cannot be read (e.g. on non-IBM-Z
// systems or when the device does not exist).
func getDASDUID(name string) (string, error) {
	path := fmt.Sprintf("/sys/block/%s/device/uid", name)
	data, err := readFileFunc(path)
	if err != nil {
		return "", fmt.Errorf("failed to read DASD UID from %s: %w", path, err)
	}
	uid := strings.TrimSpace(strings.ToLower(string(data)))
	tokens := strings.Split(uid, ".")
	// Extended UIDs have an extra trailing token; strip it so both UID types
	// share the same 4-token base representation.
	if len(tokens) >= 5 {
		tokens = tokens[:len(tokens)-1]
	}
	return strings.Join(tokens, "-"), nil
}

var supportedDeviceTypes = sets.NewString("mpath", "disk")

// DeviceDiscovery instance
type DeviceDiscovery struct {
	kubeClient client.Client
	disks      []types.DiscoveredDevice
}

// NewDeviceDiscovery returns a new DeviceDiscovery instance
func NewDeviceDiscovery() (*DeviceDiscovery, error) {
	// Create Kubernetes client
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add corev1 to scheme: %w", err)
	}

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	return &DeviceDiscovery{kubeClient: k8sClient}, nil
}

// Start the device discovery process
func (discovery *DeviceDiscovery) Start() error {
	klog.Info("starting device discovery")

	err := discovery.discoverDevices()
	if err != nil {
		return fmt.Errorf("failed to discover devices: %w", err)
	}

	// Watch udev events for continuous discovery of devices
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM)

	udevEvents := make(chan string)
	go udevBlockMonitor(udevEvents, udevEventPeriod)
	for {
		select {
		case <-sigc:
			klog.Info("shutdown signal received, exiting...")
			return nil
		case <-time.After(probeInterval):
			if err := discovery.discoverDevices(); err != nil {
				klog.Errorf("failed to discover devices during probe interval. %v", err)
			}
		case _, ok := <-udevEvents:
			if ok {
				klog.Info("trigger probe from udev event")
				if err := discovery.discoverDevices(); err != nil {
					klog.Errorf("failed to discover devices triggered from udev event. %v", err)
				}
			} else {
				klog.Warningf("disabling udev monitoring")
				udevEvents = nil
			}
		}
	}
}

// discoverDevices identifies the list of usable disks on the current node
func (discovery *DeviceDiscovery) discoverDevices() error {
	// List all the valid block devices on the node
	validDevices, err := getValidBlockDevices()
	if err != nil {
		message := "failed to discover devices"
		klog.Errorf("%s. Error: %+v", message, err)
		return fmt.Errorf("%s: %w", message, err)
	}

	klog.Infof("valid block devices: %+v", validDevices)

	discoveredDisks := getDiscoverdDevices(validDevices)
	klog.Infof("discovered devices: %+v", discoveredDisks)

	// Update discovered devices in the ConfigMap
	if !reflect.DeepEqual(discovery.disks, discoveredDisks) {
		klog.Info("device list updated. Updating ConfigMap...")
		discovery.disks = discoveredDisks
		err = discovery.updateConfigMap()
		if err != nil {
			message := "failed to update ConfigMap"
			klog.Errorf("%s. Error: %+v", message, err)
			return fmt.Errorf("%s: %w", message, err)
		}
		klog.Info("successfully updated discovered device details in the ConfigMap")
	}

	return nil
}

// getValidBlockDevices fetches and unmarshalls all the block devices sutitable for discovery
func getValidBlockDevices() ([]diskutils.BlockDevice, error) {
	lDevices := diskutils.BlockDeviceList{}
	blockDevices, err := diskutils.GetBlockDevices()
	if err != nil {
		return lDevices.BlockDevices, err
	}
	err = json.Unmarshal(blockDevices, &lDevices)
	if err != nil {
		return lDevices.BlockDevices, err
	}

	return lDevices.BlockDevices, nil
}

// getDiscoverdDevices creates types.DiscoveredDevice from diskutil.BlockDevices
func getDiscoverdDevices(blockDevices []diskutils.BlockDevice) []types.DiscoveredDevice {
	discoveredDevices := make([]types.DiscoveredDevice, 0)
	for idx := range blockDevices {
		if ignoreDevices(&blockDevices[idx]) {
			continue
		}

		var deviceID string
		if blockDevices[idx].IsDASD() {
			var err error
			deviceID, err = getDASDUID(blockDevices[idx].Name)
			if err != nil {
				klog.Warningf(
					"failed to get DASD UID for device %q. Error %v",
					blockDevices[idx].Name,
					err,
				)
				continue
			}
		} else {
			var err error
			deviceID, err = blockDevices[idx].GetPathByID()
			if err != nil {
				klog.Warningf(
					"failed to get persistent ID for the device %q. Error %v",
					blockDevices[idx].Name,
					err,
				)
			}
		}

		path, err := blockDevices[idx].GetDevPath()
		if err != nil {
			klog.Warningf(
				"failed to parse path for the device %q. Error %v",
				blockDevices[idx].KName,
				err,
			)
		}
		discoveredDevice := types.DiscoveredDevice{
			Path:     path,
			Model:    blockDevices[idx].Model,
			Vendor:   blockDevices[idx].Vendor,
			Type:     parseDeviceType(blockDevices[idx].Type),
			DeviceID: deviceID,
			Size:     blockDevices[idx].Size,
			WWN:      blockDevices[idx].WWN,
		}
		discoveredDevices = append(discoveredDevices, discoveredDevice)
	}
	return uniqueDevices(discoveredDevices)
}

// uniqueDevices removes duplicate devices from the list.
// For devices with a WWN the WWN is used as the deduplication key.
// DASD devices have no WWN; they are deduplicated by DeviceID (the DASD UID)
// when available, and fall back to Path when the UID could not be read.
func uniqueDevices(sample []types.DiscoveredDevice) []types.DiscoveredDevice {
	var unique []types.DiscoveredDevice
	type key struct{ value string }
	m := make(map[key]int)
	for _, v := range sample {
		// Prefer WWN, then DeviceID (populated for DASD devices via the UID
		// sysfs file), then fall back to Path.
		deduplicationKey := v.WWN
		if deduplicationKey == "" {
			deduplicationKey = v.DeviceID
		}
		if deduplicationKey == "" {
			deduplicationKey = v.Path
		}
		k := key{deduplicationKey}
		if i, ok := m[k]; ok {
			unique[i] = v
		} else {
			m[k] = len(unique)
			unique = append(unique, v)
		}
	}
	return unique
}

// ignoreDevices checks if a device should be ignored during discovery
func ignoreDevices(dev *diskutils.BlockDevice) bool {
	if dev.ReadOnly {
		klog.Infof("ignoring read only device %q", dev.Name)
		return true
	}

	if dev.State == diskutils.StateSuspended {
		klog.Infof("ignoring device %q with invalid state %q", dev.Name, dev.State)
		return true
	}

	if !supportedDeviceTypes.Has(dev.Type) {
		klog.Infof("ignoring device %q with unsupported type %q", dev.Name, dev.Type)
		return true
	}

	if dev.Removable {
		klog.Infof("ignoring device %s with removable capability", dev.Name)
		return true
	}

	if dev.BiosPartition() {
		klog.Infof("ignoring device %q with partition with bios/boot label", dev.Name)
		return true
	}

	if dev.Mountpoint != "" {
		klog.Infof("ignoring device %q with at least one mountpoint", dev.Mountpoint)
		return true
	}

	if dev.Size == 0 {
		klog.Infof("ignoring device %q with 0 size", dev.Name)
		return true
	}

	// DASD devices on IBM Z never carry a WWN; accept them based on their
	// device type rather than requiring a WWN.
	if dev.WWN == "" && !dev.IsDASD() {
		klog.Infof("ignoring device %q without WWN", dev.Name)
		return true
	}

	if dev.FSType != "" && dev.FSType != "mpath_member" {
		klog.Infof("ignoring device %q with FS", dev.FSType)
		return true
	}
	// Ignore children which has partition/fs on them
	if dev.Children != nil {
		for idx := range dev.Children {
			if dev.Children[idx].Type == "part" {
				klog.Infof("ignoring device %q with partitions", dev.Children[idx].Name)
				return true
			}
			if dev.Children[idx].FSType != "" {
				klog.Infof(
					"ignoring device %q with filesystem %s",
					dev.Children[idx].Name,
					dev.Children[idx].FSType,
				)
				return true
			}
			if dev.Children[idx].Children != nil {
				for idx2 := range dev.Children[idx].Children {
					return ignoreDevices(&dev.Children[idx].Children[idx2])
				}
			}
		}
	}

	return false
}

func parseDeviceType(deviceType string) types.DiscoveredDeviceType {
	switch deviceType {
	case "disk":
		return types.DiskType
	case "part":
		return types.PartType
	case "lvm":
		return types.LVMType
	case "mpath":
		return types.MultiPathType
	default:
		return ""
	}
}
