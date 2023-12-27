package cs_service

import (
	"fmt"
	gpuModels "go-deploy/models/sys/gpu"
	vmModels "go-deploy/models/sys/vm"
	"strings"
)

func CreateExtraConfig(gpu *gpuModels.GPU) string {
	data := fmt.Sprintf(`
<devices> <hostdev mode='subsystem' type='pci' managed='yes'> <driver name='vfio' />
	<source> <address domain='0x0000' bus='0x%s' slot='0x00' function='0x0' /> </source> 
	<alias name='nvidia0' /> <address type='pci' domain='0x0000' bus='0x00' slot='0x00' function='0x0' /> 
</hostdev> </devices>`, gpu.Data.Bus)

	data = strings.Replace(data, "\n", "", -1)
	data = strings.Replace(data, "\t", "", -1)

	return data
}

func HasExtraConfig(vm *vmModels.VM) bool {
	return vm.Subsystems.CS.VM.ExtraConfig != "" && vm.Subsystems.CS.VM.ExtraConfig != "none"
}
