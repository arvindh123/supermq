package api

import "github.com/mainflux/mainflux/internal"

type provisionReq struct {
	token       string
	Name        string `json:"name"`
	ExternalID  string `json:"external_id"`
	ExternalKey string `json:"external_key"`
}

func (req provisionReq) validate() error {
	if req.ExternalID == "" {
		return internal.ErrMissingID
	}

	if req.ExternalKey == "" {
		return internal.ErrBearerKey
	}

	return nil
}

type mappingReq struct {
	token string
}

func (req mappingReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}
	return nil
}
