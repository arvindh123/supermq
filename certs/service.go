// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"strings"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/certs/pki"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
)

var (
	// ErrFailedCertCreation failed to create certificate
	ErrFailedCertCreation = errors.New("failed to create client certificate")

	// ErrFailedCertRevocation failed to revoke certificate
	ErrFailedCertRevocation = errors.New("failed to revoke certificate in PKI")

	errFailedToUpdateCertRenew = errors.New("failed to update certificate while renewing")

	errFailedToUpdateCertBSRenew = errors.New("failed to update certificate in bootstrap while renewing")

	errFailedToRemoveCertFromDB = errors.New("failed to remove cert serial from db")

	errFailedToRetriveByThingID = errors.New("failed to retrive by thing ID ")

	errFailedToGetThingService = errors.New("failed to get from thing service")

	errFailedToReadFromPKI = errors.New("failed to read cert cert from PKI")
)

var _ Service = (*certsService)(nil)

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// IssueCert issues certificate for given thing id if access is granted with token
	IssueCert(ctx context.Context, token, thingID, ttl string, keyBits int, keyType string) (Cert, error)

	// ListCerts lists certificates issued for a given thing ID
	ListCerts(ctx context.Context, token, thingID string, offset, limit uint64) (Page, error)

	// ListSerials lists certificate serial IDs issued for a given thing ID
	ListSerials(ctx context.Context, token, thingID string, offset, limit uint64) (Page, error)

	// ViewCert retrieves the certificate issued for a given serial ID
	ViewCert(ctx context.Context, token, serialID string) (Cert, error)

	// RevokeCert revokes a certificate for a given serial ID
	RevokeCert(ctx context.Context, token, serialID string) (Revoke, error)

	// Renew the exipred certificate from certs repo
	RenewCerts(ctx context.Context, bsUpdateRenewCert bool) error

	// Automatically trigger RenewCert function for given renew interval time
	AutoRenew(ctx context.Context, bsUpdateRenewCert bool, renewInterval time.Duration) error

	// Automatically trigger revokes certificate of the given thing ID without authentication , used for Auto Renewal of certificate.
	AutoRevokeCerts(ctx context.Context, thingID string) (Revoke, error)
}

// Config defines the service parameters
type Config struct {
	LogLevel            string
	ClientTLS           bool
	CaCerts             string
	HTTPPort            string
	ServerCert          string
	ServerKey           string
	CertsURL            string
	JaegerURL           string
	AuthURL             string
	AuthTimeout         time.Duration
	SignTLSCert         tls.Certificate
	SignX509Cert        *x509.Certificate
	SignRSABits         int
	SignHoursValid      string
	PKIHost             string
	PKIPath             string
	PKIRole             string
	PKIToken            string
	NumOfRenewInOneScan uint64
	RenewBeforeExpiry   time.Duration
	ExpireCheckInterval time.Duration
}

type certsService struct {
	auth      mainflux.AuthServiceClient
	bsClient  BootstrapClient
	certsRepo Repository
	sdk       mfsdk.SDK
	conf      Config
	pki       pki.Agent
}

// New returns new Certs service.
func New(auth mainflux.AuthServiceClient, certs Repository, sdk mfsdk.SDK, bsClient BootstrapClient, config Config, pki pki.Agent) Service {
	return &certsService{
		certsRepo: certs,
		sdk:       sdk,
		bsClient:  bsClient,
		auth:      auth,
		conf:      config,
		pki:       pki,
	}
}

// Revoke defines the conditions to revoke a certificate
type Revoke struct {
	RevocationTime time.Time `mapstructure:"revocation_time"`
}

// Cert defines the certificate paremeters
type Cert struct {
	OwnerID        string    `json:"owner_id" mapstructure:"owner_id"`
	ThingID        string    `json:"thing_id" mapstructure:"thing_id"`
	ClientCert     string    `json:"client_cert" mapstructure:"certificate"`
	IssuingCA      string    `json:"issuing_ca" mapstructure:"issuing_ca"`
	CAChain        []string  `json:"ca_chain" mapstructure:"ca_chain"`
	ClientKey      string    `json:"client_key" mapstructure:"private_key"`
	PrivateKeyType string    `json:"private_key_type" mapstructure:"private_key_type"`
	Serial         string    `json:"serial" mapstructure:"serial_number"`
	Expire         time.Time `json:"expire" mapstructure:"-"`
}

func (cs *certsService) IssueCert(ctx context.Context, token, thingID string, ttl string, keyBits int, keyType string) (Cert, error) {
	owner, err := cs.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Cert{}, err
	}

	thing, err := cs.sdk.Thing(thingID, token)
	if err != nil {
		return Cert{}, errors.Wrap(ErrFailedCertCreation, err)
	}

	cert, err := cs.pki.IssueCert(thing.Key, ttl, keyType, keyBits)
	if err != nil {
		return Cert{}, errors.Wrap(ErrFailedCertCreation, err)
	}

	c := Cert{
		ThingID:        thingID,
		OwnerID:        owner.GetId(),
		ClientCert:     cert.ClientCert,
		IssuingCA:      cert.IssuingCA,
		CAChain:        cert.CAChain,
		ClientKey:      cert.ClientKey,
		PrivateKeyType: cert.PrivateKeyType,
		Serial:         cert.Serial,
		Expire:         cert.Expire,
	}

	_, err = cs.certsRepo.Save(context.Background(), c)
	return c, err
}

func (cs *certsService) RevokeCert(ctx context.Context, token, thingID string) (Revoke, error) {
	var revoke Revoke
	u, err := cs.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return revoke, err
	}
	thing, err := cs.sdk.Thing(thingID, token)
	if err != nil {
		return revoke, errors.Wrap(ErrFailedCertRevocation, err)
	}

	// TODO: Replace offset and limit
	offset, limit := uint64(0), uint64(10000)
	cp, err := cs.certsRepo.RetrieveByThing(ctx, u.GetId(), thing.ID, offset, limit)
	if err != nil {
		return revoke, errors.Wrap(ErrFailedCertRevocation, err)
	}

	for _, c := range cp.Certs {
		revTime, err := cs.pki.Revoke(c.Serial)
		if err != nil {
			return revoke, errors.Wrap(ErrFailedCertRevocation, err)
		}
		revoke.RevocationTime = revTime
		if err = cs.certsRepo.Remove(context.Background(), u.GetId(), c.Serial); err != nil {
			return revoke, errors.Wrap(errFailedToRemoveCertFromDB, err)
		}
	}

	return revoke, nil
}

func (cs *certsService) ListCerts(ctx context.Context, token, thingID string, offset, limit uint64) (Page, error) {
	u, err := cs.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Page{}, err
	}

	cp, err := cs.certsRepo.RetrieveByThing(ctx, u.GetId(), thingID, offset, limit)
	if err != nil {
		return Page{}, err
	}

	for i, cert := range cp.Certs {
		vcert, err := cs.pki.Read(cert.Serial)
		if err != nil {
			return Page{}, err
		}
		cp.Certs[i].ClientCert = vcert.ClientCert
		cp.Certs[i].ClientKey = vcert.ClientKey
	}

	return cp, nil
}

func (cs *certsService) ListSerials(ctx context.Context, token, thingID string, offset, limit uint64) (Page, error) {
	u, err := cs.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Page{}, err
	}

	return cs.certsRepo.RetrieveByThing(ctx, u.GetId(), thingID, offset, limit)
}

func (cs *certsService) ViewCert(ctx context.Context, token, serialID string) (Cert, error) {
	u, err := cs.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Cert{}, err
	}

	cert, err := cs.certsRepo.RetrieveBySerial(ctx, u.GetId(), serialID)
	if err != nil {
		return Cert{}, err
	}

	vcert, err := cs.pki.Read(serialID)
	if err != nil {
		return Cert{}, err
	}

	c := Cert{
		ThingID:    cert.ThingID,
		ClientCert: vcert.ClientCert,
		Serial:     cert.Serial,
		Expire:     cert.Expire,
	}

	return c, nil
}

func (cs *certsService) RenewCerts(ctx context.Context, bsUpdateRenewCert bool) error {
	cr, err := cs.certsRepo.ListExpiredCerts(ctx, cs.conf.RenewBeforeExpiry, cs.conf.NumOfRenewInOneScan, 0)
	if err != nil {
		return err
	}
	for _, repoExpCert := range cr.Certs {
		cert, err := cs.pki.IssueCert(repoExpCert.ThingID, cs.conf.SignHoursValid, "", cs.conf.SignRSABits)
		if err != nil {
			return errors.Wrap(ErrFailedCertCreation, err)
		}
		c := Cert{
			ThingID:        repoExpCert.ThingID,
			OwnerID:        repoExpCert.OwnerID,
			ClientCert:     cert.ClientCert,
			IssuingCA:      cert.IssuingCA,
			CAChain:        cert.CAChain,
			ClientKey:      cert.ClientKey,
			PrivateKeyType: cert.PrivateKeyType,
			Serial:         cert.Serial,
			Expire:         cert.Expire,
		}
		if bsUpdateRenewCert {
			if err := cs.bsClient.UpdateCerts(ctx, repoExpCert.ThingID, cert.ClientCert, cert.ClientKey, strings.Join(cert.CAChain, "\n")); err != nil {
				return errors.Wrap(errFailedToUpdateCertBSRenew, err)
			}
		}
		if err := cs.certsRepo.Update(ctx, c); err != nil {
			return errors.Wrap(errFailedToUpdateCertRenew, err)
		}
	}
	return nil
}

func (cs *certsService) AutoRenew(ctx context.Context, bsUpdateRenewCert bool, renewInterval time.Duration) error {
	ticker := time.NewTicker(renewInterval)
	rCtx, rCancel := context.WithCancel(context.Background())

	renewErrChan := make(chan error, 1)
	renewCertsFn := func(ctx context.Context, bsUpdateRenewCert bool, err chan error, renewCert func(ctx context.Context, bsUpdateRenewCert bool) error) {
		err <- renewCert(ctx, bsUpdateRenewCert)
	}

	go renewCertsFn(rCtx, bsUpdateRenewCert, renewErrChan, cs.RenewCerts)

	for {
		select {
		case <-ctx.Done():
			rCancel()
			return nil
		case renewErr := <-renewErrChan:
			rCancel()
			return renewErr
		case <-ticker.C:
			go renewCertsFn(rCtx, bsUpdateRenewCert, renewErrChan, cs.RenewCerts)
		}
	}
}

func (cs *certsService) AutoRevokeCerts(ctx context.Context, thingID string) (Revoke, error) {
	var revoke Revoke

	c, err := cs.certsRepo.AutoRetrieveByThingID(ctx, thingID)
	if err != nil {
		return revoke, errors.Wrap(errFailedToRetriveByThingID, err)
	}

	revTime, err := cs.pki.Revoke(c.Serial)
	if err != nil {
		return revoke, errors.Wrap(ErrFailedCertRevocation, err)
	}
	revoke.RevocationTime = revTime
	if err = cs.certsRepo.AutoRemoveByThingID(context.Background(), c.Serial); err != nil {
		return revoke, errors.Wrap(errFailedToRemoveCertFromDB, err)
	}
	return revoke, nil
}
