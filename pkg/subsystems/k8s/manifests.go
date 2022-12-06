package k8s

import (
	"deploy-api-go/pkg/conf"
	"deploy-api-go/utils/subsystemutils"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func getNamespaceManifest(name string) *apiv1.Namespace {
	return &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: subsystemutils.GetPrefixedName(name),
		},
	}
}

func getDeploymentManifest(name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: subsystemutils.GetPrefixedName(name),
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
							Image: getDockerImageName(name),
							Env: []apiv1.EnvVar{
								{
									Name:  "DEPLOY_APP_PORT",
									Value: string(rune(conf.Env.AppPort)),
								},
							},
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: int32(conf.Env.AppPort),
								},
							},
						},
					},
				},
			},
		},
	}
}

func getServiceManifest(name string) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: subsystemutils.GetPrefixedName(name),
			Labels:    map[string]string{manifestLabelName: name},
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Name:       "app-port",
					Protocol:   "TCP",
					Port:       int32(conf.Env.AppPort),
					TargetPort: intstr.FromInt(int(conf.Env.AppPort)),
				},
			},
			Selector: map[string]string{manifestLabelName: name},
		},
		Status: apiv1.ServiceStatus{},
	}
}
