package intializer

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/vm_port_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	"strconv"
)

// SynchronizeVmPorts synchronizes the VM ports from the config with the database.
// This includes deleting ports that are not in the config and creating ports that are in the config but not in the database.
func SynchronizeVmPorts() error {

	summary := make(map[string]int)

	type PortRange struct {
		Start int
		End   int
	}

	ranges := make(map[string]PortRange)

	for _, zone := range config.Config.Zones {
		if zone.PortRange.End != 0 {
			ranges[zone.Name] = PortRange{
				Start: zone.PortRange.Start,
				End:   zone.PortRange.End,
			}
		}
	}

	for zone, portRange := range ranges {
		// Delete all ports that are not in the config
		portsOutsideRange, err := vm_port_repo.New().WithZone(zone).ExcludePortRange(portRange.Start, portRange.End).List()
		if err != nil {
			log.Fatalln(err)
		}

		notLeasedPorts := make([]model.VmPort, 0)
		leasedPorts := make([]model.VmPort, 0)
		for _, port := range portsOutsideRange {
			if port.Lease != nil {
				leasedPorts = append(leasedPorts, port)
			} else {
				notLeasedPorts = append(notLeasedPorts, port)
			}
		}

		if len(leasedPorts) > 0 {
			for _, port := range leasedPorts {
				log.Printf("Port %d is leased by vm %s. this port will remain, but should be deleted", port.PublicPort, port.Lease.VmID)
			}
		}

		for _, port := range notLeasedPorts {
			err = vm_port_repo.New().Erase(port.PublicPort, port.Zone)
			if err != nil {
				return err
			}
		}

		existingPorts, err := vm_port_repo.New().WithZone(zone).IncludePortRange(portRange.Start, portRange.End).Count()
		if err != nil {
			return err
		}

		noInserted := 0
		if existingPorts != portRange.End-portRange.Start {
			noInserted, err = vm_port_repo.New().CreateIfNotExists(portRange.Start, portRange.End, zone)
			if err != nil {
				return err
			}
		}

		summary[zone] = noInserted
	}

	for zone, noInserted := range summary {
		log.Printf(" - " + zone + ": inserted " + strconv.Itoa(noInserted) + " new ports")
	}

	return nil
}
