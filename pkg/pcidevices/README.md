# Go-P4Pack pcidevices Package

This package helps to query DPDK compatibles pci devices and to bind/unbind drivers to pci devices in the Linux Kernel.
With this package the commands of dpdk-devbind.py can be re-created in Go.

## API

The package exports the following API functions that return a PciDevice pointer or an array of PciDevice pointers:

    New(id string) (*PciDevice, error)                             // input supports PCI identifier string or NIC name
    NewDeviceByPciID(pciID string) (*PciDevice, error)
    NewDeviceByNicName(nicName string) (*PciDevice, error)
    GetPciDevices(classFilter ClassesFilter) ([]*PciDevice, error) // returns a filtered list of current active PCI devs

The PciDevice supports the following methods:

    PciDevice.Bind(driver string) error
    PciDevice.Unbind() error
    PciDevice.Probe() error
    PciDevice.CurrentDriver() (string, error)
    PciDevice.ID() string
    PciDevice.String() string

ClassFilter values:

    AllDevices
    NetworkDevices
    BasebandDevices
    CryptoDevices
    DmaDevices
    EventdevDevices
    MempoolDevices
    CompressDevices
    RegexDevices
    MiscDevices

## Extra exported low level API functions

Lower Level API functions:

    BindPci(devID, driver, vendor, device string) (string, error)
    UnbindPci(devID, driver string) error
    ProbePci(devID string) (string, error)
    GetCurrentPciDriver(devID string) (string, error)
    IsModuleLoaded(driver string) bool