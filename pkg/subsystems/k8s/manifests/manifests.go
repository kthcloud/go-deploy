package manifests

import (
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func int32Ptr(i int32) *int32 { return &i }

var manifestLabelName = "app.kubernetes.io/name"

func CreateNamespaceManifest(name string) *apiv1.Namespace {
	return &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
}

func CreateDeploymentManifest(namespace, name string, dockerImage string, envs []apiv1.EnvVar) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				manifestLabelName: name,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					manifestLabelName: name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						manifestLabelName: name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  name,
							Image: dockerImage,
							Env:   envs,
						},
					},
				},
			},
		},
	}
}
func CreateServiceManifest(namespace, name string, port, targetPort int) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{manifestLabelName: name},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "app-port",
					Protocol:   "TCP",
					Port:       int32(port),
					TargetPort: intstr.FromInt(targetPort),
				},
			},
			Selector: map[string]string{manifestLabelName: name},
		},
		Status: apiv1.ServiceStatus{},
	}
}
