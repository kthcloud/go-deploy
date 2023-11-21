package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	appsv1 "k8s.io/api/apps/v1"
	"time"
)

type DeploymentPublic struct {
	ID               string          `bson:"id"`
	Name             string          `bson:"name"`
	Namespace        string          `bson:"namespace"`
	Image            string          `bson:"image"`
	ImagePullSecrets []string        `bson:"imagePullSecrets"`
	EnvVars          []EnvVar        `bson:"envVars"`
	Resources        Resources       `bson:"resources"`
	Command          []string        `bson:"command"`
	Args             []string        `bson:"args"`
	InitCommands     []string        `bson:"initCommands"`
	InitContainers   []InitContainer `bson:"initContainers"`
	Volumes          []Volume        `bson:"volumes"`
	CreatedAt        time.Time       `bson:"createdAt"`
}

func (d *DeploymentPublic) GetID() string {
	return d.ID
}

func (d *DeploymentPublic) Created() bool {
	return d.ID != ""
}

func (d *DeploymentPublic) IsPlaceholder() bool {
	return false
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
	var image string

	for _, k8sVolume := range deployment.Spec.Template.Spec.Volumes {
		var pvcName *string
		if k8sVolume.PersistentVolumeClaim != nil {
			pvcName = &k8sVolume.PersistentVolumeClaim.ClaimName
		}

		volumes = append(volumes, Volume{
			Name:    k8sVolume.Name,
			PvcName: pvcName,
			Init:    false,
		})
	}

	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		firstContainer := deployment.Spec.Template.Spec.Containers[0]
		resources := firstContainer.Resources
		lifecycle := firstContainer.Lifecycle
		volumeMounts := firstContainer.VolumeMounts
		image = firstContainer.Image
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
			// if we cannot find the volume mount in the volume list, then it is not a volume we care about
			for idx, volume := range volumes {
				if volume.Name == volumeMount.Name {
					volumes[idx].MountPath = volumeMount.MountPath
					break
				}
			}
		}
	}

	imagePullSecrets := make([]string, len(deployment.Spec.Template.Spec.ImagePullSecrets))
	for idx, imagePullSecret := range deployment.Spec.Template.Spec.ImagePullSecrets {
		imagePullSecrets[idx] = imagePullSecret.Name
	}

	initContainers := make([]InitContainer, len(deployment.Spec.Template.Spec.InitContainers))
	for idx, initContainer := range deployment.Spec.Template.Spec.InitContainers {
		initContainers[idx] = InitContainer{
			Name:    initContainer.Name,
			Image:   initContainer.Image,
			Command: initContainer.Command,
			Args:    initContainer.Args,
		}

		for _, volumeMount := range initContainer.VolumeMounts {
			// if we cannot find the volume mount in the volume list, then it is not a volume we care about
			for _, volume := range volumes {
				if volume.Name == volumeMount.Name {
					// if the volume is shared between standard and init containers, then we need to create a new volume
					// for the init container, otherwise just set the mount path
					if volume.MountPath == "" {
						volume.MountPath = volumeMount.MountPath
					} else {
						volumes = append(volumes, Volume{
							Name:      volume.Name,
							PvcName:   volume.PvcName,
							MountPath: volumeMount.MountPath,
							Init:      true,
						})
					}
					break
				}
			}
		}
	}

	// delete any k8sVolumes that does not have a mount path, they need to be recreated
	for i := len(volumes) - 1; i >= 0; i-- {
		if volumes[i].MountPath == "" {
			volumes = append(volumes[:i], volumes[i+1:]...)
		}
	}

	return &DeploymentPublic{
		ID:               deployment.Labels[keys.ManifestLabelID],
		Name:             deployment.Name,
		Namespace:        deployment.Namespace,
		Image:            image,
		ImagePullSecrets: imagePullSecrets,
		EnvVars:          envs,
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
