package vhd

type ProcessorArchitecture int

const (
	VHD_HEADER_SIZE_FIX          = 512
	VHD_BAT_ENTRY_SIZE           = 4
	VHD_DYNAMIC_DISK_HEADER_SIZE = 1024
	VHD_HEADER_SIZE_DYNAMIC      = 512
	VHD_FOOTER_SIZE_DYNAMIC      = 512
	VHD_BLK_SIZE_OFFSET          = 544
	VHD_SIGNATURE                = "conectix"
	VHDX_SIGNATURE               = "vhdxfile"

	HYPERV_VM_STATE_ENABLED       = 2
	HYPERV_VM_STATE_DISABLED      = 3
	HYPERV_VM_STATE_SHUTTING_DOWN = 4
	HYPERV_VM_STATE_REBOOT        = 10
	HYPERV_VM_STATE_PAUSED        = 32768
	HYPERV_VM_STATE_SUSPENDED     = 32769

	MMX    = 3
	NX     = 12
	PAE    = 9
	RDTSC  = 8
	SLAT   = 20
	SSE3   = 13
	VMX    = 21
	SSE    = 6
	SSE2   = 10
	XSAVE  = 17
	DDDNOW = 7

	WMI_JOB_STATUS_STARTED  = 4096
	WMI_JOB_STATE_RUNNING   = 4
	WMI_JOB_STATE_COMPLETED = 7

	VM_SUMMARY_NUM_PROCS     = 4
	VM_SUMMARY_ENABLED_STATE = 100
	VM_SUMMARY_MEMORY_USAGE  = 103
	VM_SUMMARY_UPTIME        = 105

	IDE_DISK        = "VHD"
	IDE_DISK_FORMAT = IDE_DISK
	IDE_DVD         = "DVD"
	IDE_DVD_FORMAT  = "ISO"

	DISK_FORMAT_VHD  = "VHD"
	DISK_FORMAT_VHDX = "VHDX"

	VHD_TYPE_FIXED   = 2
	VHD_TYPE_DYNAMIC = 3

	SCSI_CONTROLLER_SLOTS_NUMBER = 64
)

const (
	X86 ProcessorArchitecture = iota
	MIPS
	Alpha
	PowerPC
	ARM
	Itanium
	X64
)
