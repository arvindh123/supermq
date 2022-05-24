package api

import initutil "github.com/mainflux/mainflux/internal/init"

type provisionReq struct {
	token       string
	Name        string `json:"name"`
	ExternalID  string `json:"external_id"`
	ExternalKey string `json:"external_key"`
}

func (req provisionReq) validate() error {
	if req.ExternalID == "" {
		return initutil.ErrMissingID
	}

	if req.ExternalKey == "" {
		return initutil.ErrBearerKey
	}

	return nil
}

type mappingReq struct {
	token string
}

func (req mappingReq) validate() error {
	if req.token == "" {
		return initutil.ErrBearerToken
	}
	return nil
}
