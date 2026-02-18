package models

type ResourceClaimStatus struct {
	Allocated         bool                                  `bson:"allocated"`
	AllocationResults []ResourceClaimAllocationResultPublic `bson:"allocationResults,omitempty"`
	Consumers         []ResourceClaimConsumerPublic         `bson:"consumers,omitempty"`
}
