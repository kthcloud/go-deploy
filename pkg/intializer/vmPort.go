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

	// Method synchronizes the port in both VM zones and Deployment zones
	// This is a temporary solution while we support both CloudStack VMs (v1) and KubeVirt VMs (v2)

	summary := make(map[string]int)

	type PortRange struct {
		Start int
		End   int
	}

	ranges := make(map[string]PortRange)

	for _, zone := range config.Config.VM.Zones {
		ranges[zone.Name] = PortRange{
			Start: zone.PortRange.Start,
			End:   zone.PortRange.End,
		}
	}

	for _, zone := range config.Config.Deployment.Zones {
		if zone.PortRange.End != 0 {
			ranges[zone.Name] = PortRange{
				Start: zone.PortRange.Start,
				End:   zone.PortRange.End,
			}
		}
	}

	for zone, portRange := range ranges {
		// Delete all ports that are not in the config
		portsOutsideRange, err := vm_port.New().WithZone(zone).ExcludePortRange(portRange.Start, portRange.End).List()
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

		existingPorts, err := vm_port.New().WithZone(zone).IncludePortRange(portRange.Start, portRange.End).Count()
		if err != nil {
			log.Fatalln(err)
		}

		noInserted := 0
		if existingPorts != portRange.End-portRange.Start {
			noInserted, err = vm_port.New().CreateIfNotExists(portRange.Start, portRange.End, zone)
			if err != nil {
				log.Fatalln(err)
			}
		}

		summary[zone] = noInserted
	}

	summaryString := ""
	for zone, noInserted := range summary {
		summaryString += "\t- " + zone + ": inserted " + strconv.Itoa(noInserted) + " new ports\n"
	}

	log.Printf("synchronized vm ports:\n%s", summaryString)
}
