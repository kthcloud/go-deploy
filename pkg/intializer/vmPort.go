package intializer

import (
	"go-deploy/models/sys/vm_port"
	"go-deploy/pkg/config"
	"log"
	"strconv"
)

// SynchronizeVmPorts synchronizes the VM ports from the database to the database.
// This includes deleting ports that are not in the config and creating ports that are in the config but not in the database.
func SynchronizeVmPorts() {

	summary := make(map[string]int)

	for _, zone := range config.Config.VM.Zones {
		// Delete all ports that are not in the config
		portsOutsideRange, err := vm_port.New().WithZone(zone.Name).ExcludePortRange(zone.PortRange.Start, zone.PortRange.End).List()
		if err != nil {
			log.Fatalln(err)
		}

		notLeasedPorts := make([]vm_port.VmPort, 0)
		leasedPorts := make([]vm_port.VmPort, 0)
		for _, port := range portsOutsideRange {
			if port.Lease != nil {
				leasedPorts = append(leasedPorts, port)
			} else {
				notLeasedPorts = append(notLeasedPorts, port)
			}
		}

		if len(leasedPorts) > 0 {
			for _, port := range leasedPorts {
				log.Printf("port %d is leased by vm %s. this port will remain, but should be deleted", port.PublicPort, port.Lease.VmID)
			}
		}

		for _, port := range notLeasedPorts {
			err = vm_port.New().Erase(port.PublicPort, port.Zone)
			if err != nil {
				log.Fatalln(err)
			}
		}

		existingPorts, err := vm_port.New().WithZone(zone.Name).IncludePortRange(zone.PortRange.Start, zone.PortRange.End).Count()
		if err != nil {
			log.Fatalln(err)
		}

		noInserted := 0
		if existingPorts != zone.PortRange.End-zone.PortRange.Start {
			noInserted, err = vm_port.New().CreateIfNotExists(zone.PortRange.Start, zone.PortRange.End, zone.Name)
			if err != nil {
				log.Fatalln(err)
			}
		}

		summary[zone.Name] = noInserted
	}

	summaryString := ""
	for zone, noInserted := range summary {
		summaryString += "\t- " + zone + ": inserted " + strconv.Itoa(noInserted) + " new ports\n"
	}

	log.Printf("synchronized vm ports:\n%s", summaryString)
}
