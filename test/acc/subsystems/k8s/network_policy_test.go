package k8s

import (
	"github.com/stretchr/testify/assert"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/models"
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
		IpBlock: &models.IpBlock{
			CIDR:   "0.0.0.0/0",
			Except: []string{"2.2.2.2/32"},
		},
	})

	rulesCount := len(np.EgressRules)

	updated, err := c.UpdateNetworkPolicy(np)
	if err != nil {
		t.Fatalf("failed to update network policy: %s", err)
	}

	if len(updated.EgressRules) != rulesCount {
		t.Fatalf("expected %d egress rules, got %d", rulesCount, len(updated.EgressRules))
	}

	found := false
	for _, rule := range updated.EgressRules {
		if rule.IpBlock != nil {
			for _, except := range rule.IpBlock.Except {
				if except == "2.2.2.2/32" {
					found = true
					break
				}
			}
		}

		if found {
			break
		}
	}

	assert.True(t, found, "expected to find 2.2.2.2/32 in egress rules")
}
