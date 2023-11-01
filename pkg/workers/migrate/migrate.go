package migrator

import (
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"log"
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
		"addTlsSecretNameForIngressesWithoutCustomCert_2023_10_19": addTlsSecretNameForIngressesWithoutCustomCert_2023_10_19,
	}
}

func addTlsSecretNameForIngressesWithoutCustomCert_2023_10_19() error {
	deployments, err := deploymentModel.New().ListAll()
	if err != nil {
		return err
	}

	for _, deployment := range deployments {
		if k8sIngress := deployment.Subsystems.K8s.GetIngress(deployment.Name); service.Created(k8sIngress) {
			if k8sIngress.CustomCert != nil {
				continue
			}

			if k8sIngress.TlsSecret != nil {
				continue
			}

			name := "wildcard-cert"
			k8sIngress.TlsSecret = &name

			deployment.Subsystems.K8s.SetIngress(deployment.Name, *k8sIngress)

			err := deploymentModel.New().UpdateSubsystemByID(deployment.ID, "k8s.ingressMap", deployment.Subsystems.K8s.IngressMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
