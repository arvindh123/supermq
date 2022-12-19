package db

import "github.com/mainflux/mainflux/auth"

type QueryFramer interface {
	Save(g auth.Group) string
	Update() string
	Delete() string
	RetrieveByID() string
	RetrieveAllParents() (string, string)
	RetrieveAllChildren() (string, string)
	RetrieveAll(pm auth.PageMetadata) (string, string, error)

	Members(groupID string, pm auth.PageMetadata) (string, string, error)
	Memberships(pm auth.PageMetadata) (string, string, error)
	Assign() string
	Unassign() string
}
