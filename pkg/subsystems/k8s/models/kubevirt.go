package models

import "fmt"

type PermittedHostDevices struct {
	PciHostDevices []PciHostDevice
}

type PciHostDevice struct {
	PciVendorSelector string
	ResourceName      string
}

func CreatePciVendorSelector(vendorID, deviceID string) string {
	return fmt.Sprintf("%s:%s", vendorID, deviceID)
}
