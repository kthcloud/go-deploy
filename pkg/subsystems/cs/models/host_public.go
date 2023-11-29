package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"math"
	"strconv"
	"strings"
	"time"
)

type HostPublic struct {
	ID            string    `bson:"id"`
	Name          string    `bson:"name"`
	State         string    `bson:"state"`
	ResourceState string    `bson:"resourceState"`
	CpuCoresUsed  int       `bson:"cpuCoresUsed"`
	CpuCoresTotal int       `bson:"cpuCoresTotal"`
	RamUsed       int       `bson:"ramUsed"`
	RamTotal      int       `bson:"ramTotal"`
	CreatedAt     time.Time `bson:"createdAt"`
}

func CreateHostPublicFromGet(host *cloudstack.Host) *HostPublic {
	ramUsedInGB := int(host.Memoryused / 1024 / 1024 / 1024)
	ramInGB := int(host.Memorytotal / 1024 / 1024 / 1024)

	if ramUsedInGB > ramInGB {
		ramUsedInGB = ramInGB
	}

	cpuAllocatedPercentage, err := strconv.ParseFloat(strings.Trim(host.Cpuallocated, "%"), 64)
	if err != nil {
		cpuAllocatedPercentage = 0
	}

	cpuAllocatedInt := int(float64(host.Cpunumber) * (math.Floor(cpuAllocatedPercentage) / 100))

	if cpuAllocatedInt > host.Cpunumber {
		cpuAllocatedInt = host.Cpunumber
	}

	return &HostPublic{
		ID:            host.Id,
		Name:          host.Name,
		State:         host.State,
		ResourceState: host.Resourcestate,
		CpuCoresUsed:  cpuAllocatedInt,
		CpuCoresTotal: host.Cpunumber,
		RamUsed:       ramUsedInGB,
		RamTotal:      ramInGB,
		CreatedAt:     formatCreatedAt(host.Created),
	}
}
