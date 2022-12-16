package db

import "context"

type GroupQuery interface {
	// Save group
	Save(g Group) (Group, error)

	// Update a group
	Update(g Group) (Group, error)

	// Delete a group
	Delete(ctx context.Context, id string) error

	// RetrieveByID retrieves group by its id
	RetrieveByID(ctx context.Context, id string) (Group, error)

	// RetrieveAll retrieves all groups.
	RetrieveAll(ctx context.Context, pm PageMetadata) (GroupPage, error)

	// RetrieveAllParents retrieves all groups that are ancestors to the group with given groupID.
	RetrieveAllParents(ctx context.Context, groupID string, pm PageMetadata) (GroupPage, error)

	// RetrieveAllChildren retrieves all children from group with given groupID up to the hierarchy level.
	RetrieveAllChildren(ctx context.Context, groupID string, pm PageMetadata) (GroupPage, error)

	//  Retrieves list of groups that member belongs to
	Memberships(ctx context.Context, memberID string, pm PageMetadata) (GroupPage, error)

	// Members retrieves everything that is assigned to a group identified by groupID.
	Members(ctx context.Context, groupID, groupType string, pm PageMetadata) (MemberPage, error)

	// Assign adds a member to group.
	Assign(ctx context.Context, groupID, groupType string, memberIDs ...string) error

	// Unassign removes a member from a group
	Unassign(ctx context.Context, groupID string, memberIDs ...string) error
}
