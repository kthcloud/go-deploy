package k8s

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/subsystems/k8s/models"
	"go-deploy/test"
	"go-deploy/test/acc"
	v1 "k8s.io/api/core/v1"
	"testing"
)

func withContext(t *testing.T) (*k8s.Client, *models.NamespacePublic) {
	n := withNamespace(t)
	return withClient(t, n.Name), n
}

func withClient(t *testing.T, namespace string) *k8s.Client {
	zoneName := "se-flem"
	zone := config.Config.Deployment.GetZone(zoneName)
	if zone == nil {
		t.Fatalf("no zone with name %s found", zoneName)
	}

	client, err := k8s.New(zone.Client, namespace)
	if err != nil {
		t.Fatalf(err.Error())
	}

	return client
}

func withNamespace(t *testing.T) *models.NamespacePublic {
	c := withClient(t, "")

	n := &models.NamespacePublic{
		Name: acc.GenName(),
	}

	nCreated, err := c.CreateNamespace(n)
	test.NoError(t, err, "failed to create namespace")
	assert.NotZero(t, nCreated.ID, "no namespace id received from client")
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
	assert.NotZero(t, dCreated.ID, "no deployment id received from client")
	t.Cleanup(func() { cleanUpDeployment(t, c, dCreated.ID) })

	assert.Equal(t, d.Name, dCreated.Name, "deployment name does not match")
	assert.Equal(t, d.Namespace, dCreated.Namespace, "deployment namespace does not match")
	assert.Equal(t, d.Image, dCreated.Image, "deployment image does not match")
	test.EqualOrEmpty(t, d.EnvVars, dCreated.EnvVars, "deployment env vars do not match")

	return dCreated
}

func withDefaultService(t *testing.T, c *k8s.Client) *models.ServicePublic {
	s := &models.ServicePublic{
		Name:       acc.GenName(),
		Namespace:  c.Namespace,
		Port:       11111,
		TargetPort: 22222,
	}

	return withService(t, c, s)
}

func withService(t *testing.T, c *k8s.Client, s *models.ServicePublic) *models.ServicePublic {
	sCreated, err := c.CreateService(s)
	test.NoError(t, err, "failed to create service")
	assert.NotZero(t, sCreated.ID, "no service id received from client")
	t.Cleanup(func() { cleanUpService(t, c, sCreated.ID) })

	assert.Equal(t, s.Name, sCreated.Name, "service name does not match")
	assert.Equal(t, s.Namespace, sCreated.Namespace, "service namespace does not match")
	assert.Equal(t, s.Port, sCreated.Port, "service port does not match")
	assert.Equal(t, s.TargetPort, sCreated.TargetPort, "service target port does not match")

	return sCreated
}

func withDefaultIngress(t *testing.T, c *k8s.Client, s *models.ServicePublic) *models.IngressPublic {
	i := &models.IngressPublic{
		Name:         acc.GenName(),
		Namespace:    c.Namespace,
		ServiceName:  s.Name,
		ServicePort:  s.Port,
		IngressClass: "nginx",
		Hosts:        []string{acc.GenName() + ".example.com"},
	}

	return withIngress(t, c, i)
}

func withIngress(t *testing.T, c *k8s.Client, i *models.IngressPublic) *models.IngressPublic {
	iCreated, err := c.CreateIngress(i)
	test.NoError(t, err, "failed to create ingress")
	assert.NotZero(t, iCreated.ID, "no ingress id received from client")
	t.Cleanup(func() { cleanUpIngress(t, c, iCreated.ID) })

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
	assert.NotZero(t, pvcCreated.ID, "no pvc id received from client")
	t.Cleanup(func() { cleanUpPVC(t, c, pvcCreated.ID) })

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
	assert.NotZero(t, pvCreated.ID, "no pv id received from client")
	t.Cleanup(func() { cleanUpPV(t, c, pvCreated.ID) })

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
	assert.NotZero(t, hpaCreated.ID, "no hpa id received from client")
	t.Cleanup(func() { cleanUpHPA(t, c, hpaCreated.ID) })

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
	assert.NotZero(t, secretCreated.ID, "no secret id received from client")
	t.Cleanup(func() { cleanUpSecret(t, c, secretCreated.ID) })

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
	assert.NotZero(t, jobCreated.ID, "no job id received from client")
	t.Cleanup(func() { cleanUpJob(t, c, jobCreated.ID) })

	assert.Equal(t, job.Name, jobCreated.Name, "job name does not match")
	assert.Equal(t, job.Namespace, jobCreated.Namespace, "job namespace does not match")
	assert.Equal(t, job.Image, jobCreated.Image, "job image does not match")
	assert.Equal(t, job.Command, jobCreated.Command, "job command does not match")
	assert.Equal(t, job.Args, jobCreated.Args, "job args does not match")

	return jobCreated
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

func cleanUpDeployment(t *testing.T, c *k8s.Client, id string) {
	err := c.DeleteDeployment(id)
	test.NoError(t, err, "failed to delete deployment")

	deletedDeployment, err := c.ReadDeployment(id)
	test.NoError(t, err, "failed to read deployment")
	assert.Nil(t, deletedDeployment, "deployment still exists")

	err = c.DeleteDeployment(id)
	test.NoError(t, err, "failed to delete deployment again")
}

func cleanUpService(t *testing.T, c *k8s.Client, id string) {
	err := c.DeleteService(id)
	test.NoError(t, err, "failed to delete service")

	deletedService, err := c.ReadService(id)
	test.NoError(t, err, "failed to read service")
	assert.Nil(t, deletedService, "service still exists")

	err = c.DeleteService(id)
	test.NoError(t, err, "failed to delete service again")
}

func cleanUpIngress(t *testing.T, c *k8s.Client, id string) {
	err := c.DeleteIngress(id)
	test.NoError(t, err, "failed to delete ingress")

	deletedIngress, err := c.ReadIngress(id)
	test.NoError(t, err, "failed to read ingress")
	assert.Nil(t, deletedIngress, "ingress still exists")

	err = c.DeleteIngress(id)
	test.NoError(t, err, "failed to delete ingress again")
}

func cleanUpPVC(t *testing.T, c *k8s.Client, id string) {
	err := c.DeletePVC(id)
	test.NoError(t, err, "failed to delete pvc")

	deletedPVC, err := c.ReadPVC(id)
	test.NoError(t, err, "failed to read pvc")
	assert.Nil(t, deletedPVC, "pvc still exists")

	err = c.DeletePVC(id)
	test.NoError(t, err, "failed to delete pvc again")
}

func cleanUpPV(t *testing.T, c *k8s.Client, id string) {
	err := c.DeletePV(id)
	test.NoError(t, err, "failed to delete pv")

	deletedPV, err := c.ReadPV(id)
	test.NoError(t, err, "failed to read pv")
	assert.Nil(t, deletedPV, "pv still exists")

	err = c.DeletePV(id)
	test.NoError(t, err, "failed to delete pv again")
}

func cleanUpHPA(t *testing.T, c *k8s.Client, id string) {
	err := c.DeleteHPA(id)
	test.NoError(t, err, "failed to delete hpa")

	deletedHPA, err := c.ReadHPA(id)
	test.NoError(t, err, "failed to read hpa")
	assert.Nil(t, deletedHPA, "hpa still exists")

	err = c.DeleteHPA(id)
	test.NoError(t, err, "failed to delete hpa again")
}

func cleanUpSecret(t *testing.T, c *k8s.Client, id string) {
	err := c.DeleteSecret(id)
	test.NoError(t, err, "failed to delete secret")

	deletedSecret, err := c.ReadSecret(id)
	test.NoError(t, err, "failed to read secret")
	assert.Nil(t, deletedSecret, "secret still exists")

	err = c.DeleteSecret(id)
	test.NoError(t, err, "failed to delete secret again")
}

func cleanUpJob(t *testing.T, c *k8s.Client, id string) {
	err := c.DeleteJob(id)
	test.NoError(t, err, "failed to delete job")

	deletedJob, err := c.ReadJob(id)
	test.NoError(t, err, "failed to read job")
	assert.Nil(t, deletedJob, "job still exists")

	err = c.DeleteJob(id)
	test.NoError(t, err, "failed to delete job again")
}
