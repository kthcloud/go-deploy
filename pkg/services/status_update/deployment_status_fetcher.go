package status_update

import (
	"go-deploy/models/model"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/service"
)

func DeploymentStatusFetcher() error {
	drc := deployment_repo.New()

	// Enabled Deployments have a correlating Pod in the cluster
	// So we can list all the statuses in the cluster and match them with the Deployments
	var allStatus []model.DeploymentStatus

	for _, zone := range config.Config.EnabledZones() {
		z := zone
		statuses, err := service.V1().Deployments().K8s().ListDeploymentStatus(&z)
		if err != nil {
			return err
		}

		allStatus = append(allStatus, statuses...)
	}

	for _, status := range allStatus {
		err := drc.SetStatusByName(status.Name, parseDeploymentStatus(&status), &model.ReplicaStatus{
			DesiredReplicas:     status.DesiredReplicas,
			ReadyReplicas:       status.ReadyReplicas,
			AvailableReplicas:   status.AvailableReplicas,
			UnavailableReplicas: status.UnavailableReplicas,
		})
		if err != nil {
			return err
		}
	}

	// Disabled Deployments have no correlating Pod in the cluster, and uses a fallback Pod
	disabledDeployments, err := deployment_repo.New().WithZone(config.Config.EnabledZoneNames()...).Disabled().List()
	if err != nil {
		return err
	}

	for _, deployment := range disabledDeployments {
		err = drc.SetStatusByName(deployment.Name, status_codes.GetMsg(status_codes.ResourceDisabled), model.EmptyReplicaStatus)
		if err != nil {
			return err
		}
	}

	return nil
}
