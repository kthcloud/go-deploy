package k8s

import (
	"context"

	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/keys"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const (
	// PodEventAdded is emitted when a pod is added
	PodEventAdded = "added"
	// PodEventDeleted is emitted when a pod is deleted
	PodEventDeleted = "deleted"
	// PodEventUpdated is emitted when a pod is updated
	PodEventUpdated = "updated"
)

// CountPods returns a list of pods in the cluster.
func (client *Client) CountPods() (int, error) {
	pods, err := client.K8sClient.CoreV1().Pods(client.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return 0, err
	}

	return len(pods.Items), nil
}

// ListPods returns a list of pods in the cluster.
func (client *Client) ListPods() ([]models.PodPublic, error) {
	pods, err := client.K8sClient.CoreV1().Pods(client.Namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var res []models.PodPublic
	for _, pod := range pods.Items {
		res = append(res, *models.CreatePodPublicFromRead(pod))
	}

	return res, nil
}

// PodExists checks if a pod exists in the cluster.
func (client *Client) PodExists(podName string) (bool, error) {
	_, err := client.K8sClient.CoreV1().Pods(client.Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		if IsNotFoundErr(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// SetupPodWatcher is a function that sets up a pod watcher with a callback.
// It triggers the callback when a pod event occurs.
func (client *Client) SetupPodWatcher(ctx context.Context, callback func(podName, event string)) error {
	factory := informers.NewSharedInformerFactoryWithOptions(client.K8sClient, 0, informers.WithNamespace(client.Namespace))
	podInformer := factory.Core().V1().Pods().Informer()

	// Returns the name of the deployment, when it was created and whether the pod is allowed
	allowedPod := func(pod *v1.Pod) bool {
		if _, ok := pod.Labels[keys.LabelDeployName]; !ok {
			return false
		}

		allowedStatuses := []v1.PodPhase{
			v1.PodRunning,
			v1.PodFailed,
		}

		allowed := false
		for _, status := range allowedStatuses {
			if pod.Status.Phase == status {
				allowed = true
				break
			}
		}

		return allowed
	}

	_, err := podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				return
			}

			if !allowedPod(pod) {
				return
			}

			callback(pod.Name, PodEventAdded)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			pod, ok := newObj.(*v1.Pod)
			if !ok {
				return
			}

			if !allowedPod(pod) {
				return
			}

			callback(pod.Name, PodEventUpdated)
		},
		DeleteFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				return
			}

			if _, ok = pod.Labels[keys.LabelDeployName]; !ok {
				return
			}

			callback(pod.Name, PodEventDeleted)
		},
	})
	if err != nil {
		return err
	}

	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())

	return nil
}
