package team

import "time"

type Member struct {
	// ID is UserID
	ID       string    `bson:"id"`
	TeamRole string    `bson:"teamRole"`
	JoinedAt time.Time `bson:"joinedAt,omitempty"`
}

type Team struct {
	ID        string            `bson:"id"`
	Name      string            `bson:"name"`
	MemberMap map[string]Member `bson:"memberMap"`
}

func (t *Team) GetMemberMap() map[string]Member {
	if t.MemberMap == nil {
		t.MemberMap = make(map[string]Member)
	}

	return t.MemberMap
}

func (t *Team) AddMember(member Member) {
	t.GetMemberMap()[member.ID] = member
}

func (t *Team) RemoveMember(id string) {
	delete(t.GetMemberMap(), id)
}

func (t *Team) HasMember(id string) bool {
	_, ok := t.GetMemberMap()[id]
	return ok
}
