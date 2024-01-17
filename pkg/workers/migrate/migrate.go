package migrator

import (
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/subsystems"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
	subsystemsVM "go-deploy/models/sys/vm/subsystems"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

// Migrate will run as early as possible in the program, and it will never be called again.
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
//
// add a date to the migration name to make it easier to identify.
func getMigrations() map[string]func() error {
	return map[string]func() error{
		"migrateCreateAtFieldsForK8sSubsystems": migrateCreateAtFieldsForK8sSubsystems,
	}
}

func migrateCreateAtFieldsForK8sSubsystems() error {
	deployments, err := deploymentModels.New().List()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		k := &deployment.Subsystems.K8s

		fixK8sSubsystemsCreatedAt(k)

		err = deploymentModels.New().SetWithBsonByID(deployment.ID, bson.D{
			{"subsystems.k8s", k},
		})
		if err != nil {
			return err
		}
	}

	sms, err := smModels.New().List()
	if err != nil {
		return err
	}

	for _, sm := range sms {
		k := &sm.Subsystems.K8s

		fixK8sSubsystemsCreatedAt(k)

		err = smModels.New().SetWithBsonByID(sm.ID, bson.D{
			{"subsystems.k8s", k},
		})
		if err != nil {
			return err
		}
	}

	vms, err := vmModels.New().List()
	if err != nil {
		return err
	}

	for _, vm := range vms {
		k := &vm.Subsystems.K8s

		fixK8sSubsystemsCreatedAtVM(k)

		err = vmModels.New().SetWithBsonByID(vm.ID, bson.D{
			{"subsystems.k8s", k},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func fixK8sSubsystemsCreatedAtVM(k *subsystemsVM.K8s) {
	now := time.Now()

	if len(k.DeploymentMap) > 0 {
		for mapName, d := range k.DeploymentMap {
			if len(d.Name) > 0 && d.CreatedAt.IsZero() {
				d.CreatedAt = now
				k.DeploymentMap[mapName] = d
			}
		}
	}

	if len(k.ServiceMap) > 0 {
		for mapName, s := range k.ServiceMap {
			if len(s.Name) > 0 && s.CreatedAt.IsZero() {
				s.CreatedAt = now
				k.ServiceMap[mapName] = s
			}
		}
	}

	if len(k.IngressMap) > 0 {
		for mapName, i := range k.IngressMap {
			if len(i.Name) > 0 && i.CreatedAt.IsZero() {
				i.CreatedAt = now
				k.IngressMap[mapName] = i
			}
		}
	}

	if len(k.SecretMap) > 0 {
		for mapName, s := range k.SecretMap {
			if len(s.Name) > 0 && s.CreatedAt.IsZero() {
				s.CreatedAt = now
				k.SecretMap[mapName] = s
			}
		}
	}

	if len(k.Namespace.Name) > 0 && k.Namespace.CreatedAt.IsZero() {
		k.Namespace.CreatedAt = now
	}
}

func fixK8sSubsystemsCreatedAt(k *subsystems.K8s) {
	now := time.Now()

	if len(k.DeploymentMap) > 0 {
		for mapName, d := range k.DeploymentMap {
			if len(d.Name) > 0 && d.CreatedAt.IsZero() {
				d.CreatedAt = now
				k.DeploymentMap[mapName] = d
			}
		}
	}

	if len(k.ServiceMap) > 0 {
		for mapName, s := range k.ServiceMap {
			if len(s.Name) > 0 && s.CreatedAt.IsZero() {
				s.CreatedAt = now
				k.ServiceMap[mapName] = s
			}
		}
	}

	if len(k.IngressMap) > 0 {
		for mapName, i := range k.IngressMap {
			if len(i.Name) > 0 && i.CreatedAt.IsZero() {
				i.CreatedAt = now
				k.IngressMap[mapName] = i
			}
		}
	}

	if len(k.PvMap) > 0 {
		for mapName, pv := range k.PvMap {
			if len(pv.Name) > 0 && pv.CreatedAt.IsZero() {
				pv.CreatedAt = now
				k.PvMap[mapName] = pv
			}
		}
	}

	if len(k.PvcMap) > 0 {
		for mapName, pvc := range k.PvcMap {
			if len(pvc.Name) > 0 && pvc.CreatedAt.IsZero() {
				pvc.CreatedAt = now
				k.PvcMap[mapName] = pvc
			}
		}
	}

	if len(k.JobMap) > 0 {
		for mapName, j := range k.JobMap {
			if len(j.Name) > 0 && j.CreatedAt.IsZero() {
				j.CreatedAt = now
				k.JobMap[mapName] = j
			}
		}
	}

	if len(k.SecretMap) > 0 {
		for mapName, s := range k.SecretMap {
			if len(s.Name) > 0 && s.CreatedAt.IsZero() {
				s.CreatedAt = now
				k.SecretMap[mapName] = s
			}
		}
	}

	if len(k.HpaMap) > 0 {
		for mapName, h := range k.HpaMap {
			if len(h.Name) > 0 && h.CreatedAt.IsZero() {
				h.CreatedAt = now
				k.HpaMap[mapName] = h
			}
		}
	}

	if len(k.Namespace.Name) > 0 && k.Namespace.CreatedAt.IsZero() {
		k.Namespace.CreatedAt = now
	}
}
