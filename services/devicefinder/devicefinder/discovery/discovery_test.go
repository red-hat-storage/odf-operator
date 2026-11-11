//nolint:lll,dupl
package discovery

import (
	"encoding/json"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/red-hat-storage/odf-operator/services/devicefinder/diskutils"
	"github.com/red-hat-storage/odf-operator/services/devicefinder/types"
)

var _ = Describe("Device Discovery", func() {
	var (
		deviceList7Disk              diskutils.BlockDeviceList
		deviceList0Disk              diskutils.BlockDeviceList
		deviceList2Disk2MultiPath    diskutils.BlockDeviceList
		deviceList0Disk0MultiPath    diskutils.BlockDeviceList
		deviceListSanDisk            diskutils.BlockDeviceList
		deviceList0Disk0MultiPath0DM diskutils.BlockDeviceList
		deviceListDasdDisk           diskutils.BlockDeviceList
	)
	Context("When scanning for disks", func() {

		BeforeEach(func() {
			By("before tests")
			LsblkOut2Disk2MultiPath, err := os.ReadFile(
				"../../test/data/mpath-4-available-disk.json",
			)
			Expect(err).To(Not(HaveOccurred()))

			LsblkOut0Disk0MultiPath, err := os.ReadFile(
				"../../test/data/mpath-0-available-disk.json",
			)
			Expect(err).To(Not(HaveOccurred()))

			LsblkOut7Disk, err := os.ReadFile("../../test/data/7-available-disk.json")
			Expect(err).To(Not(HaveOccurred()))

			LsblkOut0Disk, err := os.ReadFile("../../test/data/0-available-disk.json")
			Expect(err).To(Not(HaveOccurred()))

			LsblkOutSanDisk, err := os.ReadFile("../../test/data/e1-fast.json")
			Expect(err).To(Not(HaveOccurred()))

			LsblkOutDasdDisk, err := os.ReadFile("../../test/data/dasd-available-disk.json")
			Expect(err).To(Not(HaveOccurred()))

			err = json.Unmarshal(LsblkOut0Disk, &deviceList0Disk)
			Expect(err).To(Not(HaveOccurred()))

			err = json.Unmarshal(LsblkOut7Disk, &deviceList7Disk)
			Expect(err).To(Not(HaveOccurred()))

			err = json.Unmarshal(LsblkOut2Disk2MultiPath, &deviceList2Disk2MultiPath)
			Expect(err).To(Not(HaveOccurred()))

			err = json.Unmarshal(LsblkOut0Disk0MultiPath, &deviceList0Disk0MultiPath)
			Expect(err).To(Not(HaveOccurred()))

			err = json.Unmarshal(LsblkOutSanDisk, &deviceListSanDisk)
			Expect(err).To(Not(HaveOccurred()))

			err = json.Unmarshal(LsblkOutDasdDisk, &deviceListDasdDisk)
			Expect(err).To(Not(HaveOccurred()))
		})

		AfterEach(func() {
			// // TODO(user): Cleanup logic after each test
		})

		It(
			"should have the correct number of discovered disks with multipath (input data 1)",
			func() {
				discoveredDisks := getDiscoverdDevices(deviceList2Disk2MultiPath.BlockDevices)
				Expect(discoveredDisks).To(HaveLen(4))
			},
		)
		It("should have the correct disks with multipath (input data 1)", func() {
			discoveredDisks := getDiscoverdDevices(deviceList2Disk2MultiPath.BlockDevices)

			Expect(discoveredDisks).To(ContainElement(
				types.DiscoveredDevice{
					DeviceID: "/dev/disk/by-id/dm-name-mpathb",
					Path:     "/dev/dm-1",
					Model:    "iscsi_disk1",
					Type:     "disk",
					Vendor:   "LIO-ORG ",
					Size:     75161927680,
					WWN:      "0x6001405c595842b2d484d0bb11e42179",
				},
			))
			Expect(discoveredDisks).To(ContainElement(
				types.DiscoveredDevice{
					DeviceID: "/dev/disk/by-id/dm-name-mpatha",
					Path:     "/dev/dm-0",
					Model:    "iscsi_disk2",
					Type:     "disk",
					Vendor:   "LIO-ORG ",
					Size:     85899345920,
					WWN:      "0x60014056ade16393c8f412da451430e4",
				},
			))
			Expect(discoveredDisks).To(ContainElement(
				types.DiscoveredDevice{
					DeviceID: "/dev/disk/by-id/scsi-35000c50015ff75aa",
					Path:     "/dev/sdb",
					Model:    "QEMU HARDDISK",
					Type:     "disk",
					Vendor:   "QEMU    ",
					Size:     10737418240,
					WWN:      "0x5000c50015ff75aa",
				},
			))
			Expect(discoveredDisks).To(ContainElement(
				types.DiscoveredDevice{
					DeviceID: "/dev/disk/by-id/scsi-35000c50015ea75bb",
					Path:     "/dev/sde",
					Model:    "QEMU HARDDISK",
					Type:     "disk",
					Vendor:   "QEMU    ",
					Size:     53687091200,
					WWN:      "0x5000c50015ea75bb",
				},
			))

		})

		It(
			"should have the correct number of discovered disks with multipath (input data 2)",
			func() {
				discoveredDisks := getDiscoverdDevices(deviceList0Disk0MultiPath.BlockDevices)
				Expect(discoveredDisks).To(BeEmpty())
			},
		)

		It(
			"should have the correct number of discovered disks without multipath (input data 3)",
			func() {
				discoveredDisks := getDiscoverdDevices(deviceList7Disk.BlockDevices)
				Expect(discoveredDisks).To(HaveLen(2))
			},
		)

		It("should have the correct disks without multipath (input data 3)", func() {
			discoveredDisks := getDiscoverdDevices(deviceList7Disk.BlockDevices)
			Expect(discoveredDisks).To(ContainElement(
				types.DiscoveredDevice{
					DeviceID: "",
					Path:     "/dev/sdk",
					Model:    "LUN C-Mode",
					Type:     "disk",
					Vendor:   "NETAPP  ",
					Size:     4294967296,
					WWN:      "0x600a098038304437415d4b6a5968624d",
				},
			))
			Expect(discoveredDisks).To(ContainElement(
				types.DiscoveredDevice{
					DeviceID: "",
					Path:     "/dev/sdj",
					Model:    "LUN C-Mode",
					Type:     "disk",
					Vendor:   "NETAPP  ",
					Size:     1288490188800,
					WWN:      "0x600a098038304437415d4b6a5968624f",
				},
			))
		})

		It("should have the correct number of discovered disks (input data 4)", func() {
			discoveredDisks := getDiscoverdDevices(deviceList0Disk.BlockDevices)
			Expect(discoveredDisks).To(BeEmpty())
		})

		It("should have the correct number of discovered disks (san disk env)", func() {
			discoveredDisks := getDiscoverdDevices(deviceListSanDisk.BlockDevices)
			Expect(discoveredDisks).To(BeEmpty())
		})

		It("should have the correct number of discovered disks (qe env data)", func() {
			LsblkOut0Disk0MultiPath0DM, err := os.ReadFile("../../test/data/mpath-0-available-disk-0dm-s.json")
			Expect(err).To(Not(HaveOccurred()))

			err = json.Unmarshal(LsblkOut0Disk0MultiPath0DM, &deviceList0Disk0MultiPath0DM)
			Expect(err).To(Not(HaveOccurred()))

			discoveredDisks := getDiscoverdDevices(deviceList0Disk0MultiPath0DM.BlockDevices)
			Expect(discoveredDisks).To(BeEmpty())

		})

		// IBM Z DASD device tests
		Context("DASD devices on IBM Z (s390x)", func() {
			// Fake UIDs returned by the injected readFileFunc for each DASD device.
			// dasda uses a normal UID (4 tokens), dasdb uses an extended UID (5 tokens),
			// and dasdc uses another normal UID.
			const (
				// Normal UID — 4 dot-separated tokens.
				dasdaRawUID = "IBM.75000000092461.e900.10"
				// Extended UID — 5 dot-separated tokens; the trailing hex token is stripped.
				dasdbRawUID = "IBM.750000000FMZ21.db80.34.0000000000021f7400000000000000"
				dasdcRawUID = "IBM.75000000092462.e900.11"

				// Expected DeviceID values after normalisation (dots → dashes,
				// extended trailing token removed).
				dasdaUID = "ibm-75000000092461-e900-10"
				dasdbUID = "ibm-750000000fmz21-db80-34"
				dasdcUID = "ibm-75000000092462-e900-11"
			)

			BeforeEach(func() {
				// Inject a fake readFileFunc so that getDASDUID does not touch
				// the real /sys/block filesystem during tests.
				readFileFunc = func(name string) ([]byte, error) {
					uids := map[string]string{
						"/sys/block/dasda/device/uid": dasdaRawUID,
						"/sys/block/dasdb/device/uid": dasdbRawUID,
						"/sys/block/dasdc/device/uid": dasdcRawUID,
					}
					if uid, ok := uids[name]; ok {
						return []byte(uid + "\n"), nil
					}
					return nil, os.ErrNotExist
				}
			})

			AfterEach(func() {
				readFileFunc = os.ReadFile
			})

			It("should discover exactly 3 available DASD disks and skip sda (boot) and dasdx (has FS)", func() {
				discoveredDisks := getDiscoverdDevices(deviceListDasdDisk.BlockDevices)
				Expect(discoveredDisks).To(HaveLen(3))
			})

			It("should discover dasda (normal UID) with correct properties", func() {
				discoveredDisks := getDiscoverdDevices(deviceListDasdDisk.BlockDevices)
				Expect(discoveredDisks).To(ContainElement(
					types.DiscoveredDevice{
						DeviceID: dasdaUID,
						Path:     "/dev/dasda",
						Model:    "",
						Type:     types.DiskType,
						Vendor:   "IBM     ",
						Size:     102574080000,
						WWN:      "",
					},
				))
			})

			It("should discover dasdb (extended UID) with trailing token stripped", func() {
				discoveredDisks := getDiscoverdDevices(deviceListDasdDisk.BlockDevices)
				Expect(discoveredDisks).To(ContainElement(
					types.DiscoveredDevice{
						DeviceID: dasdbUID,
						Path:     "/dev/dasdb",
						Model:    "",
						Type:     types.DiskType,
						Vendor:   "IBM     ",
						Size:     102574080000,
						WWN:      "",
					},
				))
			})

			It("should discover dasdc (normal UID) with correct properties", func() {
				discoveredDisks := getDiscoverdDevices(deviceListDasdDisk.BlockDevices)
				Expect(discoveredDisks).To(ContainElement(
					types.DiscoveredDevice{
						DeviceID: dasdcUID,
						Path:     "/dev/dasdc",
						Model:    "",
						Type:     types.DiskType,
						Vendor:   "IBM     ",
						Size:     102574080000,
						WWN:      "",
					},
				))
			})

			It("should not discover sda (SCSI boot disk with boot partition label)", func() {
				discoveredDisks := getDiscoverdDevices(deviceListDasdDisk.BlockDevices)
				for _, d := range discoveredDisks {
					Expect(d.Path).NotTo(Equal("/dev/sda"))
				}
			})

			It("should not discover dasdx (DASD disk with existing filesystem)", func() {
				discoveredDisks := getDiscoverdDevices(deviceListDasdDisk.BlockDevices)
				for _, d := range discoveredDisks {
					Expect(d.Path).NotTo(Equal("/dev/dasdx"))
				}
			})
		})
	})
})
