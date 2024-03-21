package model

import "time"

const (
	// TeamResourceDeployment is the type used for deployment resources in a team.
	TeamResourceDeployment = "deployment"
	// TeamResourceVM is the type used for VM resources in a team.
	TeamResourceVM = "vm"

	// TeamMemberRoleAdmin is the role used for admin members in a team.
	// This is currently not used, and every member is an admin.
	TeamMemberRoleAdmin = "admin"

	// TeamMemberStatusInvited is the status used for users that have been invited to a team.
	TeamMemberStatusInvited = "invited"
	// TeamMemberStatusJoined is the status used for users that have joined a team.
	TeamMemberStatusJoined = "joined"
)

type TeamMember struct {
	// ID is the same as UserID
	ID           string `bson:"id"`
	TeamRole     string `bson:"teamRole"`
	MemberStatus string `bson:"memberStatus"`

	AddedAt  time.Time `bson:"addedAt"`
	JoinedAt time.Time `bson:"joinedAt"`

	InvitationCode string `bson:"invitationCode"`
}

type TeamResource struct {
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

	ResourceMap map[string]TeamResource `bson:"resourceMap"`
	MemberMap   map[string]TeamMember   `bson:"memberMap"`
}

func (t *Team) GetID() string {
	return t.ID
}

func (t *Team) GetMemberMap() map[string]TeamMember {
	if t.MemberMap == nil {
		t.MemberMap = make(map[string]TeamMember)
	}

	return t.MemberMap
}

func (t *Team) GetMember(id string) *TeamMember {
	res, ok := t.GetMemberMap()[id]
	if !ok {
		return nil
	}

	return &res
}

func (t *Team) GetResourceMap() map[string]TeamResource {
	if t.ResourceMap == nil {
		t.ResourceMap = make(map[string]TeamResource)
	}

	return t.ResourceMap
}

func (t *Team) GetResource(id string) *TeamResource {
	res, ok := t.GetResourceMap()[id]
	if !ok {
		return nil
	}

	return &res
}

func (t *Team) AddMember(member TeamMember) {
	t.GetMemberMap()[member.ID] = member
}

func (t *Team) AddResource(resource TeamResource) {
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
