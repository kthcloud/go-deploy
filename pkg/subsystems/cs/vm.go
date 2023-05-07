package cs

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"go-deploy/pkg/subsystems/cs/commands"
	"go-deploy/pkg/subsystems/cs/models"
	"gopkg.in/yaml.v3"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

type ManagedByAnnotation struct {
	ManagedBy string `json:"managedBy"`
}

func (client *Client) ReadVM(id string) (*models.VmPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cs vm %s. details: %s", id, err)
	}

	if id == "" {
		return nil, fmt.Errorf("id required")
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

func (client *Client) CreateVM(public *models.VmPublic, userSshPublicKey, adminSshPublicKey string) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cs vm %s. details: %s", public.Name, err)
	}

	listVmParams := client.CsClient.VirtualMachine.NewListVirtualMachinesParams()
	listVmParams.SetName(public.Name)
	listVmParams.SetProjectid(client.ProjectID)

	listVms, err := client.CsClient.VirtualMachine.ListVirtualMachines(listVmParams)
	if err != nil {
		return "", makeError(err)
	}

	var vmID string
	if listVms.Count == 0 {
		createVmParams := client.CsClient.VirtualMachine.NewDeployVirtualMachineParams(
			public.ServiceOfferingID,
			public.TemplateID,
			client.ZoneID,
		)

		createVmParams.SetName(public.Name)
		createVmParams.SetDisplayname(public.Name)
		createVmParams.SetNetworkids([]string{client.NetworkID})
		createVmParams.SetProjectid(client.ProjectID)
		createVmParams.SetExtraconfig(public.ExtraConfig)

		userData := createUserData(public.Name, userSshPublicKey, adminSshPublicKey)
		userData = "#cloud-config\n" + userData
		userDataB64 := base64.StdEncoding.EncodeToString([]byte(userData))

		createVmParams.SetUserdata(userDataB64)

		created, err := client.CsClient.VirtualMachine.DeployVirtualMachine(createVmParams)
		if err != nil {
			return "", makeError(err)
		}

		vmID = created.Id
	} else {
		vmID = listVms.VirtualMachines[0].Id
	}

	err = client.AssertVirtualMachineTags(vmID, public.Tags)
	if err != nil {
		return "", makeError(err)
	}

	return vmID, nil
}

func (client *Client) UpdateVM(public *models.VmPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update vm %s. details: %s", public.Name, err)
	}

	if public.ID == "" {
		return fmt.Errorf("id required")
	}

	params := client.CsClient.VirtualMachine.NewUpdateVirtualMachineParams(public.ID)
	params.SetName(public.Name)
	params.SetDisplayname(public.Name)

	if public.ExtraConfig == "" {
		params.SetExtraconfig(url.QueryEscape("none"))
	} else {
		params.SetExtraconfig(url.QueryEscape(public.ExtraConfig))
	}

	_, err := client.CsClient.VirtualMachine.UpdateVirtualMachine(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) DeleteVM(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete vm %s. details: %s", id, err)
	}

	if id == "" {
		return fmt.Errorf("id required")
	}

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return makeError(err)
		}
	}

	if vm == nil {
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

func (client *Client) GetVmStatus(id string) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get vm %s status. details: %s", id, err)
	}

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(id)
	if err != nil {
		return "", makeError(err)
	}

	return vm.State, nil
}

func (client *Client) DoVmCommand(id string, requiredHost *string, command commands.Command) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to execute csVM command to csVM %s. details: %s", id, err)
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
					return makeError(fmt.Errorf("failed to get host %s. details: %s", *requiredHost, err))
				}

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
		if csVM.State != "Stopping" && csVM.State != "Starting" && csVM.State != "Migrating" {
			params := client.CsClient.VirtualMachine.NewRebootVirtualMachineParams(id)
			_, err = client.CsClient.VirtualMachine.RebootVirtualMachine(params)
			if err != nil {
				return makeError(err)
			}
		}
	}

	return nil
}

type cloudInit struct {
	FQDN            string          `yaml:"fqdn"`
	Users           []cloudInitUser `yaml:"users"`
	SshPasswordAuth bool            `yaml:"ssh_pwauth"`
}

type cloudInitUser struct {
	Name              string   `yaml:"name"`
	Sudo              []string `yaml:"sudo"`
	Passwd            string   `yaml:"passwd"`
	LockPasswd        bool     `yaml:"lock_passwd"`
	Shell             string   `yaml:"shell"`
	SshAuthorizedKeys []string `yaml:"ssh_authorized_keys"`
}

func createUserData(vmName, userSshPublicKey, adminSshPublicKey string) string {
	init := cloudInit{}
	init.FQDN = vmName
	init.SshPasswordAuth = false

	// imitate mkpasswd --method=SHA-512 --rounds=4096
	passwd := hashPassword("root", generateSalt())

	init.Users = append(init.Users, cloudInitUser{
		Name:              "cloud",
		Sudo:              []string{"ALL=(ALL) NOPASSWD:ALL"},
		Passwd:            passwd,
		LockPasswd:        true,
		Shell:             "/bin/bash",
		SshAuthorizedKeys: []string{userSshPublicKey},
	})

	init.Users = append(init.Users, cloudInitUser{
		Name:              "deploy",
		Sudo:              []string{"ALL=(ALL) NOPASSWD:ALL"},
		Passwd:            passwd,
		LockPasswd:        true,
		Shell:             "/bin/bash",
		SshAuthorizedKeys: []string{adminSshPublicKey},
	})

	// marshal the struct to yaml
	data, err := yaml.Marshal(init)
	if err != nil {
		panic(err)
	}

	return string(data)
}

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateSalt() string {
	//goland:noinspection GoDeprecation
	rand.Seed(time.Now().UnixNano())
	salt := make([]byte, 16)
	for i := range salt {
		salt[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(salt)
}

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
