package k8s

import (
	"context"
	"fmt"
	"go-deploy/utils/subsystemutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

func Restart(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart k8s for project %s. details: %s", name, err)
	}

	req := client.AppsV1().Deployments(subsystemutils.GetPrefixedName(name))

	data := fmt.Sprintf(`{"spec": {"template": {"metadata": {"annotations": {"kubectl.kubernetes.io/restartedAt": "%s"}}}}}`, time.Now().Format("20060102150405"))

	_, err := req.Patch(context.TODO(), name, types.StrategicMergePatchType, []byte(data), metav1.PatchOptions{})
	if err != nil {
		return makeError(err)
	}

	return nil
}
