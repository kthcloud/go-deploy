package team

import "time"

const (
	// ResourceTypeDeployment is the type used for deployment resources in a team.
	ResourceTypeDeployment = "deployment"
	// ResourceTypeVM is the type used for VM resources in a team.
	ResourceTypeVM = "vm"

	// MemberRoleAdmin is the role used for admin members in a team.
	// This is currently not used, and every member is an admin.
	MemberRoleAdmin = "admin"

	// MemberStatusInvited is the status used for users that have been invited to a team.
	MemberStatusInvited = "invited"
	// MemberStatusJoined is the status used for users that have joined a team.
	MemberStatusJoined = "joined"
)

type Member struct {
	// ID is the same as UserID
	ID           string `bson:"id"`
	TeamRole     string `bson:"teamRole"`
	MemberStatus string `bson:"memberStatus"`

	AddedAt  time.Time `bson:"addedAt"`
	JoinedAt time.Time `bson:"joinedAt"`

	InvitationCode string `bson:"invitationCode"`
}

type Resource struct {
	ID      string    `bson:"id"`
	Type    string    `bson:"type"`
	AddedAt time.Time `bson:"addedAt"`
}

type Team struct {
	ID          string    `bson:"id"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	OwnerID     string    `bson:"ownerId"`
	CreatedAt   time.Time `bson:"createdAt"`
	UpdatedAt   time.Time `bson:"updatedAt"`

	ResourceMap map[string]Resource `bson:"resourceMap"`
	MemberMap   map[string]Member   `bson:"memberMap"`
}

func (t *Team) GetID() string {
	return t.ID
}

func (t *Team) GetMemberMap() map[string]Member {
	if t.MemberMap == nil {
		t.MemberMap = make(map[string]Member)
	}

	return t.MemberMap
}

func (t *Team) GetMember(id string) *Member {
	res, ok := t.GetMemberMap()[id]
	if !ok {
		return nil
	}

	return &res
}

func (t *Team) GetResourceMap() map[string]Resource {
	if t.ResourceMap == nil {
		t.ResourceMap = make(map[string]Resource)
	}

	return t.ResourceMap
}

func (t *Team) GetResource(id string) *Resource {
	res, ok := t.GetResourceMap()[id]
	if !ok {
		return nil
	}

	return &res
}

func (t *Team) AddMember(member Member) {
	t.GetMemberMap()[member.ID] = member
}

func (t *Team) AddResource(resource Resource) {
	t.GetResourceMap()[resource.ID] = resource
}

func (t *Team) RemoveMember(id string) {
	delete(t.GetMemberMap(), id)
}

func (t *Team) RemoveResource(id string) {
	delete(t.GetResourceMap(), id)
}

func (t *Team) HasMember(id string) bool {
	_, ok := t.GetMemberMap()[id]
	return ok
}

func (t *Team) HasResource(id string) bool {
	_, ok := t.GetResourceMap()[id]
	return ok
}
