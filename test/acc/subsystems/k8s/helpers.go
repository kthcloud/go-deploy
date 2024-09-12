package k8s

import (
	"github.com/stretchr/testify/assert"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	"github.com/kthcloud/go-deploy/test"
	"github.com/kthcloud/go-deploy/test/acc"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func withContext(t *testing.T, zoneName ...string) (*k8s.Client, *models.NamespacePublic) {
	n := withNamespace(t, zoneName...)
	return withClient(t, n.Name, zoneName...), n
}

func withClient(t *testing.T, namespace string, zoneName ...string) *k8s.Client {
	zName := config.Config.VM.DefaultZone
	if len(zoneName) > 0 {
		zName = zoneName[0]
	}

	zone := config.Config.GetZone(zName)
	if zone == nil {
		t.Fatalf("no zone with name %s found", zoneName)
	}

	client, err := k8s.New(&k8s.ClientConf{
		K8sClient:         zone.K8s.Client,
		KubeVirtK8sClient: zone.K8s.KubeVirtClient,
		Namespace:         namespace,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}

	return client
}

func withNamespace(t *testing.T, zoneName ...string) *models.NamespacePublic {
	c := withClient(t, "", zoneName...)

	n := &models.NamespacePublic{
		Name: acc.GenName(),
	}

	nCreated, err := c.CreateNamespace(n)
	test.NoError(t, err, "failed to create namespace")
	assert.True(t, nCreated.Created(), "namespace was not created")
	t.Cleanup(func() { cleanUpNamespace(t, c, nCreated.Name) })

	assert.Equal(t, n.Name, nCreated.Name, "namespace name does not match")

	return n
}

func withDefaultDeployment(t *testing.T, c *k8s.Client) *models.DeploymentPublic {
	d := &models.DeploymentPublic{
		Name:      acc.GenName(),
		Namespace: c.Namespace,
		Image:     config.Config.Registry.PlaceholderImage,
		EnvVars:   []models.EnvVar{{Name: acc.GenName(), Value: acc.GenName()}},
	}

	return withDeployment(t, c, d)
}

func withDeployment(t *testing.T, c *k8s.Client, d *models.DeploymentPublic) *models.DeploymentPublic {
	dCreated, err := c.CreateDeployment(d)
	test.NoError(t, err, "failed to create deployment")
	assert.True(t, dCreated.Created(), "deployment was not created")
	t.Cleanup(func() { cleanUpDeployment(t, c, dCreated.Name) })

	assert.Equal(t, d.Name, dCreated.Name, "deployment name does not match")
	assert.Equal(t, d.Namespace, dCreated.Namespace, "deployment namespace does not match")
	assert.Equal(t, d.Image, dCreated.Image, "deployment image does not match")
	test.EqualOrEmpty(t, d.EnvVars, dCreated.EnvVars, "deployment env vars do not match")

	return dCreated
}

func withDefaultVM(t *testing.T, c *k8s.Client) *models.VmPublic {
	name := acc.GenName()
	sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDbFXrLEF2PYNodfwNfGe+4qM3FeZ/FxcfYLZwxStKVW/eTgYn3Y0DQSti86mA+Jrzkx2aSvHDPPJEQUUTiZUMwTiJlR4ud3FBDYZXQVsNhfJO5zduqLpEEdjtFMP/Y3jGpoh+Eq8U08yWRfs1sKay/THS5MoKIprFVU+yIgHsxNcrU2hymTnt+A43dxKHXd4aZXhfW5qHwBZzNBggRXPFb6RpABx2qk9dQGGHFrGp5p0cC2sekXNFg7lV8PEx8pspu+Kf5mSBd1aphRde8ATR61zEDbAGKi1wbGHhrrZ/dAigcSB5YNgllTg5l09CwtjWBFHGY1oxwb+F3foXH19dDIlkB7wsyoT/XD7NMOfNyqDYLlOrVVMPfEdNkBXdCScK8Q3rrT/LL/7fefo/OirDnCvL3dxEA/9ay0hVEHyef6E++tiO9DU/HBVAY6NYjYQCZZ92rqVPzM94ppBU4XocxzAQ7zL+pFABbZkYtXTH8VaE4A1MTgRXvte1CmzeFPQs= emil@thinkpad"

	cloudInitString := `#cloud-config
fqdn: ` + name + `
users:
    - name: root
      sudo:
        - ALL=(ALL) NOPASSWD:ALL
      passwd: $6$rounds=4096$e1qOwC3bflSjcbvK$/Otkkdsr+tcKV0pIFSNe0esOBw2lDUBlUWJMNiKFhQaQJW6wa2Cz4OQmmr3woItXRQV7WG2ooGMX9S0Ikc2Fmw==
      lock_passwd: false
      shell: /bin/bash
      ssh_authorized_keys:
        - ` + sshPublicKey + `
ssh_pwauth: false
runcmd:
    - git clone https://github.com/kthcloud/boostrap-vm.git init && cd init && chmod +x run.sh && ./run.sh
`

	vm := &models.VmPublic{
		Name:      name,
		Namespace: c.Namespace,
		// Temporary image
		Image:     "docker://registry.cloud.cbh.kth.se/images/ubuntu:24.04",
		CpuCores:  1,
		RAM:       4,
		DiskSize:  15,
		CloudInit: cloudInitString,
	}

	return withVM(t, c, vm)
}

func withVM(t *testing.T, c *k8s.Client, vm *models.VmPublic) *models.VmPublic {
	vmCreated, err := c.CreateVM(vm)
	test.NoError(t, err, "failed to create vm")
	assert.True(t, vmCreated.Created(), "vm was not created")
	t.Cleanup(func() { cleanUpVM(t, c, vmCreated.Name) })

	assert.Equal(t, vm.Name, vmCreated.Name, "vm name does not match")
	assert.Equal(t, vm.Namespace, vmCreated.Namespace, "vm namespace does not match")
	assert.Equal(t, vm.Image, vmCreated.Image, "vm image does not match")
	assert.Equal(t, vm.CpuCores, vmCreated.CpuCores, "vm cpu cores do not match")
	assert.Equal(t, vm.RAM, vmCreated.RAM, "vm ram does not match")
	assert.Equal(t, vm.DiskSize, vmCreated.DiskSize, "vm disk size does not match")
	test.EqualOrEmpty(t, vm.GPUs, vmCreated.GPUs, "vm gpus do not match")

	return vmCreated
}

func withDefaultService(t *testing.T, c *k8s.Client) *models.ServicePublic {
	s := &models.ServicePublic{
		Name:      acc.GenName(),
		Namespace: c.Namespace,
		Ports:     []models.Port{{Port: 11111, TargetPort: 22222}},
	}

	return withService(t, c, s)
}

func withService(t *testing.T, c *k8s.Client, s *models.ServicePublic) *models.ServicePublic {
	sCreated, err := c.CreateService(s)
	test.NoError(t, err, "failed to create service")
	assert.True(t, sCreated.Created(), "service was not created")
	t.Cleanup(func() { cleanUpService(t, c, sCreated.Name) })

	assert.Equal(t, s.Name, sCreated.Name, "service name does not match")
	assert.Equal(t, s.Namespace, sCreated.Namespace, "service namespace does not match")
	assert.Equal(t, s.Ports[0].Port, sCreated.Ports[0].Port, "service port does not match")
	assert.Equal(t, s.Ports[0].TargetPort, sCreated.Ports[0].TargetPort, "service target port does not match")

	return sCreated
}

func withDefaultIngress(t *testing.T, c *k8s.Client, s *models.ServicePublic) *models.IngressPublic {
	i := &models.IngressPublic{
		Name:         acc.GenName(),
		Namespace:    c.Namespace,
		ServiceName:  s.Name,
		ServicePort:  s.Ports[0].Port,
		IngressClass: "nginx",
		Hosts:        []string{acc.GenName() + ".example.com"},
	}

	return withIngress(t, c, i)
}

func withIngress(t *testing.T, c *k8s.Client, i *models.IngressPublic) *models.IngressPublic {
	iCreated, err := c.CreateIngress(i)
	test.NoError(t, err, "failed to create ingress")
	assert.True(t, iCreated.Created(), "ingress was not created")
	t.Cleanup(func() { cleanUpIngress(t, c, iCreated.Name) })

	assert.Equal(t, i.Name, iCreated.Name, "ingress name does not match")
	assert.Equal(t, i.Namespace, iCreated.Namespace, "ingress namespace does not match")
	assert.Equal(t, i.ServiceName, iCreated.ServiceName, "ingress service name does not match")
	assert.Equal(t, i.ServicePort, iCreated.ServicePort, "ingress service port does not match")
	assert.Equal(t, i.IngressClass, iCreated.IngressClass, "ingress ingress class does not match")
	test.EqualOrEmpty(t, i.Hosts, iCreated.Hosts, "ingress hosts do not match")

	return iCreated
}

func withDefaultPVC(t *testing.T, c *k8s.Client, pv *models.PvPublic) *models.PvcPublic {
	pvc := &models.PvcPublic{
		Name:      acc.GenName(),
		Namespace: c.Namespace,
		Capacity:  "1Gi",
		PvName:    pv.Name,
	}

	return withPVC(t, c, pvc)
}

func withPVC(t *testing.T, c *k8s.Client, pvc *models.PvcPublic) *models.PvcPublic {
	pvcCreated, err := c.CreatePVC(pvc)
	test.NoError(t, err, "failed to create pvc")
	assert.True(t, pvcCreated.Created(), "pvc was not created")
	t.Cleanup(func() { cleanUpPVC(t, c, pvcCreated.Name) })

	assert.Equal(t, pvc.Name, pvcCreated.Name, "pvc name does not match")
	assert.Equal(t, pvc.Namespace, pvcCreated.Namespace, "pvc namespace does not match")
	assert.Equal(t, pvc.PvName, pvcCreated.PvName, "pvc pv name does not match")

	return pvcCreated
}

func withDefaultPV(t *testing.T, c *k8s.Client) *models.PvPublic {
	pv := &models.PvPublic{
		Name:      acc.GenName(),
		Capacity:  "1Gi",
		NfsServer: "some.nfs.server",
		NfsPath:   "/some/nfs/path",
	}

	return withPV(t, c, pv)
}

func withPV(t *testing.T, c *k8s.Client, pv *models.PvPublic) *models.PvPublic {
	pvCreated, err := c.CreatePV(pv)
	test.NoError(t, err, "failed to create pv")
	assert.True(t, pvCreated.Created(), "pv was not created")
	t.Cleanup(func() { cleanUpPV(t, c, pvCreated.Name) })

	assert.Equal(t, pv.Name, pvCreated.Name, "pv name does not match")
	assert.Equal(t, pv.Capacity, pvCreated.Capacity, "pv capacity does not match")
	assert.Equal(t, pv.NfsServer, pvCreated.NfsServer, "pv nfs server does not match")
	assert.Equal(t, pv.NfsPath, pvCreated.NfsPath, "pv nfs path does not match")

	return pvCreated
}

func withDefaultHPA(t *testing.T, c *k8s.Client, d *models.DeploymentPublic) *models.HpaPublic {
	hpa := &models.HpaPublic{
		Name:        acc.GenName(),
		Namespace:   c.Namespace,
		MinReplicas: 1,
		MaxReplicas: 2,
		Target: models.Target{
			Kind:       "Deployment",
			Name:       d.Name,
			ApiVersion: "apps/v1",
		},
		CpuAverageUtilization:    50,
		MemoryAverageUtilization: 50,
	}

	return withHPA(t, c, hpa)
}

func withHPA(t *testing.T, c *k8s.Client, hpa *models.HpaPublic) *models.HpaPublic {
	hpaCreated, err := c.CreateHPA(hpa)
	test.NoError(t, err, "failed to create hpa")
	assert.True(t, hpaCreated.Created(), "hpa was not created")
	t.Cleanup(func() { cleanUpHPA(t, c, hpaCreated.Name) })

	assert.Equal(t, hpa.Name, hpaCreated.Name, "hpa name does not match")
	assert.Equal(t, hpa.Namespace, hpaCreated.Namespace, "hpa namespace does not match")
	assert.Equal(t, hpa.MinReplicas, hpaCreated.MinReplicas, "hpa min replicas does not match")
	assert.Equal(t, hpa.MaxReplicas, hpaCreated.MaxReplicas, "hpa max replicas does not match")
	assert.Equal(t, hpa.Target, hpaCreated.Target, "hpa target does not match")
	assert.Equal(t, hpa.CpuAverageUtilization, hpaCreated.CpuAverageUtilization, "hpa cpu average utilization does not match")
	assert.Equal(t, hpa.MemoryAverageUtilization, hpaCreated.MemoryAverageUtilization, "hpa memory average utilization does not match")

	return hpaCreated
}

func withDefaultSecret(t *testing.T, c *k8s.Client) *models.SecretPublic {
	secret := &models.SecretPublic{
		Name:      acc.GenName(),
		Namespace: c.Namespace,
		Data:      map[string][]byte{"key": []byte("value")},
		Type:      string(v1.SecretTypeOpaque),
	}

	return withSecret(t, c, secret)
}

func withSecret(t *testing.T, c *k8s.Client, secret *models.SecretPublic) *models.SecretPublic {
	secretCreated, err := c.CreateSecret(secret)
	test.NoError(t, err, "failed to create secret")
	assert.True(t, secretCreated.Created(), "secret was not created")
	t.Cleanup(func() { cleanUpSecret(t, c, secretCreated.Name) })

	assert.Equal(t, secret.Name, secretCreated.Name, "secret name does not match")
	assert.Equal(t, secret.Namespace, secretCreated.Namespace, "secret namespace does not match")
	assert.Equal(t, secret.Data, secretCreated.Data, "secret data does not match")
	assert.Equal(t, secret.Type, secretCreated.Type, "secret type does not match")
	assert.Equal(t, secret.Placeholder, secretCreated.Placeholder, "secret placeholder does not match")

	return secretCreated
}

func withDefaultJob(t *testing.T, c *k8s.Client) *models.JobPublic {
	job := &models.JobPublic{
		Name:      acc.GenName(),
		Namespace: c.Namespace,
		Image:     config.Config.Registry.PlaceholderImage,
		Command:   []string{"echo", "hello world"},
		Args:      []string{"hello", "world"},
	}

	return withJob(t, c, job)
}

func withJob(t *testing.T, c *k8s.Client, job *models.JobPublic) *models.JobPublic {
	jobCreated, err := c.CreateJob(job)
	test.NoError(t, err, "failed to create job")
	assert.True(t, jobCreated.Created(), "job was not created")
	t.Cleanup(func() { cleanUpJob(t, c, jobCreated.Name) })

	assert.Equal(t, job.Name, jobCreated.Name, "job name does not match")
	assert.Equal(t, job.Namespace, jobCreated.Namespace, "job namespace does not match")
	assert.Equal(t, job.Image, jobCreated.Image, "job image does not match")
	assert.Equal(t, job.Command, jobCreated.Command, "job command does not match")
	assert.Equal(t, job.Args, jobCreated.Args, "job args does not match")

	return jobCreated
}

func withDefaultNetworkPolicy(t *testing.T, c *k8s.Client) *models.NetworkPolicyPublic {
	np := &models.NetworkPolicyPublic{
		Name:      acc.GenName(),
		Namespace: c.Namespace,
		EgressRules: []models.EgressRule{
			{
				IpBlock: &models.IpBlock{
					CIDR:   "0.0.0.0/0",
					Except: []string{"1.1.1.1/32"},
				},
			},
			{
				PodSelector:       map[string]string{"name": "block-me"},
				NamespaceSelector: map[string]string{"name": "block-me"},
			},
		},
	}

	return withNetworkPolicy(t, c, np)
}

func withNetworkPolicy(t *testing.T, c *k8s.Client, np *models.NetworkPolicyPublic) *models.NetworkPolicyPublic {
	npCreated, err := c.CreateNetworkPolicy(np)
	test.NoError(t, err, "failed to create network policy")
	assert.True(t, npCreated.Created(), "network policy was not created")
	t.Cleanup(func() { cleanUpNetworkPolicy(t, c, npCreated.Name) })

	assert.Equal(t, np.Name, npCreated.Name, "network policy name does not match")
	assert.Equal(t, np.Namespace, npCreated.Namespace, "network policy namespace does not match")
	assert.Equal(t, np.EgressRules, npCreated.EgressRules, "network policy egress rules do not match")

	return npCreated
}

func cleanUpNamespace(t *testing.T, c *k8s.Client, name string) {
	err := c.DeleteNamespace(name)
	test.NoError(t, err, "failed to delete namespace")

	deletedNamespace, err := c.ReadNamespace(name)
	test.NoError(t, err, "failed to read namespace")
	assert.Nil(t, deletedNamespace, "namespace still exists")

	err = c.DeleteNamespace(name)
	test.NoError(t, err, "failed to delete namespace again")
}

func cleanUpDeployment(t *testing.T, c *k8s.Client, name string) {
	err := c.DeleteDeployment(name)
	test.NoError(t, err, "failed to delete deployment")

	deletedDeployment, err := c.ReadDeployment(name)
	test.NoError(t, err, "failed to read deployment")
	assert.Nil(t, deletedDeployment, "deployment still exists")

	err = c.DeleteDeployment(name)
	test.NoError(t, err, "failed to delete deployment again")
}

func cleanUpVM(t *testing.T, c *k8s.Client, name string) {
	err := c.DeleteVM(name)
	test.NoError(t, err, "failed to delete vm")

	deletedVM, err := c.ReadVM(name)
	test.NoError(t, err, "failed to read vm")
	assert.Nil(t, deletedVM, "vm still exists")

	err = c.DeleteVM(name)
	test.NoError(t, err, "failed to delete vm again")
}

func cleanUpService(t *testing.T, c *k8s.Client, name string) {
	err := c.DeleteService(name)
	test.NoError(t, err, "failed to delete service")

	deletedService, err := c.ReadService(name)
	test.NoError(t, err, "failed to read service")
	assert.Nil(t, deletedService, "service still exists")

	err = c.DeleteService(name)
	test.NoError(t, err, "failed to delete service again")
}

func cleanUpIngress(t *testing.T, c *k8s.Client, name string) {
	err := c.DeleteIngress(name)
	test.NoError(t, err, "failed to delete ingress")

	deletedIngress, err := c.ReadIngress(name)
	test.NoError(t, err, "failed to read ingress")
	assert.Nil(t, deletedIngress, "ingress still exists")

	err = c.DeleteIngress(name)
	test.NoError(t, err, "failed to delete ingress again")
}

func cleanUpPVC(t *testing.T, c *k8s.Client, name string) {
	err := c.DeletePVC(name)
	test.NoError(t, err, "failed to delete pvc")

	deletedPVC, err := c.ReadPVC(name)
	test.NoError(t, err, "failed to read pvc")
	assert.Nil(t, deletedPVC, "pvc still exists")

	err = c.DeletePVC(name)
	test.NoError(t, err, "failed to delete pvc again")
}

func cleanUpPV(t *testing.T, c *k8s.Client, name string) {
	err := c.DeletePV(name)
	test.NoError(t, err, "failed to delete pv")

	deletedPV, err := c.ReadPV(name)
	test.NoError(t, err, "failed to read pv")
	assert.Nil(t, deletedPV, "pv still exists")

	err = c.DeletePV(name)
	test.NoError(t, err, "failed to delete pv again")
}

func cleanUpHPA(t *testing.T, c *k8s.Client, name string) {
	err := c.DeleteHPA(name)
	test.NoError(t, err, "failed to delete hpa")

	deletedHPA, err := c.ReadHPA(name)
	test.NoError(t, err, "failed to read hpa")
	assert.Nil(t, deletedHPA, "hpa still exists")

	err = c.DeleteHPA(name)
	test.NoError(t, err, "failed to delete hpa again")
}

func cleanUpSecret(t *testing.T, c *k8s.Client, name string) {
	err := c.DeleteSecret(name)
	test.NoError(t, err, "failed to delete secret")

	deletedSecret, err := c.ReadSecret(name)
	test.NoError(t, err, "failed to read secret")
	assert.Nil(t, deletedSecret, "secret still exists")

	err = c.DeleteSecret(name)
	test.NoError(t, err, "failed to delete secret again")
}

func cleanUpJob(t *testing.T, c *k8s.Client, name string) {
	err := c.DeleteJob(name)
	test.NoError(t, err, "failed to delete job")

	deletedJob, err := c.ReadJob(name)
	test.NoError(t, err, "failed to read job")
	assert.Nil(t, deletedJob, "job still exists")

	err = c.DeleteJob(name)
	test.NoError(t, err, "failed to delete job again")
}

func cleanUpNetworkPolicy(t *testing.T, c *k8s.Client, name string) {
	err := c.DeleteNetworkPolicy(name)
	test.NoError(t, err, "failed to delete network policy")

	deletedNetworkPolicy, err := c.ReadNetworkPolicy(name)
	test.NoError(t, err, "failed to read network policy")
	assert.Nil(t, deletedNetworkPolicy, "network policy still exists")

	err = c.DeleteNetworkPolicy(name)
	test.NoError(t, err, "failed to delete network policy again")
}
