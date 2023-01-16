// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"context"
	"crypto/x509"
	"strings"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/certs/pki"
	"github.com/mainflux/mainflux/internal/client/events/things"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
)

// Key types and format : https://developer.hashicorp.com/vault/api-docs/secret/pki#key_type
const (
	caChainJoinSep = "\n\n"
	keyFormat      = "der"
	certFormat     = "pem"
)

var (
	ErrThingRetrieve = errors.New("failed to retrieve thing details")

	ErrPKIIssue = errors.New("failed to issue certificate in PKI")

	errPKIRevoke = errors.New("failed to revoke certificate in PKI")

	errRepoRetrieve = errors.New("failed to retrieve certificate from repo")

	errRepoUpdate = errors.New("failed to update certificate from repo")

	errRepoRemove = errors.New("failed to remove the certificate from db")

	errParseCert = errors.New("failed to parse the certificate, invalid certificate")
)

var _ Service = (*certsService)(nil)

var _ things.EventHandler = (*thingsEventHandlers)(nil)

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// IssueCert issues certificate for given thing id if access is granted with token
	IssueCert(ctx context.Context, token, thingID, name, ttl string) (Cert, error)

	// ViewCert retrieves the certificate issued for a given certificate ID
	ViewCert(ctx context.Context, token, certID string) (Cert, error)

	// RenewCert the expired certificate from certs repo
	RenewCert(ctx context.Context, token, certID string) (Cert, error)

	// RevokeCert revokes a certificate for a given certificate ID
	RevokeCert(ctx context.Context, token, certID string) error

	// RemoveCert revoke and delete entry  the certificate for a given certificate ID
	RemoveCert(ctx context.Context, token, certID string) error

	// ListCerts lists certificates issued for a given certificate ID
	ListCerts(ctx context.Context, token, certID, thingID, serial, name string, offset, limit uint64) (Page, error)

	// RevokeThingCerts revokes a all the certificates for a given thing ID with given limited count
	RevokeThingCerts(ctx context.Context, token, thingID string, limit int64) error

	// RenewThingCerts renew all the certificates for a given thing ID with given limited count
	RenewThingCerts(ctx context.Context, token, thingID string, limit int64) error

	// RemoveThingCerts revoke and delete entries of all the certificate for a given thing ID with given limited count
	RemoveThingCerts(ctx context.Context, token, certID string, limit int64) error
}

type certsService struct {
	auth       mainflux.AuthServiceClient
	idProvider mainflux.IDProvider
	repo       Repository
	sdk        mfsdk.SDK
	pki        pki.Agent
}

type thingsEventHandlers struct {
	pki  pki.Agent
	repo Repository
}

// New returns new Certs service.
func New(auth mainflux.AuthServiceClient, repo Repository, idp mainflux.IDProvider, pki pki.Agent) Service {
	return &certsService{
		repo:       repo,
		idProvider: idp,
		auth:       auth,
		pki:        pki,
	}
}

func NewThingsEventHandlers(repo Repository, pki pki.Agent) things.EventHandler {
	return &thingsEventHandlers{repo: repo, pki: pki}
}

// Revoke defines the conditions to revoke a certificate
type Revoke struct {
	RevocationTime time.Time `mapstructure:"revocation_time"`
}

// Cert defines the certificate paremeters
type Cert struct {
	ID          string    `json:"id"            db:"id"`
	Name        string    `json:"name"          db:"name"`
	OwnerID     string    `json:"owner_id"      db:"owner_id"`
	ThingID     string    `json:"thing_id"      db:"thing_id"`
	Serial      string    `json:"serial"        db:"serial"`
	Certificate string    `json:"certificate"   db:"certificate"`
	PrivateKey  string    `json:"private_key"   db:"private_key"`
	CAChain     string    `json:"ca_chain"      db:"ca_chain"`
	IssuingCA   string    `json:"issuing_ca"    db:"issuing_ca"`
	TTL         string    `json:"ttl"           db:"ttl"`
	Expire      time.Time `json:"expire"        db:"expire"`
	Revocation  time.Time `json:"revocation"    db:"revocation"`
}

func (cs *certsService) IssueCert(ctx context.Context, token, name string, thingID string, ttl string) (Cert, error) {
	owner, err := cs.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Cert{}, err
	}

	thing, err := cs.sdk.Thing(thingID, token)
	if err != nil {
		return Cert{}, errors.Wrap(ErrThingRetrieve, err)
	}

	id, err := cs.idProvider.ID()
	if err != nil {
		return Cert{}, err
	}

	cert, err := cs.pki.IssueCert(thing.Key, ttl)
	if err != nil {
		return Cert{}, errors.Wrap(ErrPKIIssue, err)
	}

	c := Cert{
		ID:          id,
		Name:        name,
		ThingID:     thingID,
		OwnerID:     owner.GetId(),
		Certificate: cert.Certificate,
		IssuingCA:   cert.IssuingCA,
		CAChain:     strings.Join(cert.CAChain, caChainJoinSep),
		PrivateKey:  cert.PrivateKey,
		Serial:      cert.Serial,
		TTL:         ttl,
		Expire:      cert.Expire,
	}

	err = cs.repo.Save(context.Background(), c)
	if err != nil {
		return Cert{}, err
	}
	return c, nil
}

func (cs *certsService) ListCerts(ctx context.Context, token, certID, thingID, name, serial string, offset, limit uint64) (Page, error) {
	p, _, err := cs.identifyAndRetrieve(ctx, token, certID, thingID, serial, name, offset, int64(limit))
	return p, err
}

func (cs *certsService) ViewCert(ctx context.Context, token, certID string) (Cert, error) {
	cp, u, err := cs.identifyAndRetrieve(ctx, token, certID, "", "", "", 0, 1)
	if err != nil {
		return Cert{}, err
	}
	if len(cp.Certs) < 1 {
		return Cert{}, errors.ErrNotFound
	}

	cert := cp.Certs[0]
	if time.Until(cert.Expire) < time.Duration(1*time.Hour) {
		cert, err = cs.renewAndUpdate(ctx, u.GetId(), cert)
		if err != nil {
			return Cert{}, err
		}
	}
	return cert, nil
}

func (cs *certsService) RenewCert(ctx context.Context, token, certID string) (Cert, error) {
	cp, u, err := cs.identifyAndRetrieve(ctx, token, certID, "", "", "", 0, 1)
	if err != nil {
		return Cert{}, err
	}
	if len(cp.Certs) < 1 {
		return Cert{}, errors.ErrNotFound
	}
	// ToDo don't renew before revoke , To check revoke is zero logic should be  time.Now().Sub(revokeTime) != time.Now()
	return cs.renewAndUpdate(ctx, u.GetId(), cp.Certs[0])
}

func (cs *certsService) RevokeCert(ctx context.Context, token, certID string) error {
	cp, u, err := cs.identifyAndRetrieve(ctx, token, certID, "", "", "", 0, 1)
	if err != nil {
		return err
	}
	if len(cp.Certs) < 1 {
		return errors.ErrNotFound
	}

	return cs.revokeAndUpdate(ctx, u.GetId(), cp.Certs[0])
}

func (cs *certsService) RemoveCert(ctx context.Context, token, certID string) error {
	cp, u, err := cs.identifyAndRetrieve(ctx, token, certID, "", "", "", 0, 1)
	if err != nil {
		return err
	}
	if len(cp.Certs) < 1 {
		return nil
	}

	return cs.revokeAndRemove(ctx, u.GetId(), cp.Certs[0])
}

func (cs *certsService) RenewThingCerts(ctx context.Context, token, thingID string, limit int64) error {
	cp, u, err := cs.identifyAndRetrieve(ctx, token, "", thingID, "", "", 0, limit)
	if err != nil {
		return err
	}
	if len(cp.Certs) < 1 {
		return errors.ErrNotFound
	}

	for _, cert := range cp.Certs {
		// ToDo don't renew before revoke , To check revoke is zero logic should be  time.Now().Sub(revokeTime) != time.Now()
		_, err := cs.renewAndUpdate(ctx, u.GetId(), cert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cs *certsService) RevokeThingCerts(ctx context.Context, token, thingID string, limit int64) error {
	cp, u, err := cs.identifyAndRetrieve(ctx, token, "", thingID, "", "", 0, limit)
	if err != nil {
		return err
	}
	if len(cp.Certs) < 1 {
		return errors.ErrNotFound
	}

	for _, cert := range cp.Certs {
		err := cs.revokeAndUpdate(ctx, u.GetId(), cert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cs *certsService) RemoveThingCerts(ctx context.Context, token, thingID string, limit int64) error {
	cp, u, err := cs.identifyAndRetrieve(ctx, token, "", thingID, "", "", 0, limit)
	if err != nil {
		return err
	}
	if len(cp.Certs) < 1 {
		return nil
	}

	for _, cert := range cp.Certs {
		err := cs.revokeAndRemove(ctx, u.GetId(), cert)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cs *certsService) identifyAndRetrieve(ctx context.Context, token, certID, thingID, serial, name string, offset uint64, limit int64) (Page, *mainflux.UserIdentity, error) {
	u, err := cs.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Page{}, u, errors.Wrap(errors.ErrAuthentication, err)
	}
	cp, err := cs.repo.Retrieve(ctx, u.GetId(), certID, thingID, serial, name, offset, limit)

	if err != nil {
		return Page{}, u, errors.Wrap(errRepoRetrieve, err)
	}
	return cp, u, nil
}

func (cs *certsService) renewAndUpdate(ctx context.Context, ownerID string, cert Cert) (Cert, error) {
	xCert, err := parseCert(cert.Certificate)
	if err != nil {
		return Cert{}, errors.Wrap(errParseCert, err)
	}
	pkiCert, err := cs.pki.IssueCert(xCert.Subject.CommonName, cert.TTL)
	if err != nil {
		return Cert{}, errors.Wrap(ErrPKIIssue, err)
	}

	cert.CAChain = strings.Join(pkiCert.CAChain, caChainJoinSep)
	cert.Certificate = pkiCert.Certificate
	cert.Expire = pkiCert.Expire
	cert.IssuingCA = pkiCert.IssuingCA
	cert.PrivateKey = pkiCert.PrivateKey
	cert.Serial = pkiCert.Serial
	cert.Revocation = time.Time{}

	if err = cs.repo.Update(context.Background(), ownerID, cert); err != nil {
		return Cert{}, errors.Wrap(errRepoUpdate, err)
	}
	return cert, nil
}

func (cs *certsService) revokeAndUpdate(ctx context.Context, ownerID string, c Cert) error {
	if time.Until(c.Revocation) < 0 {
		revTime, err := cs.pki.Revoke(c.Serial)
		if err != nil {
			return errors.Wrap(errPKIRevoke, err)
		}

		c.Revocation = revTime
		if err = cs.repo.Update(context.Background(), ownerID, c); err != nil {
			return errors.Wrap(errRepoUpdate, err)
		}
	}

	return nil
}

func (cs *certsService) revokeAndRemove(ctx context.Context, ownerID string, c Cert) error {
	if time.Until(c.Revocation) < 0 {
		revTime, err := cs.pki.Revoke(c.Serial)
		if err != nil {
			return errors.Wrap(errPKIRevoke, err)
		}
		c.Revocation = revTime
	}

	if err := cs.repo.Remove(context.Background(), ownerID, c.ID); err != nil {
		return errors.Wrap(errRepoRemove, err)
	}
	return nil
}

func parseCert(certificate string) (*x509.Certificate, error) {
	cert, err := x509.ParseCertificate([]byte(certificate))
	if err != nil {
		return nil, err
	}
	return cert, nil
}

func (teh *thingsEventHandlers) ThingCreated(ctx context.Context, cte things.CreateThingEvent) error {
	return nil
}
func (teh *thingsEventHandlers) ThingUpdated(ctx context.Context, ute things.UpdateThingEvent) error {
	return nil
}
func (teh *thingsEventHandlers) ThingRemoved(ctx context.Context, rte things.RemoveThingEvent) error {
	cp, err := teh.repo.RetrieveThingCerts(ctx, rte.ID)
	if err != nil {
		return err
	}
	if len(cp.Certs) < 1 {
		return nil
	}

	// create async thing event handler with go routine and return error via channels
	var retErr error
	for _, cert := range cp.Certs {
		_, err := teh.pki.Revoke(cert.Serial)
		if err != nil {
			retErr = errors.Wrap(retErr, err)
		}
	}
	err = teh.repo.RemoveThingCerts(ctx, rte.ID)
	if err != nil {
		retErr = errors.Wrap(retErr, err)
	}
	return retErr
}
func (teh *thingsEventHandlers) ChannelCreated(ctx context.Context, cce things.CreateChannelEvent) error {
	return nil
}
func (teh *thingsEventHandlers) ChannelUpdated(ctx context.Context, uce things.UpdateChannelEvent) error {
	return nil
}
func (teh *thingsEventHandlers) ChannelRemoved(ctx context.Context, rce things.RemoveChannelEvent) error {
	return nil
}
func (teh *thingsEventHandlers) ThingConnected(ctx context.Context, cte things.ConnectThingEvent) error {
	return nil
}
func (teh *thingsEventHandlers) ThingDisconnected(ctx context.Context, dte things.DisconnectThingEvent) error {
	return nil
}
