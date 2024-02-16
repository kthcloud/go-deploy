package k8s

import (
	"go-deploy/pkg/subsystems/k8s/models"
	"testing"
)

func TestCreateNetworkPolicy(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	withDefaultNetworkPolicy(t, c)
}

func TestUpdateNetworkPolicy(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t)
	np := withDefaultNetworkPolicy(t, c)

	np.EgressRules = append(np.EgressRules, models.EgressRule{
		CIDR: "1.1.1.1/32",
	})

	rulesCount := len(np.EgressRules)

	updated, err := c.UpdateNetworkPolicy(np)
	if err != nil {
		t.Fatalf("failed to update network policy: %s", err)
	}

	if len(updated.EgressRules) != rulesCount {
		t.Fatalf("expected %d egress rules, got %d", rulesCount, len(updated.EgressRules))
	}
}
