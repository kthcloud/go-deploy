package k8s_service

import (
	"fmt"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/gpu_claim_repo"
	"github.com/kthcloud/go-deploy/service/resources"
)

func (c *Client) Create(id string, params *model.GpuClaimCreateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create gpu claim in k8s. details: %w", err)
	}

	_, kc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return makeError(err)
	}

	// Namespace
	err = resources.SsCreator(kc.CreateNamespace).
		WithDbFunc(dbFunc(id, "namespace")).
		WithPublic(g.Namespace()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	for _, public := range g.ResourceClaims() {
		err = resources.SsCreator(kc.CreateResourceClaim).
			WithDbFunc(dbFunc(id, "resourceClaimMap."+public.Name)).
			WithPublic(&public).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func (c *Client) Delete(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create gpu claim in k8s. details: %w", err)
	}

	gc, kc, _, err := c.Get(OptsNoGenerator(id))
	if err != nil {
		return makeError(err)
	}

	for mapName, rc := range gc.Subsystems.K8s.ResourceClaimMap {
		err = resources.SsDeleter(kc.DeleteResourceClaim).
			WithResourceID(rc.Name).
			WithDbFunc(dbFunc(id, "resourceClaimMap."+mapName)).
			Exec()
	}

	// Namespace
	// (not deleted in k8s, since it is shared)
	err = resources.SsDeleter(func(string) error { return nil }).
		WithResourceID(gc.Subsystems.K8s.Namespace.Name).
		WithDbFunc(dbFunc(id, "namespace")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	return nil
}

// dbFunc returns a function that updates the K8s subsystem.
func dbFunc(id, key string) func(any) error {
	return func(data any) error {
		if data == nil {
			return gpu_claim_repo.New().DeleteSubsystem(id, "k8s."+key)
		}
		return gpu_claim_repo.New().SetSubsystem(id, "k8s."+key, data)
	}
}
