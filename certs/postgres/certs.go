// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // required for SQL access
	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/internal/sqlxt"
	"github.com/mainflux/mainflux/pkg/errors"
)

var _ certs.Repository = (*certsRepository)(nil)

// Cert holds info on expiration date for specific cert issued for specific Thing.
type Cert struct {
	ThingID string
	Serial  string
	Expire  time.Time
}

type certsRepository struct {
	db sqlxt.Database
}

// NewRepository instantiates a PostgreSQL implementation of certs
// repository.
func NewRepository(db sqlxt.Database) certs.Repository {
	return &certsRepository{
		db: db,
	}
}

func (cr certsRepository) Save(ctx context.Context, cert certs.Cert) error {

	q := `INSERT INTO certs
			(id, name, owner_id, thing_id, serial, private_key, certificate, ca_chain, issuing_ca, ttl, expire)
		VALUES
			(:id, :name, :owner_id, :thing_id, :serial, :private_key, :certificate, :ca_chain, :issuing_ca, :ttl, :expire)
		`
	if _, err, txErr := cr.db.NamedCUDContext(ctx, q, CertToDbCert(cert)); err != nil || txErr != nil {
		wErr := errors.Wrap(errors.ErrCreateEntity, err)
		return errors.Wrap(wErr, txErr)
	}
	return nil
}

func (cr certsRepository) Update(ctx context.Context, certID string, cert certs.Cert) error {
	q := `
		UPDATE
			certs
		SET
			serial = :serial,
			private_key = :private_key,
			certificate = :certificate,
			ca_chain = :ca_chain,
			issuing_ca = :issuing_ca,
			expire = :expire
			revocation = :revocation
		WHERE id = :id AND owner_id = :owner_id
	`
	if _, err, txErr := cr.db.NamedCUDContext(ctx, q, CertToDbCert(cert)); err != nil || txErr != nil {
		wErr := errors.Wrap(errors.ErrUpdateEntity, err)
		return errors.Wrap(wErr, txErr)
	}
	return nil
}

func (cr certsRepository) Remove(ctx context.Context, ownerID, certID string) error {
	q := `DELETE FROM certs WHERE id = :id`
	if _, err, txErr := cr.db.NamedCUDContext(ctx, q, CertToDbCert(certs.Cert{ID: certID})); err != nil || txErr != nil {
		wErr := errors.Wrap(errors.ErrUpdateEntity, err)
		return errors.Wrap(wErr, txErr)
	}
	return nil
}

func (cr certsRepository) Retrieve(ctx context.Context, ownerID, certID, thingID, serial, name string, offset uint64, limit int64) (certs.Page, error) {
	q := `
	SELECT
		id, name, owner_id, thing_id, serial, private_key, certificate, ca_chain, issuing_ca, ttl, expire, revocation
	FROM
		certs
	WHERE ownerID = :ownerID
		%s
	ORDER BY expire %s;
	`

	q = fmt.Sprintf(q, whereClause(certID, thingID, serial, name), orderClause(limit))

	params := map[string]interface{}{
		"limit":   limit,
		"offset":  offset,
		"ownerID": ownerID,
		"id":      certID,
		"thingID": thingID,
		"serial":  serial,
		"name":    name,
	}

	rows, err := cr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return certs.Page{}, err
	}
	defer rows.Close()

	certificates := []certs.Cert{}
	for rows.Next() {
		dbc := dbCert{}
		if err := rows.Scan(&dbc); err != nil {
			return certs.Page{}, err
		}
		certificates = append(certificates, dbc.ToCert())
	}

	qc := `
	SELECT
		COUNT(*)
	FROM
		certs
	WHERE ownerID = :ownerID
		%s
	ORDER BY expire %s;
	`
	qc = fmt.Sprintf(qc, whereClause(certID, thingID, serial, name), orderClause(limit))
	total, err := cr.db.NamedTotalQueryContext(ctx, qc, params)
	if err != nil {
		return certs.Page{}, err
	}

	return certs.Page{
		Total:  total,
		Limit:  limit,
		Offset: offset,
		Certs:  certificates,
	}, nil
}

func (cr certsRepository) RetrieveThingCerts(ctx context.Context, thingID string) (certs.Page, error) {
	q := `
	SELECT
		id, name, owner_id, thing_id, serial, private_key, certificate, ca_chain, issuing_ca, ttl, expire, revocation
	FROM
		certs
	WHERE thing_id = :thingID
	ORDER BY expire;
	`

	params := certs.Cert{ThingID: thingID}

	rows, err := cr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return certs.Page{}, err
	}
	defer rows.Close()

	certificates := []certs.Cert{}
	for rows.Next() {
		dbc := dbCert{}
		if err := rows.Scan(&dbc); err != nil {
			return certs.Page{}, err
		}
		certificates = append(certificates, dbc.ToCert())
	}

	qc := `
	SELECT
		COUNT(*)
	FROM
		certs
	WHERE thing_id = :thingID
	ORDER BY expire;
	`
	total, err := cr.db.NamedTotalQueryContext(ctx, qc, params)
	if err != nil {
		return certs.Page{}, err
	}

	return certs.Page{
		Total:  total,
		Limit:  0,
		Offset: 0,
		Certs:  certificates,
	}, nil
}

func (cr certsRepository) RemoveThingCerts(ctx context.Context, thingID string) error {
	q := `DELETE FROM certs WHERE thing_id = thingID`
	if _, err, txErr := cr.db.NamedCUDContext(ctx, q, CertToDbCert(certs.Cert{ThingID: thingID})); err != nil || txErr != nil {
		wErr := errors.Wrap(errors.ErrUpdateEntity, err)
		return errors.Wrap(wErr, txErr)
	}
	return nil
}

type dbCert struct {
	id          string    `db:"id"`
	name        string    `db:"name"`
	ownerID     string    `db:"owner_id"`
	thingID     string    `db:"thing_id"`
	serial      string    `db:"serial"`
	certificate string    `db:"certificate"`
	privateKey  string    `db:"private_key"`
	caChain     string    `db:"ca_chain"`
	issuingCA   string    `db:"issuing_ca"`
	ttl         string    `db:"ttl"`
	expire      time.Time `db:"expire"`
	revocation  time.Time `db:"revocation"`
}

func (c *dbCert) ToCert() certs.Cert {
	return certs.Cert{
		ID:          c.id,
		Name:        c.name,
		OwnerID:     c.ownerID,
		ThingID:     c.thingID,
		Serial:      c.serial,
		Certificate: c.certificate,
		PrivateKey:  c.privateKey,
		CAChain:     c.caChain,
		IssuingCA:   c.issuingCA,
		TTL:         c.ttl,
		Expire:      c.expire,
		Revocation:  c.revocation,
	}
}

func CertToDbCert(c certs.Cert) dbCert {
	return dbCert{
		id:          c.ID,
		name:        c.Name,
		ownerID:     c.OwnerID,
		thingID:     c.ThingID,
		serial:      c.Serial,
		certificate: c.Certificate,
		privateKey:  c.PrivateKey,
		caChain:     c.CAChain,
		issuingCA:   c.IssuingCA,
		ttl:         c.TTL,
		expire:      c.Expire,
		revocation:  c.Revocation,
	}
}

func whereClause(certID, thingID, serial, name string) string {
	var clause []string
	if certID != "" {
		clause = append(clause, " id = :id ")
	}

	if thingID != "" {
		clause = append(clause, " thing_id = :thingID ")
	}

	if serial != "" {
		clause = append(clause, " serial = :serial ")
	}

	if name != "" {
		clause = append(clause, " name = :name ")
	}
	return strings.Join(clause, " AND ")
}

func orderClause(limit int64) string {
	var clause []string
	if limit >= 0 {
		clause = append(clause, " LIMIT :limit ")
	}
	clause = append(clause, " OFFSET = :offset ")
	return strings.Join(clause, "  ")
}
