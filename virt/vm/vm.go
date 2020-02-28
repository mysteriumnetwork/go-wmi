package vm

import (
	"fmt"
	"go-wmi/wmi"
	"runtime"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/pkg/errors"
)

// NewVMManager returns a new Manager type
func NewVMManager() (*Manager, error) {
	w, err := wmi.NewConnection(".", `root\virtualization\v2`)
	if err != nil {
		return nil, err
	}

	// Get virtual machine management service
	svc, err := w.GetOne(VMManagementService, []string{}, []wmi.Query{})
	if err != nil {
		return nil, err
	}

	sw := &Manager{
		con: w,
		svc: svc,
	}
	return sw, nil
}

// Manager manages a VM switch
type Manager struct {
	con *wmi.WMI
	svc *wmi.Result
}

// GetVM returns the virtual machine identified by instanceID
func (m *Manager) GetVM(instanceID string) (*VirtualMachine, error) {
	fields := []string{}
	qParams := []wmi.Query{
		&wmi.AndQuery{
			wmi.QueryFields{
				Key:   "VirtualSystemType",
				Value: VirtualSystemTypeRealized,
				Type:  wmi.Equals},
		},
		&wmi.AndQuery{
			wmi.QueryFields{
				Key:   "VirtualSystemIdentifier",
				Value: instanceID,
				Type:  wmi.Equals},
		},
	}

	result, err := m.con.Gwmi(VirtualSystemSettingDataClass, fields, qParams)
	if err != nil {
		return nil, errors.Wrap(err, "VirtualSystemSettingDataClass")
	}

	vssd, err := result.ItemAtIndex(0)
	if err != nil {
		return nil, errors.Wrap(err, "fetching element")
	}
	cs, err := vssd.Get("associators_", nil, ComputerSystemClass)
	if err != nil {
		return nil, errors.Wrap(err, "getting ComputerSystemClass")
	}
	elem, err := cs.Elements()
	if err != nil || len(elem) == 0 {
		return nil, errors.Wrap(err, "getting elements")
	}
	return &VirtualMachine{
		mgr:                m,
		activeSettingsData: vssd,
		computerSystem:     elem[0],
	}, nil
}

// ListVM returns a list of virtual machines
func (m *Manager) ListVM() ([]*VirtualMachine, error) {
	fields := []string{}
	qParams := []wmi.Query{
		&wmi.AndQuery{
			wmi.QueryFields{
				Key:   "VirtualSystemType",
				Value: VirtualSystemTypeRealized,
				Type:  wmi.Equals},
		},
	}

	result, err := m.con.Gwmi(VirtualSystemSettingDataClass, fields, qParams)
	if err != nil {
		return nil, errors.Wrap(err, "VirtualSystemSettingDataClass")
	}

	elements, err := result.Elements()
	if err != nil {
		return nil, errors.Wrap(err, "Elements")
	}
	vms := make([]*VirtualMachine, len(elements))
	for idx, val := range elements {
		cs, err := val.Get("associators_", nil, ComputerSystemClass)
		if err != nil {
			return nil, errors.Wrap(err, "getting ComputerSystemClass")
		}
		elem, err := cs.Elements()
		if err != nil || len(elem) == 0 {
			return nil, errors.Wrap(err, "getting elements")
		}
		vms[idx] = &VirtualMachine{
			mgr:                m,
			activeSettingsData: val,
			computerSystem:     elem[0],
		}
	}
	return vms, nil
}

// CreateVM creates a new virtual machine
func (m *Manager) CreateVM(name string, memoryMB int64, cpus int, limitCPUFeatures bool, notes []string, generation GenerationType) (*VirtualMachine, error) {
	vmSettingsDataInstance, err := m.con.Get(VirtualSystemSettingDataClass)
	if err != nil {
		return nil, err
	}

	newVMInstance, err := vmSettingsDataInstance.Get("SpawnInstance_")
	if err != nil {
		return nil, errors.Wrap(err, "calling SpawnInstance_")
	}

	if err := newVMInstance.Set("ElementName", name); err != nil {
		return nil, errors.Wrap(err, "Set ElementName")
	}
	if err := newVMInstance.Set("VirtualSystemSubType", string(generation)); err != nil {
		return nil, errors.Wrap(err, "Set VirtualSystemSubType")
	}
	if notes != nil && len(notes) > 0 {
		// Don't ask...
		// Well, ok...if you must. The Msvm_VirtualSystemSettingData has a Notes
		// property of type []string. But in reality, it only cares about the first
		// element of that array. So we join the notes into one newline delimited
		// string, and set that as the first and only element in a new []string{}
		vmNotes := []string{strings.Join(notes, "\n")}
		if err := newVMInstance.Set("Notes", vmNotes); err != nil {
			return nil, errors.Wrap(err, "Set Notes")
		}
	}

	vmText, err := newVMInstance.GetText(1)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get VM instance XML")
	}

	jobPath := ole.VARIANT{}
	resultingSystem := ole.VARIANT{}
	jobState, err := m.svc.Get("DefineSystem", vmText, nil, nil, &resultingSystem, &jobPath)
	if err != nil {
		return nil, errors.Wrap(err, "calling DefineSystem")
	}
	if jobState.Value().(int32) == wmi.JobStatusStarted {
		err := wmi.WaitForJob(jobPath.Value().(string))
		if err != nil {
			return nil, errors.Wrap(err, "waiting for job")
		}
	}

	// The resulting system  value is always a string containing the
	// location of the newly created resource
	locationURI := resultingSystem.Value().(string)
	loc, err := wmi.NewLocation(locationURI)
	if err != nil {
		return nil, errors.Wrap(err, "getting location")
	}

	result, err := loc.GetResult()
	if err != nil {
		return nil, errors.Wrap(err, "getting result")
	}

	// The name field of the returning class is actually the InstanceID...
	id, err := result.GetProperty("Name")
	if err != nil {
		return nil, errors.Wrap(err, "fetching VM ID")
	}

	vm, err := m.GetVM(id.Value().(string))
	if err != nil {
		return nil, errors.Wrap(err, "fetching VM")
	}

	if err := vm.SetMemory(memoryMB); err != nil {
		return nil, errors.Wrap(err, "setting memory limit")
	}

	if err := vm.SetNumCPUs(cpus); err != nil {
		return nil, errors.Wrap(err, "setting CPU limit")
	}

	return vm, nil
}

// Release closes the WMI connection associated with this
// Manager
func (m *Manager) Release() {
	m.con.Close()
}

// VirtualMachine represents a single virtual machine
type VirtualMachine struct {
	mgr *Manager

	activeSettingsData *wmi.Result
	computerSystem     *wmi.Result
}

// Name returns the current name of this virtual machine
func (v *VirtualMachine) Name() (string, error) {
	name, err := v.computerSystem.GetProperty("ElementName")
	if err != nil {
		return "", errors.Wrap(err, "getting ElementName")
	}
	return name.Value().(string), nil
}

// ID returns the instance ID of this Virtual machine
func (v *VirtualMachine) ID() (string, error) {
	id, err := v.activeSettingsData.GetProperty("VirtualSystemIdentifier")
	if err != nil {
		return "", errors.Wrap(err, "fetching VM ID")
	}
	return id.Value().(string), nil
}

// AttachDisks attaches the supplied disks, to this virtual machine
func (v *VirtualMachine) AttachDisks(disks []string) error {
	return nil
}

// SetBootOrder sets the VM boot order
func (v *VirtualMachine) SetBootOrder(bootOrder []BootOrderType) error {
	// bootOrder := []int32{
	// 	int32(BootHDD),
	// 	int32(BootPXE),
	// 	int32(BootCDROM),
	// 	int32(BootFloppy),
	// }

	// if err := newVMInstance.Set("BootOrder", bootOrder); err != nil {
	// 	return nil, errors.Wrap(err, "Set BootOrder")
	// }
	return nil
}

// SetMemory sets the virtual machine memory allocation
func (v *VirtualMachine) SetMemory(memoryMB int64) error {
	return nil
}

// SetNumCPUs sets the number of CPU cores on the VM
func (v *VirtualMachine) SetNumCPUs(cpus int) error {
	hostCpus := runtime.NumCPU()
	if hostCpus < cpus {
		return fmt.Errorf("Number of cpus exceeded available host resources")
	}
	return nil
}
