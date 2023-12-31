package intializer

import (
	"go-deploy/models/sys/vmPort"
	"go-deploy/pkg/config"
	"log"
	"strconv"
)

func SynchronizeVmPorts() {

	summary := make(map[string]int)

	for _, zone := range config.Config.VM.Zones {
		// Delete all ports that are not in the config
		portsOutsideRange, err := vmPort.New().WithZone(zone.Name).ExcludePortRange(zone.PortRange.Start, zone.PortRange.End).List()
		if err != nil {
			log.Fatalln(err)
		}

		notLeasedPorts := make([]vmPort.VmPort, 0)
		leasedPorts := make([]vmPort.VmPort, 0)
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
			err = vmPort.New().Delete(port.PublicPort, port.Zone)
			if err != nil {
				log.Fatalln(err)
			}
		}

		existingPorts, err := vmPort.New().WithZone(zone.Name).IncludePortRange(zone.PortRange.Start, zone.PortRange.End).Count()
		if err != nil {
			log.Fatalln(err)
		}

		noInserted := 0
		if existingPorts != zone.PortRange.End-zone.PortRange.Start {
			noInserted, err = vmPort.New().CreateIfNotExists(zone.PortRange.Start, zone.PortRange.End, zone.Name)
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
