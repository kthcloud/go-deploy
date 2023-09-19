package helpers

import (
	"fmt"
	"go-deploy/models/sys/enviroment"
	vmModel "go-deploy/models/sys/vm"
	gpuModel "go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"strings"
	"time"
)

func WithCsClient(zone *enviroment.VmZone) (*cs.Client, error) {
	return cs.New(&cs.ClientConf{
		URL:         conf.Env.CS.URL,
		ApiKey:      conf.Env.CS.ApiKey,
		Secret:      conf.Env.CS.Secret,
		IpAddressID: zone.IpAddressID,
		ProjectID:   zone.ProjectID,
		NetworkID:   zone.NetworkID,
		ZoneID:      zone.ZoneID,
	})
}

func CreateNonRootPortForwardingRuleName(name, networkName string) string {
	return fmt.Sprintf("%s-%s", name, networkName)
}

func CreateDeployTags(name string, deployName string) []csModels.Tag {
	return []csModels.Tag{
		{Key: "name", Value: name},
		{Key: "managedBy", Value: conf.Env.Manager},
		{Key: "deployName", Value: deployName},
		{Key: "createdAt", Value: time.Now().Format(time.RFC3339)},
	}
}

func CreateExtraConfig(gpu *gpuModel.GPU) string {
	data := fmt.Sprintf(`
<devices> <hostdev mode='subsystem' type='pci' managed='yes'> <driver name='vfio' />
	<source> <address domain='0x0000' bus='0x%s' slot='0x00' function='0x0' /> </source> 
	<alias name='nvidia0' /> <address type='pci' domain='0x0000' bus='0x00' slot='0x00' function='0x0' /> 
</hostdev> </devices>`, gpu.Data.Bus)

	data = strings.Replace(data, "\n", "", -1)
	data = strings.Replace(data, "\t", "", -1)

	return data
}

func HasExtraConfig(vm *vmModel.VM) bool {
	return vm.Subsystems.CS.VM.ExtraConfig != "" && vm.Subsystems.CS.VM.ExtraConfig != "none"
}

func GetRequiredHost(gpuID string) (*string, error) {
	gpu, err := gpuModel.New().GetByID(gpuID)
	if err != nil {
		return nil, err
	}

	if gpu.Host == "" {
		return nil, fmt.Errorf("no host found for gpu %s", gpu.ID)
	}

	return &gpu.Host, nil
}
