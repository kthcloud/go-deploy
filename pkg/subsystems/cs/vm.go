package cs

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"go-deploy/pkg/subsystems/cs/commands"
	"go-deploy/pkg/subsystems/cs/models"
	"gopkg.in/yaml.v3"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

// ReadVM reads the VM from CloudStack.
func (client *Client) ReadVM(id string) (*models.VmPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cs vm %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("cs vm id not supplied when reading. assuming it was deleted")
		return nil, nil
	}

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return nil, makeError(err)
		}
	}

	var public *models.VmPublic
	if vm != nil {
		public = models.CreateVmPublicFromGet(vm)
	}

	return public, nil
}

// CreateVM creates the VM in CloudStack.
func (client *Client) CreateVM(public *models.VmPublic) (*models.VmPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cs vm %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("cs vm name not supplied when creating. assuming it was deleted")
		return nil, nil
	}

	listVmParams := client.CsClient.VirtualMachine.NewListVirtualMachinesParams()
	listVmParams.SetName(public.Name)
	listVmParams.SetProjectid(client.ProjectID)

	listVms, err := client.CsClient.VirtualMachine.ListVirtualMachines(listVmParams)
	if err != nil {
		return nil, makeError(err)
	}

	var id string
	if listVms.Count == 0 {
		createVmParams := client.CsClient.VirtualMachine.NewDeployVirtualMachineParams(
			client.CustomServiceOfferingID,
			client.TemplateID,
			client.ZoneID,
		)

		createVmParams.SetDetails(map[string]string{
			DetailsCpuCores: fmt.Sprintf("%d", public.CpuCores),
			DetailsCpuSpeed: "1",
			DetailsRAM:      fmt.Sprintf("%d", public.RAM*1024),
		})
		createVmParams.SetName(public.Name)
		createVmParams.SetDisplayname(public.Name)
		createVmParams.SetNetworkids([]string{client.RootNetworkID})
		createVmParams.SetProjectid(client.ProjectID)
		createVmParams.SetExtraconfig(public.ExtraConfig)

		userData := createUserData(public.Name, client.userSshPublicKey, client.adminSshPublicKey)
		userData = "#cloud-config\n" + userData
		userDataB64 := base64.StdEncoding.EncodeToString([]byte(userData))

		createVmParams.SetUserdata(userDataB64)

		created, err := client.CsClient.VirtualMachine.DeployVirtualMachine(createVmParams)
		if err != nil {
			return nil, makeError(err)
		}

		id = created.Id
	} else {
		id = listVms.VirtualMachines[0].Id
	}

	err = client.AssertVmTags(id, public.Tags)
	if err != nil {
		return nil, makeError(err)
	}

	return client.ReadVM(id)
}

// UpdateVM updates the VM in CloudStack.
func (client *Client) UpdateVM(public *models.VmPublic) (*models.VmPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update cs vm %s. details: %w", public.ID, err)
	}

	if public.ID == "" {
		log.Println("cs vm id not supplied when updating. assuming it was deleted")
		return nil, nil
	}

	vm, err := client.ReadVM(public.ID)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		log.Println("cs vm", public.ID, "not found when updating. assuming it was deleted")
		return nil, nil
	}

	if vm.Name != public.Name || vm.ExtraConfig != public.ExtraConfig {
		params := client.CsClient.VirtualMachine.NewUpdateVirtualMachineParams(public.ID)
		params.SetName(public.Name)
		params.SetDisplayname(public.Name)

		if public.ExtraConfig == "" {
			params.SetExtraconfig(url.QueryEscape("none"))
		} else {
			params.SetExtraconfig(url.QueryEscape(public.ExtraConfig))
		}

		_, err = client.CsClient.VirtualMachine.UpdateVirtualMachine(params)
		if err != nil {
			return nil, makeError(err)
		}
	}

	vmStatus, err := client.GetVmStatus(public.ID)
	if err != nil {
		return nil, makeError(err)
	}

	if vmStatus == "Stopped" {
		params := client.CsClient.VirtualMachine.NewChangeServiceForVirtualMachineParams(public.ID, client.CustomServiceOfferingID)
		params.SetDetails(map[string]string{
			DetailsCpuCores: fmt.Sprintf("%d", public.CpuCores),
			DetailsCpuSpeed: "1",
			DetailsRAM:      fmt.Sprintf("%d", public.RAM*1024),
		})
		_, err = client.CsClient.VirtualMachine.ChangeServiceForVirtualMachine(params)
		if err != nil {
			return nil, makeError(err)
		}
	} else {
		log.Println("cs vm", public.ID, "is not stopped when updating service offering. skipping updating service offering")
	}

	err = client.AssertVmTags(public.ID, public.Tags)
	if err != nil {
		return nil, makeError(err)
	}

	return client.ReadVM(public.ID)
}

// DeleteVM deletes the VM in CloudStack.
func (client *Client) DeleteVM(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete cs vm %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("cs vm id not supplied when deleting. assuming it was deleted")
		return nil
	}

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return makeError(err)
		}
	}

	if vm == nil {
		log.Println("cs vm", id, "not found when deleting. assuming it was deleted")
		return nil
	}

	if vm.State == "Stopping" || vm.State == "DestroyRequested" || vm.State == "Expunging" {
		return nil
	}

	params := client.CsClient.VirtualMachine.NewDestroyVirtualMachineParams(id)

	params.SetExpunge(true)

	_, err = client.CsClient.VirtualMachine.DestroyVirtualMachine(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// GetVmStatus gets the VM status in CloudStack.
func (client *Client) GetVmStatus(id string) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get cs vm %s status. details: %w", id, err)
	}

	if id == "" {
		log.Println("cs vm id not supplied when getting status. assuming it was deleted")
		return "", nil
	}

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(id)
	if err != nil {
		errString := err.Error()
		if !strings.Contains(errString, "No match found for") {
			return "", makeError(err)
		}

		return "", nil
	}

	return vm.State, nil
}

// ListAllStatus lists all the VM status in CloudStack.
func (client *Client) ListAllStatus() (map[string]string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to list all cs vm status. details: %w", err)
	}

	params := client.CsClient.VirtualMachine.NewListVirtualMachinesMetricsParams()
	params.SetProjectid(client.ProjectID)
	params.SetListall(true)

	vms, err := client.CsClient.VirtualMachine.ListVirtualMachinesMetrics(params)
	if err != nil {
		return nil, makeError(err)
	}

	statuses := make(map[string]string)
	for _, vm := range vms.VirtualMachinesMetrics {
		statuses[vm.Id] = vm.State
	}

	return statuses, nil
}

// DoVmCommand executes the VM command in CloudStack.
func (client *Client) DoVmCommand(id string, requiredHost *string, command commands.Command) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to execute csVM command to csVM %s. details: %w", id, err)
	}

	csVM, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(id)
	if err != nil {
		return makeError(err)
	}

	switch command {
	case commands.Start:
		if csVM.State != "Running" && csVM.State != "Starting" && csVM.State != "Stopping" && csVM.State != "Migrating" {
			params := client.CsClient.VirtualMachine.NewStartVirtualMachineParams(id)
			if requiredHost != nil {
				host, _, err := client.CsClient.Host.GetHostByName(*requiredHost)
				if err != nil {
					return makeError(fmt.Errorf("failed to get host %s. details: %w", *requiredHost, err))
				}

				params.SetHostid(host.Id)
			} else {
				// Temporarily randomize host to start on
				hosts, err := client.CsClient.Host.ListHosts(client.CsClient.Host.NewListHostsParams())
				if err != nil {
					return makeError(err)
				}

				host := hosts.Hosts[rand.Intn(len(hosts.Hosts))]
				params.SetHostid(host.Id)
			}

			_, err = client.CsClient.VirtualMachine.StartVirtualMachine(params)
			if err != nil {
				return makeError(err)
			}
		}
	case commands.Stop:
		if csVM.State != "Stopped" && csVM.State != "Stopping" && csVM.State != "Starting" && csVM.State != "Migrating" {
			params := client.CsClient.VirtualMachine.NewStopVirtualMachineParams(id)
			_, err = client.CsClient.VirtualMachine.StopVirtualMachine(params)
			if err != nil {
				return makeError(err)
			}
		}
	case commands.Reboot:
		if csVM.State == "Stopped" {
			params := client.CsClient.VirtualMachine.NewStartVirtualMachineParams(id)
			_, err = client.CsClient.VirtualMachine.StartVirtualMachine(params)
			if err != nil {
				return makeError(err)
			}
		} else if csVM.State != "Stopping" && csVM.State != "Starting" && csVM.State != "Migrating" {
			params := client.CsClient.VirtualMachine.NewRebootVirtualMachineParams(id)
			_, err = client.CsClient.VirtualMachine.RebootVirtualMachine(params)
			if err != nil {
				return makeError(err)
			}
		}
	}

	return nil
}

// HasCapacity checks if the VM has capacity in CloudStack.
func (client *Client) HasCapacity(vmID, hostName string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if vm %s can start on host %s. details: %w", vmID, hostName, err)
	}

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(vmID)
	if err != nil {
		return false, makeError(err)
	}

	host, _, err := client.CsClient.Host.GetHostByName(hostName)
	if err != nil {
		return false, makeError(err)
	}

	// if vm is running on the host we are comparing with, we can return true immediately
	if vm.State == "Running" && vm.Hostid == host.Id {
		return true, nil
	}

	vmMemoryByes := int64(vm.Memory) * 1024 * 1024

	if host.Memoryallocatedbytes+vmMemoryByes > host.Memorytotal {
		return false, nil
	}

	// this check SHOULD not be needed since over provisioning is allowed
	//if host.Cpunumber+vm.Cpunumber > host.Cputotal {
	//	return false, nil
	//}

	return true, nil
}

type cloudInit struct {
	FQDN            string          `yaml:"fqdn"`
	Users           []cloudInitUser `yaml:"users"`
	SshPasswordAuth bool            `yaml:"ssh_pwauth"`
	RunCMD          []string        `yaml:"runcmd"`
}

type cloudInitUser struct {
	Name              string   `yaml:"name"`
	Sudo              []string `yaml:"sudo"`
	Passwd            string   `yaml:"passwd"`
	LockPasswd        bool     `yaml:"lock_passwd"`
	Shell             string   `yaml:"shell"`
	SshAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
}

// createUserData creates the user data for the VM in CloudStack.
func createUserData(vmName string, userSshPublicKey, adminSshPublicKey *string) string {
	var authorizedKeys []string
	if userSshPublicKey != nil {
		authorizedKeys = append(authorizedKeys, *userSshPublicKey)
	}

	if adminSshPublicKey != nil {
		authorizedKeys = append(authorizedKeys, *adminSshPublicKey)
	}

	init := cloudInit{}
	init.FQDN = vmName
	init.SshPasswordAuth = false

	// imitate mkpasswd --method=SHA-512 --rounds=4096
	passwd := hashPassword("root", generateSalt())

	init.Users = append(init.Users, cloudInitUser{
		Name:              "root",
		Sudo:              []string{"ALL=(ALL) NOPASSWD:ALL"},
		Passwd:            passwd,
		LockPasswd:        true,
		Shell:             "/bin/bash",
		SshAuthorizedKeys: authorizedKeys,
	})

	// marshal the struct to yaml
	data, err := yaml.Marshal(init)
	if err != nil {
		panic(err)
	}

	return string(data)
}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// generateSalt generates the salt used for the password in CloudStack.
func generateSalt() string {
	//goland:noinspection GoDeprecation
	rand.Seed(time.Now().UnixNano())
	salt := make([]byte, 16)
	for i := range salt {
		salt[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(salt)
}

// hashPassword hashes the password in CloudStack.
func hashPassword(password, salt string) string {
	rounds := 4096
	hash := sha512.New()
	for i := 0; i < rounds; i++ {
		hash.Write([]byte(salt + password))
	}
	hashSum := hash.Sum(nil)
	encodedHash := base64.StdEncoding.EncodeToString(hashSum)
	return fmt.Sprintf("$6$rounds=%d$%s$%s", rounds, salt, encodedHash)
}
