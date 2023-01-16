// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"time"

	"github.com/mainflux/mainflux/certs"
)

type pageRes struct {
	Total  uint64 `json:"total"`
	Offset uint64 `json:"offset"`
	Limit  int64  `json:"limit"`
}

type certsPageRes struct {
	pageRes
	Certs []certsRes `json:"certs"`
}

type certsRes struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	OwnerID     string    `json:"owner_id"`
	ThingID     string    `json:"thing_id"`
	Serial      string    `json:"serial"`
	Certificate string    `json:"certificate"`
	PrivateKey  string    `json:"private_key"`
	CAChain     string    `json:"ca_chain"`
	IssuingCA   string    `json:"issuing_ca"`
	KeyType     string    `json:"key_type"`
	KeyBits     int       `json:"key_bits"`
	TTL         string    `json:"ttl"`
	Expire      time.Time `json:"expire"`
	Revocation  time.Time `json:"revocation"`
	created     bool
}

type revokeCertsRes struct {
	RevocationTime time.Time `json:"revocation_time"`
}

func (res certsPageRes) Code() int {
	return http.StatusOK
}

func (res certsPageRes) Headers() map[string]string {
	return map[string]string{}
}

func (res certsPageRes) Empty() bool {
	return false
}

func (res certsRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res certsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res certsRes) Empty() bool {
	return false
}

func CertToCertResponse(cert certs.Cert, created bool) certsRes {
	return certsRes{
		ID:          cert.ID,
		Name:        cert.Name,
		OwnerID:     cert.OwnerID,
		ThingID:     cert.ThingID,
		Serial:      cert.Serial,
		Certificate: cert.Certificate,
		PrivateKey:  cert.PrivateKey,
		CAChain:     cert.CAChain,
		IssuingCA:   cert.IssuingCA,
		TTL:         cert.TTL,
		Expire:      cert.Expire,
		Revocation:  cert.Revocation,
		created:     created,
	}
}
