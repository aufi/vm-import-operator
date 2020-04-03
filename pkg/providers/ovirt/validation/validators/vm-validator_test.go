package validators_test

import (
	"github.com/kubevirt/vm-import-operator/pkg/providers/ovirt/validation/validators"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	ovirtsdk "github.com/ovirt/go-ovirt"
)

var _ = Describe("Validating VM", func() {
	It("should accept vm ", func() {
		var vm = newVM()

		failures := validators.ValidateVM(vm)

		Expect(failures).To(BeEmpty())
	})
	It("should reject VM with no status ", func() {
		var vm = newVMWithStatusControl(false)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMStatusID))
	})
	table.DescribeTable("should accept VM with legal status", func(status ovirtsdk.VmStatus) {
		vm := newVM()
		vm.SetStatus(status)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(BeEmpty())
	},
		table.Entry(string(ovirtsdk.VMSTATUS_DOWN), ovirtsdk.VMSTATUS_DOWN),
		table.Entry(string(ovirtsdk.VMSTATUS_UP), ovirtsdk.VMSTATUS_UP),
	)
	table.DescribeTable("should flag VM with illegal status", func(status ovirtsdk.VmStatus) {
		vm := newVM()
		vm.SetStatus(status)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMStatusID))
	},
		table.Entry(string(ovirtsdk.VMSTATUS_IMAGE_LOCKED), ovirtsdk.VMSTATUS_IMAGE_LOCKED),
		table.Entry(string(ovirtsdk.VMSTATUS_MIGRATING), ovirtsdk.VMSTATUS_MIGRATING),
		table.Entry(string(ovirtsdk.VMSTATUS_NOT_RESPONDING), ovirtsdk.VMSTATUS_NOT_RESPONDING),
		table.Entry(string(ovirtsdk.VMSTATUS_PAUSED), ovirtsdk.VMSTATUS_PAUSED),
		table.Entry(string(ovirtsdk.VMSTATUS_POWERING_DOWN), ovirtsdk.VMSTATUS_POWERING_DOWN),
		table.Entry(string(ovirtsdk.VMSTATUS_POWERING_UP), ovirtsdk.VMSTATUS_POWERING_UP),
		table.Entry(string(ovirtsdk.VMSTATUS_REBOOT_IN_PROGRESS), ovirtsdk.VMSTATUS_REBOOT_IN_PROGRESS),
		table.Entry(string(ovirtsdk.VMSTATUS_RESTORING_STATE), ovirtsdk.VMSTATUS_RESTORING_STATE),
		table.Entry(string(ovirtsdk.VMSTATUS_SAVING_STATE), ovirtsdk.VMSTATUS_SAVING_STATE),
		table.Entry(string(ovirtsdk.VMSTATUS_SUSPENDED), ovirtsdk.VMSTATUS_SUSPENDED),
		table.Entry(string(ovirtsdk.VMSTATUS_UNASSIGNED), ovirtsdk.VMSTATUS_UNASSIGNED),
		table.Entry(string(ovirtsdk.VMSTATUS_UNKNOWN), ovirtsdk.VMSTATUS_UNKNOWN),
		table.Entry(string(ovirtsdk.VMSTATUS_WAIT_FOR_LAUNCH), ovirtsdk.VMSTATUS_WAIT_FOR_LAUNCH),
	)
	It("should flag vm with boot menu enabled ", func() {
		var vm = newVM()
		bios := vm.MustBios()
		bootMenu := ovirtsdk.BootMenu{}
		bootMenu.SetEnabled(true)
		bios.SetBootMenu(&bootMenu)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMBiosBootMenuID))
	})
	It("should flag vm with no bios type ", func() {
		var vm = newVM()
		bios := ovirtsdk.Bios{}
		vm.SetBios(&bios)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMBiosTypeID))
	})
	It("should flag vm with q35_secure_boot bios ", func() {
		var vm = newVM()
		bios := vm.MustBios()
		bios.SetType("q35_secure_boot")

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMBiosTypeQ35SecureBootID))
	})
	It("should flag vm with s390x CPU ", func() {
		var vm = newVM()
		cpu := ovirtsdk.Cpu{}
		cpu.SetArchitecture("s390x")
		vm.SetCpu(&cpu)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMCpuArchitectureID))
	})
	table.DescribeTable("should flag CPU with illegal pinning for", func(pins []*ovirtsdk.VcpuPin) {
		vm := newVM()
		vm.MustCpu().MustCpuTune().MustVcpuPins().SetSlice(pins)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMCpuTuneID))
	},
		table.Entry("duplicate pins", []*ovirtsdk.VcpuPin{newCPUPin(0, "0"), newCPUPin(1, "0")}),
		table.Entry("cpu range", []*ovirtsdk.VcpuPin{newCPUPin(0, "0-1"), newCPUPin(1, "0-1")}),
		table.Entry("cpu set", []*ovirtsdk.VcpuPin{newCPUPin(0, "0,1"), newCPUPin(1, "0,1")}),
		table.Entry("cpu exclusion", []*ovirtsdk.VcpuPin{newCPUPin(0, "^1")}),
	)
	It("should flag vm with CPU shares ", func() {
		var vm = newVM()
		vm.SetCpuShares(1024)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMCpuSharesID))
	})
	It("should flag vm with custom properties ", func() {
		var vm = newVM()
		cps := ovirtsdk.CustomPropertySlice{}
		p1 := ovirtsdk.CustomProperty{}
		properties := []*ovirtsdk.CustomProperty{&p1}
		cps.SetSlice(properties)
		vm.SetCustomProperties(&cps)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMCustomPropertiesID))
	})
	It("should flag vm with spice display ", func() {
		var vm = newVM()
		display := ovirtsdk.Display{}
		display.SetType("spice")
		vm.SetDisplay(&display)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMDisplayTypeID))
	})
	It("should flag vm with illegal images ", func() {
		var vm = newVM()
		vm.SetHasIllegalImages(true)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMHasIllegalImagesID))
	})
	It("should flag vm with high availability priority ", func() {
		var vm = newVM()
		vm.MustHighAvailability().SetPriority(1)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMHighAvailabilityPriorityID))
	})
	It("should flag vm with IO Threads configured ", func() {
		var vm = newVM()
		io := ovirtsdk.Io{}
		io.SetThreads(4)
		vm.SetIo(&io)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMIoThreadsID))
	})
	It("should flag vm with memory balooning ", func() {
		var vm = newVM()
		memPolicy := ovirtsdk.MemoryPolicy{}
		memPolicy.SetBallooning(true)
		vm.SetMemoryPolicy(&memPolicy)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMMemoryPolicyBallooningID))
	})
	It("should flag vm with guaranteed memory ", func() {
		var vm = newVM()
		memPolicy := ovirtsdk.MemoryPolicy{}
		memPolicy.SetGuaranteed(1024)
		vm.SetMemoryPolicy(&memPolicy)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMMemoryPolicyGuaranteedID))
	})
	It("should flag vm with overcommit percent ", func() {
		var vm = newVM()
		memPolicy := ovirtsdk.MemoryPolicy{}
		memOverCommit := ovirtsdk.MemoryOverCommit{}
		memOverCommit.SetPercent(10)
		memPolicy.SetOverCommit(&memOverCommit)
		vm.SetMemoryPolicy(&memPolicy)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMMemoryPolicyOvercommitPercentID))
	})
	It("should flag vm with migration options ", func() {
		var vm = newVM()
		migration := ovirtsdk.MigrationOptions{}
		vm.SetMigration(&migration)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMMigrationID))
	})
	It("should flag vm with migration downtime ", func() {
		var vm = newVM()
		vm.SetMigrationDowntime(5)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMMigrationDowntimeID))
	})
	It("should flag vm with NUMA tune mode ", func() {
		var vm = newVM()
		vm.SetNumaTuneMode("strict")

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMNumaTuneModeID))
	})
	It("should flag vm with origin == kubevirt ", func() {
		var vm = newVM()
		vm.SetOrigin("kubevirt")

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMOriginID))
	})
	table.DescribeTable("should flag VM with illegal random number generator source", func(source string) {
		vm := newVM()
		vm.MustRngDevice().SetSource(ovirtsdk.RngSource(source))

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMRngDeviceSourceID))
	},
		table.Entry("hwrng", "hwrng"),
		table.Entry("random", "random"),

		table.Entry("garbage", "safdwlfkq332"),
		table.Entry("empty", ""),
	)
	It("should flag vm with sound card enabled", func() {
		var vm = newVM()
		vm.SetSoundcardEnabled(true)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMSoundcardEnabledID))
	})
	It("should flag vm with start paused enabled", func() {
		var vm = newVM()
		vm.SetStartPaused(true)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMStartPausedID))
	})
	It("should flag vm with storage error resume behaviour specified", func() {
		var vm = newVM()
		vm.SetStorageErrorResumeBehaviour("auto_resume")

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMStorageErrorResumeBehaviourID))
	})
	It("should flag vm with tunnel migration enabled", func() {
		var vm = newVM()
		vm.SetTunnelMigration(true)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMTunnelMigrationID))
	})
	It("should flag vm with USB enabled", func() {
		var vm = newVM()
		usb := ovirtsdk.Usb{}
		usb.SetEnabled(true)
		vm.SetUsb(&usb)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMUsbID))
	})
	It("should flag vm with spice console configured", func() {
		var vm = newVM()
		consoles := []*ovirtsdk.GraphicsConsole{newGraphicsConsole("spice"), newGraphicsConsole("vnc")}
		vm.MustGraphicsConsoles().SetSlice(consoles)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMGraphicConsolesID))
	})
	It("should flag vm's host devices", func() {
		var vm = newVM()
		devices := []*ovirtsdk.HostDevice{&ovirtsdk.HostDevice{}}
		hostDevices := ovirtsdk.HostDeviceSlice{}
		hostDevices.SetSlice(devices)
		vm.SetHostDevices(&hostDevices)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMHostDevicesID))
	})
	It("should flag vm's reported devices", func() {
		var vm = newVM()
		devices := []*ovirtsdk.ReportedDevice{&ovirtsdk.ReportedDevice{}}
		reportedDevices := ovirtsdk.ReportedDeviceSlice{}
		reportedDevices.SetSlice(devices)
		vm.SetReportedDevices(&reportedDevices)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMReportedDevicesID))
	})
	It("should flag vm with quota", func() {
		var vm = newVM()
		quota := ovirtsdk.Quota{}
		quota.SetId("quota_id")
		vm.SetQuota(&quota)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMQuotaID))
	})
	It("should flag illegal watchdog - diag288", func() {
		var vm = newVM()
		watchdog := ovirtsdk.Watchdog{}
		watchdog.SetModel("diag288")
		vm.MustWatchdogs().SetSlice([]*ovirtsdk.Watchdog{&watchdog})

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMWatchdogsID))
	})
	It("should flag CD ROM with image stored in non-data domain", func() {
		var vm = newVM()
		storageDomain := ovirtsdk.StorageDomain{}
		storageDomain.SetType("iso")
		file := ovirtsdk.File{}
		file.SetStorageDomain(&storageDomain)
		cdrom := ovirtsdk.Cdrom{}
		cdrom.SetId("cd_id")
		cdrom.SetFile(&file)
		cdroms := []*ovirtsdk.Cdrom{&cdrom}
		vm.MustCdroms().SetSlice(cdroms)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMCdromsID))
	})
	It("should flag floppies", func() {
		var vm = newVM()

		floppies := []*ovirtsdk.Floppy{&ovirtsdk.Floppy{}}
		floppySlice := ovirtsdk.FloppySlice{}
		floppySlice.SetSlice(floppies)
		vm.SetFloppies(&floppySlice)

		failures := validators.ValidateVM(vm)

		Expect(failures).To(HaveLen(1))
		Expect(failures[0].ID).To(Equal(validators.VMFloppiesID))
	})
})

func newGraphicsConsole(protocol string) *ovirtsdk.GraphicsConsole {
	console := ovirtsdk.GraphicsConsole{}
	console.SetProtocol(ovirtsdk.GraphicsType(protocol))
	return &console
}

func newVM() *ovirtsdk.Vm {
	return newVMWithStatusControl(true)
}

func newVMWithStatusControl(withStatus bool) *ovirtsdk.Vm {
	vm := ovirtsdk.Vm{}
	if withStatus {
		vm.SetStatus(ovirtsdk.VMSTATUS_UP)
	}
	bios := ovirtsdk.Bios{}
	bios.SetType("q35_sea_bios")
	vm.SetBios(&bios)

	cpu := ovirtsdk.Cpu{}
	cpu.SetArchitecture("x86_64")
	cpuTune := ovirtsdk.CpuTune{}
	pinSlice := ovirtsdk.VcpuPinSlice{}
	pins := []*ovirtsdk.VcpuPin{newCPUPin(0, "0"), newCPUPin(1, "1")}
	pinSlice.SetSlice(pins)
	cpuTune.SetVcpuPins(&pinSlice)
	cpu.SetCpuTune(&cpuTune)
	vm.SetCpu(&cpu)

	ha := ovirtsdk.HighAvailability{}
	ha.SetEnabled(true)
	vm.SetHighAvailability(&ha)

	vm.SetOrigin("ovirt")

	rng := ovirtsdk.RngDevice{}
	rng.SetSource("urandom")

	gfxConsoles := ovirtsdk.GraphicsConsoleSlice{}
	consoles := []*ovirtsdk.GraphicsConsole{newGraphicsConsole("vnc")}
	gfxConsoles.SetSlice(consoles)
	vm.SetGraphicsConsoles(&gfxConsoles)
	vm.SetRngDevice(&rng)

	watchdog := ovirtsdk.Watchdog{}
	watchdog.SetModel("i6300esb")
	wdSlice := ovirtsdk.WatchdogSlice{}
	wdSlice.SetSlice([]*ovirtsdk.Watchdog{&watchdog})
	vm.SetWatchdogs(&wdSlice)

	storageDomain := ovirtsdk.StorageDomain{}
	storageDomain.SetType("data")
	file := ovirtsdk.File{}
	file.SetStorageDomain(&storageDomain)
	cdrom := ovirtsdk.Cdrom{}
	cdrom.SetFile(&file)
	cdroms := []*ovirtsdk.Cdrom{&cdrom}
	cdromSlice := ovirtsdk.CdromSlice{}
	cdromSlice.SetSlice(cdroms)
	vm.SetCdroms(&cdromSlice)

	return &vm
}
func newCPUPin(cpu int64, cpuSet string) *ovirtsdk.VcpuPin {
	pin := ovirtsdk.VcpuPin{}
	pin.SetVcpu(cpu)
	pin.SetCpuSet(cpuSet)
	return &pin
}
