package k8s

import (
	"fmt"
	"go-deploy/pkg/subsystems/k8s/models"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func int32Ptr(i int32) *int32 { return &i }

var manifestLabelName = "app.kubernetes.io/name"

func CreateNamespaceManifest(public *models.NamespacePublic) *apiv1.Namespace {
	return &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: public.Name,
			Labels: map[string]string{
				manifestLabelName: public.Name,
			},
		},
	}
}

func CreateDeploymentManifest(public *models.DeploymentPublic) *appsv1.Deployment {
	fullName := fmt.Sprintf("%s-%s", public.Name, public.ID)

	var envs []apiv1.EnvVar
	for _, env := range public.EnvVars {
		envs = append(envs, env.ToK8sEnvVar())
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fullName,
			Namespace: public.Namespace,
			Labels: map[string]string{
				"id":              public.ID,
				manifestLabelName: public.Name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					manifestLabelName: public.Name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deployId":        public.ID,
						manifestLabelName: public.Name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  public.Name,
							Image: public.DockerImage,
							Env:   envs,
						},
					},
				},
			},
		},
	}
}
func CreateServiceManifest(public *models.ServicePublic) *apiv1.Service {
	fullName := fmt.Sprintf("%s-%s", public.Name, public.ID)

	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fullName,
			Namespace: public.Namespace,
			Labels: map[string]string{
				"deployId":        public.ID,
				manifestLabelName: public.Name,
			},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "app-port",
					Protocol:   "TCP",
					Port:       int32(public.Port),
					TargetPort: intstr.FromInt(public.TargetPort),
				},
			},
			Selector: map[string]string{
				manifestLabelName: public.Name,
			},
		},
		Status: apiv1.ServiceStatus{},
	}
}
