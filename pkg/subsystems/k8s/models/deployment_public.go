package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	appsv1 "k8s.io/api/apps/v1"
	"time"
)

type DeploymentPublic struct {
	ID             string          `bson:"id"`
	Name           string          `bson:"name"`
	Namespace      string          `bson:"namespace"`
	DockerImage    string          `bson:"dockerImage"`
	EnvVars        []EnvVar        `bson:"envVars"`
	Resources      Resources       `bson:"resources"`
	Command        []string        `bson:"command"`
	Args           []string        `bson:"args"`
	InitCommands   []string        `bson:"initCommands"`
	InitContainers []InitContainer `bson:"initContainers"`
	Volumes        []Volume        `bson:"volumes"`
	CreatedAt      time.Time       `bson:"createdAt"`
}

func (d *DeploymentPublic) Created() bool {
	return d.ID != ""
}

func CreateDeploymentPublicFromRead(deployment *appsv1.Deployment) *DeploymentPublic {
	var envs []EnvVar
	for _, env := range deployment.Spec.Template.Spec.Containers[0].Env {
		envs = append(envs, EnvVarFromK8s(&env))
	}

	var limits = Limits{}
	var requests = Requests{}
	var initCommands []string
	var command []string
	var args []string
	var volumes []Volume

	for _, k8sVolume := range deployment.Spec.Template.Spec.Volumes {
		var pvcName *string
		if k8sVolume.PersistentVolumeClaim != nil {
			pvcName = &k8sVolume.PersistentVolumeClaim.ClaimName
		}

		volumes = append(volumes, Volume{
			Name:    k8sVolume.Name,
			PvcName: pvcName,
		})
	}

	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		firstContainer := deployment.Spec.Template.Spec.Containers[0]
		resources := firstContainer.Resources
		lifecycle := firstContainer.Lifecycle
		volumeMounts := firstContainer.VolumeMounts

		command = firstContainer.Command
		args = firstContainer.Args

		if resources.Limits != nil {
			if resources.Limits.Cpu() != nil {
				limits.CPU = resources.Limits.Cpu().String()
			}
			if resources.Limits.Memory() != nil {
				limits.Memory = resources.Limits.Memory().String()
			}
		}

		if resources.Requests != nil {
			if resources.Requests.Cpu() != nil {
				requests.CPU = resources.Requests.Cpu().String()
			}
			if resources.Requests.Memory() != nil {
				requests.Memory = resources.Requests.Memory().String()
			}
		}

		if lifecycle != nil && lifecycle.PostStart != nil && lifecycle.PostStart.Exec != nil {
			initCommands = append(initCommands, lifecycle.PostStart.Exec.Command...)
		}

		for _, volumeMount := range volumeMounts {
			// if we cannot find the volume mount in the volumes list, then it is not a volume we care about
			for _, volume := range volumes {
				if volume.Name == volumeMount.Name {
					volume.MountPath = volumeMount.MountPath
				}
			}
		}
	}

	initContainers := make([]InitContainer, len(deployment.Spec.Template.Spec.InitContainers))
	for _, initContainer := range deployment.Spec.Template.Spec.InitContainers {
		initContainers = append(initContainers, InitContainer{
			Name:    initContainer.Name,
			Image:   initContainer.Image,
			Command: initContainer.Command,
			Args:    initContainer.Args,
		})
	}

	// delete any k8sVolumes that does not have a mount path, they need to be recreated
	for i := len(volumes) - 1; i >= 0; i-- {
		if volumes[i].MountPath == "" {
			volumes = append(volumes[:i], volumes[i+1:]...)
		}
	}

	return &DeploymentPublic{
		ID:          deployment.Labels[keys.ManifestLabelID],
		Name:        deployment.Labels[keys.ManifestLabelName],
		Namespace:   deployment.Namespace,
		DockerImage: deployment.Spec.Template.Spec.Containers[0].Image,
		EnvVars:     envs,
		Resources: Resources{
			Limits:   limits,
			Requests: requests,
		},
		Command:        command,
		Args:           args,
		InitCommands:   initCommands,
		InitContainers: initContainers,
		Volumes:        volumes,
		CreatedAt:      formatCreatedAt(deployment.Annotations),
	}
}
