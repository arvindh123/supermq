package policies

import "github.com/mainflux/mainflux/internal"

// Action represents an enum for the policies used in the Mainflux.
type Action int

const (
	Create Action = iota
	Read
	Write
	Delete
	Access
	Member
	Unknown
)

var actions = map[string]Action{
	"create": Create,
	"read":   Read,
	"write":  Write,
	"delete": Delete,
	"access": Access,
	"member": Member,
}

type policiesReq struct {
	token      string
	SubjectIDs []string `json:"subjects"`
	Policies   []string `json:"policies"`
	Object     string   `json:"object"`
}

func (req policiesReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if len(req.SubjectIDs) == 0 {
		return internal.ErrEmptyList
	}

	if len(req.Policies) == 0 {
		return internal.ErrEmptyList
	}

	if req.Object == "" {
		return internal.ErrMissingPolicyObj
	}

	for _, policy := range req.Policies {
		if _, ok := actions[policy]; !ok {
			return internal.ErrMalformedPolicy
		}
	}

	for _, subID := range req.SubjectIDs {
		if subID == "" {
			return internal.ErrMissingPolicySub
		}
	}

	return nil
}
