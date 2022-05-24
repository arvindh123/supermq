package policies

import initutil "github.com/mainflux/mainflux/internal/init"

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
		return initutil.ErrBearerToken
	}

	if len(req.SubjectIDs) == 0 {
		return initutil.ErrEmptyList
	}

	if len(req.Policies) == 0 {
		return initutil.ErrEmptyList
	}

	if req.Object == "" {
		return initutil.ErrMissingPolicyObj
	}

	for _, policy := range req.Policies {
		if _, ok := actions[policy]; !ok {
			return initutil.ErrMalformedPolicy
		}
	}

	for _, subID := range req.SubjectIDs {
		if subID == "" {
			return initutil.ErrMissingPolicySub
		}
	}

	return nil
}
