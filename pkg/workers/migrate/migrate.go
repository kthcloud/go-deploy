package migrator

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/pkg/subsystems/k8s"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

// Migrate will  run as early as possible in the program, and it will never be called again.
func Migrate() {
	migrations := getMigrations()
	if len(migrations) > 0 {
		log.Println("migrating...")

		for name, migration := range migrations {
			log.Printf("running migration %s", name)
			if err := migration(); err != nil {
				log.Fatalf("failed to run migration %s. details: %s", name, err)
			}
		}

		log.Println("migrations done")
		return
	}

	log.Println("nothing to migrate")
}

// getMigrations returns a map of migrations to run.
// add a migration to the list of functions to run.
// clear when prod has run it once.
//
// the migrations must be **idempotent**.
func getMigrations() map[string]func() error {
	return map[string]func() error{
		"add root network id and root id if missing": addRootNetworkIdAndRootIdIfMissing,
		"update port is zero for deployments":        updatePortIsZeroForDeployments,
		"fetch creation time for subsystems":         fetchCreationTimeForSubsystems,
	}
}

func addRootNetworkIdAndRootIdIfMissing() error {
	vms, err := vmModel.New().GetAll()
	if err != nil {
		return fmt.Errorf("error fetching vms. details: %w", err)
	}

	for _, vm := range vms {
		zone := conf.Env.VM.GetZone(vm.Zone)
		if zone == nil {
			return fmt.Errorf("zone %s not found", vm.Zone)
		}

		for mapName, pfr := range vm.Subsystems.CS.GetPortForwardingRuleMap() {
			if !pfr.Created() {
				continue
			}

			if pfr.NetworkID == "" {
				pfr.NetworkID = zone.NetworkID
			}

			if pfr.IpAddressID == "" {
				pfr.IpAddressID = zone.IpAddressID
			}

			vm.Subsystems.CS.SetPortForwardingRule(mapName, pfr)
		}

		err = vmModel.New().UpdateWithBsonByID(vm.ID, bson.D{
			{"subsystems.cs.portForwardingRuleMap", vm.Subsystems.CS.GetPortForwardingRuleMap()},
		})

		if err != nil {
			return fmt.Errorf("error updating vm. details: %w", err)
		}
	}

	return nil
}

func updatePortIsZeroForDeployments() error {
	deployments, err := deploymentModel.New().GetAll()
	if err != nil {
		return fmt.Errorf("error fetching deployments. details: %w", err)
	}

	for _, deployment := range deployments {
		mainApp := deployment.GetMainApp()
		if mainApp == nil {
			continue
		}

		if mainApp.InternalPort == 0 {
			mainApp.InternalPort = conf.Env.Deployment.Port
		}

		deployment.SetMainApp(mainApp)

		err = deploymentModel.New().UpdateWithBsonByID(deployment.ID, bson.D{
			{"apps", deployment.Apps},
		})
		if err != nil {
			return fmt.Errorf("error updating deployment. details: %w", err)
		}

	}

	return nil
}

func fetchCreationTimeForSubsystems() error {

	deployments, err := deploymentModel.New().GetAll()
	if err != nil {
		return fmt.Errorf("error fetching deployments. details: %w", err)
	}

	for _, deployment := range deployments {

		zone := conf.Env.Deployment.GetZone(deployment.Zone)
		if zone == nil {
			return fmt.Errorf("zone %s not found", deployment.Zone)
		}

		{
			ss := &deployment.Subsystems.K8s

			client, err := k8s.New(zone.Client, ss.Namespace.FullName)
			if err != nil {
				return fmt.Errorf("error creating k8s client. details: %w", err)
			}

			if ss.Namespace.Created() {
				ns, err := client.ReadNamespace(ss.Namespace.ID)
				if err != nil {
					return fmt.Errorf("error fetching namespace creation time. details: %w", err)
				}

				if ns != nil {
					ss.Namespace.CreatedAt = ns.CreatedAt
				}
			}

			for _, k8sDeployment := range ss.GetDeploymentMap() {
				if k8sDeployment.Created() {
					dep, err := client.ReadDeployment(k8sDeployment.ID)
					if err != nil {
						return fmt.Errorf("error fetching deployment creation time. details: %w", err)
					}

					if dep != nil {
						k8sDeployment.CreatedAt = dep.CreatedAt
					}
				}
			}

			for _, k8sService := range ss.GetServiceMap() {
				if k8sService.Created() {
					svc, err := client.ReadService(k8sService.ID)
					if err != nil {
						return fmt.Errorf("error fetching service creation time. details: %w", err)
					}

					if svc != nil {
						k8sService.CreatedAt = svc.CreatedAt
					}
				}
			}

			for _, k8sIngress := range ss.GetIngressMap() {
				if k8sIngress.Created() {
					ing, err := client.ReadIngress(k8sIngress.ID)
					if err != nil {
						return fmt.Errorf("error fetching ingress creation time. details: %w", err)
					}

					if ing != nil {
						k8sIngress.CreatedAt = ing.CreatedAt
					}
				}
			}

			for _, k8sPvc := range ss.GetPvcMap() {
				if k8sPvc.Created() {
					pvc, err := client.ReadPVC(k8sPvc.ID)
					if err != nil {
						return fmt.Errorf("error fetching pvc creation time. details: %w", err)
					}

					if pvc != nil {
						k8sPvc.CreatedAt = pvc.CreatedAt
					}
				}
			}

			for _, k8sPv := range ss.GetPvMap() {
				if k8sPv.Created() {
					pv, err := client.ReadPV(k8sPv.ID)
					if err != nil {
						return fmt.Errorf("error fetching pv creation time. details: %w", err)
					}

					if pv != nil {
						k8sPv.CreatedAt = pv.CreatedAt
					}
				}
			}

			for _, k8sJob := range ss.GetJobMap() {
				if k8sJob.Created() {
					job, err := client.ReadJob(k8sJob.ID)
					if err != nil {
						return fmt.Errorf("error fetching job creation time. details: %w", err)
					}

					if job != nil {
						k8sJob.CreatedAt = job.CreatedAt
					}
				}
			}
		}

		{
			ss := &deployment.Subsystems.Harbor

			harborClient, err := harbor.New(&harbor.ClientConf{
				ApiUrl:   conf.Env.Harbor.URL,
				Username: conf.Env.Harbor.User,
				Password: conf.Env.Harbor.Password,
			})

			if err != nil {
				return fmt.Errorf("error creating harbor client. details: %w", err)
			}

			if ss.Project.Created() {
				project, err := harborClient.ReadProject(ss.Project.ID)
				if err != nil {
					return fmt.Errorf("error fetching project creation time. details: %w", err)
				}

				if project != nil {
					ss.Project.CreatedAt = project.CreatedAt
				}
			}

			if ss.Robot.Created() {
				robot, err := harborClient.ReadRobot(ss.Robot.ID)
				if err != nil {
					return fmt.Errorf("error fetching robot creation time. details: %w", err)
				}

				if robot != nil {
					ss.Robot.CreatedAt = robot.CreatedAt
				}
			}

			if ss.Repository.Created() {
				if ss.Repository.Placeholder == nil {
					ss.Repository.Placeholder = &models.PlaceHolder{
						ProjectName:    conf.Env.DockerRegistry.Placeholder.Project,
						RepositoryName: conf.Env.DockerRegistry.Placeholder.Repository,
					}
				}

				repo, err := harborClient.ReadRepository(ss.Project.Name, ss.Repository.Name)
				if err != nil {
					return fmt.Errorf("error fetching repository creation time. details: %w", err)
				}

				if repo != nil {
					ss.Repository.CreatedAt = repo.CreatedAt
				}
			}

		}

		err = deploymentModel.New().UpdateWithBsonByID(deployment.ID, bson.D{
			{"subsystems", deployment.Subsystems},
		})

		if err != nil {
			return fmt.Errorf("error updating deployment subsystems. details: %w", err)
		}
	}

	vms, err := vmModel.New().GetAll()
	if err != nil {
		return fmt.Errorf("error fetching vms. details: %w", err)
	}

	for _, vm := range vms {
		zone := conf.Env.VM.GetZone(vm.Zone)
		if zone == nil {
			return fmt.Errorf("zone %s not found", vm.Zone)
		}

		client, err := cs.New(&cs.ClientConf{
			URL:         conf.Env.CS.URL,
			ApiKey:      conf.Env.CS.ApiKey,
			Secret:      conf.Env.CS.Secret,
			ZoneID:      zone.ZoneID,
			ProjectID:   zone.ProjectID,
			IpAddressID: zone.IpAddressID,
			NetworkID:   zone.NetworkID,
		})

		if err != nil {
			return fmt.Errorf("error creating cs client. details: %w", err)
		}

		if vm.Subsystems.CS.VM.Created() {
			csVM, err := client.ReadVM(vm.Subsystems.CS.VM.ID)
			if err != nil {
				return fmt.Errorf("error fetching vm creation time. details: %w", err)
			}

			if csVM != nil {
				vm.Subsystems.CS.VM.CreatedAt = csVM.CreatedAt
			}
		}

		if vm.Subsystems.CS.ServiceOffering.Created() {
			so, err := client.ReadServiceOffering(vm.Subsystems.CS.ServiceOffering.ID)
			if err != nil {
				return fmt.Errorf("error fetching service offering creation time. details: %w", err)
			}

			if so != nil {
				vm.Subsystems.CS.ServiceOffering.CreatedAt = so.CreatedAt
			}
		}

		for _, pfr := range vm.Subsystems.CS.GetPortForwardingRuleMap() {
			if pfr.Created() {
				csPfr, err := client.ReadPortForwardingRule(pfr.ID)
				if err != nil {
					return fmt.Errorf("error fetching port forwarding rule creation time. details: %w", err)
				}

				if csPfr != nil {
					pfr.CreatedAt = csPfr.CreatedAt
				}
			}
		}

		err = vmModel.New().UpdateWithBsonByID(vm.ID, bson.D{
			{"subsystems", vm.Subsystems},
		})

		if err != nil {
			return fmt.Errorf("error updating vm subsystems. details: %w", err)
		}
	}

	return nil
}
