// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package bolt

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"
	"time"

	"github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"

	bolt "go.etcd.io/bbolt"
)

const (
	idKey               = "id"
	userKey             = "user"
	nameKey             = "name"
	descriptionKey      = "description"
	tokenHashKey        = "token_hash"
	scopeKey            = "scope"
	issuedAtKey         = "issued_at"
	expiresAtKey        = "expires_at"
	updatedAtKey        = "updated_at"
	lastUsedAtKey       = "last_used_at"
	revokedKey          = "revoked"
	revokedAtKey        = "revoked_at"
	platformEntitiesKey = "platform_entities"

	keySeparator = ":"
	anyID        = "*"
)

var (
	revokedValue     = []byte{0x01}
	entityValue      = []byte{0x02}
	anyIDValue       = []byte{0x03}
	selectedIDsValue = []byte{0x04}

	errBucketNotFound = errors.New("bucket not found")
)

type patRepo struct {
	db         *bolt.DB
	bucketName string
}

// NewPATSRepository instantiates a bolt
// implementation of PAT repository.
func NewPATSRepository(db *bolt.DB, bucketName string) auth.PATSRepository {
	return &patRepo{
		db:         db,
		bucketName: bucketName,
	}
}

func (pr *patRepo) Save(ctx context.Context, pat auth.PAT) error {
	kv, err := patToKeyValue(pat)
	if err != nil {
		return err
	}
	return pr.db.Update(func(tx *bolt.Tx) error {
		b, err := pr.createRetrieveUserBucket(tx, pat.User)
		if err != nil {
			return errors.Wrap(repoerr.ErrCreateEntity, err)
		}
		for key, value := range kv {
			fullKey := []byte(pat.ID + keySeparator + key)
			if err := b.Put(fullKey, value); err != nil {
				return errors.Wrap(repoerr.ErrCreateEntity, err)
			}
		}
		return nil
	})
}

func (pr *patRepo) Retrieve(ctx context.Context, userID, patID string) (auth.PAT, error) {
	var pat auth.PAT
	prefix := []byte(patID + keySeparator)
	kv := map[string][]byte{}
	if err := pr.db.Update(func(tx *bolt.Tx) error {
		b, err := pr.retrieveUserBucket(tx, userID)
		if err != nil {
			return errors.Wrap(repoerr.ErrRemoveEntity, err)
		}
		c := b.Cursor()
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			kv[string(k)] = v
		}
		return nil
	}); err != nil {
		return auth.PAT{}, err
	}

	for k, v := range kv {
		fmt.Println(k, string(v))
	}
	fmt.Println(kv)
	return pat, nil
}

func (pr *patRepo) UpdateName(ctx context.Context, userID, patID, name string) (auth.PAT, error) {
	return pr.updatePATField(ctx, userID, patID, nameKey, []byte(name))
}

func (pr *patRepo) UpdateDescription(ctx context.Context, userID, patID, description string) (auth.PAT, error) {
	return pr.updatePATField(ctx, userID, patID, descriptionKey, []byte(description))
}

func (pr *patRepo) UpdateTokenHash(ctx context.Context, userID, patID, tokenHash string, expiryAt time.Time) (auth.PAT, error) {
	var pat auth.PAT
	if err := pr.db.Update(func(tx *bolt.Tx) error {
		b, err := pr.retrieveUserBucket(tx, userID)
		if err != nil {
			return errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
		if err := b.Put([]byte(patID+keySeparator+tokenHashKey), []byte(tokenHash)); err != nil {
			return errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
		if err := b.Put([]byte(patID+keySeparator+expiresAtKey), timeToBytes(expiryAt)); err != nil {
			return errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
		pat, err = pr.retrievePAT(b, patID)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return auth.PAT{}, err
	}
	return pat, nil
}

func (pr *patRepo) RetrieveAll(ctx context.Context, userID string) (pats auth.PATSPage, err error) {
	return auth.PATSPage{}, nil
}

func (pr *patRepo) Revoke(ctx context.Context, userID, patID string) error {
	if err := pr.db.Update(func(tx *bolt.Tx) error {
		b, err := pr.retrieveUserBucket(tx, userID)
		if err != nil {
			return errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
		if err := b.Put([]byte(patID+keySeparator+revokedKey), revokedValue); err != nil {
			return errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
		if err := b.Put([]byte(patID+keySeparator+revokedAtKey), timeToBytes(time.Now())); err != nil {
			return errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (pr *patRepo) Remove(ctx context.Context, userID, patID string) error {
	prefix := []byte(patID + keySeparator)

	if err := pr.db.Update(func(tx *bolt.Tx) error {
		b, err := pr.retrieveUserBucket(tx, userID)
		if err != nil {
			return errors.Wrap(repoerr.ErrRemoveEntity, err)
		}
		c := b.Cursor()
		for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
			if err := b.Delete(k); err != nil {
				return errors.Wrap(repoerr.ErrRemoveEntity, err)
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (pr *patRepo) AddScopeEntry(ctx context.Context, userID, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (auth.Scope, error) {
	if err := pr.db.Update(func(tx *bolt.Tx) error {
		b, err := pr.retrieveUserBucket(tx, userID)
		if err != nil {
			return errors.Wrap(repoerr.ErrCreateEntity, err)
		}
		kv, err := scopeEntryToKeyValue(platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
		if err != nil {
			return err
		}
		for key, value := range kv {
			fullKey := []byte(patID + keySeparator + key)
			if err := b.Put(fullKey, value); err != nil {
				return errors.Wrap(repoerr.ErrCreateEntity, err)
			}
		}
		return nil
	}); err != nil {
		return auth.Scope{}, err
	}

	return auth.Scope{}, nil
}

func (pr *patRepo) RemoveScopeEntry(ctx context.Context, userID, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (auth.Scope, error) {
	if len(entityIDs) == 0 {
		return auth.Scope{}, repoerr.ErrMalformedEntity
	}
	if err := pr.db.Update(func(tx *bolt.Tx) error {
		b, err := pr.retrieveUserBucket(tx, userID)
		if err != nil {
			return errors.Wrap(repoerr.ErrRemoveEntity, err)
		}
		kv, err := scopeEntryToKeyValue(platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
		if err != nil {
			return err
		}
		for key, _ := range kv {
			fullKey := []byte(patID + keySeparator + key)
			if err := b.Delete(fullKey); err != nil {
				return errors.Wrap(repoerr.ErrRemoveEntity, err)
			}
		}
		return nil
	}); err != nil {
		return auth.Scope{}, err
	}
	return auth.Scope{}, nil
}

func (pr *patRepo) CheckScopeEntry(ctx context.Context, userID, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) error {
	return pr.db.Update(func(tx *bolt.Tx) error {
		b, err := pr.retrieveUserBucket(tx, userID)
		if err != nil {
			return errors.Wrap(repoerr.ErrViewEntity, err)
		}
		srootKey, err := scopeRootKey(platformEntityType, optionalDomainID, optionalDomainEntityType, operation)
		if err != nil {
			return errors.Wrap(repoerr.ErrViewEntity, err)
		}

		rootKey := patID + keySeparator + srootKey
		if value := b.Get([]byte(rootKey)); bytes.Equal(value, anyIDValue) {
			return nil
		}
		for _, entity := range entityIDs {
			value := b.Get([]byte(rootKey + keySeparator + entity))
			if !bytes.Equal(value, entityValue) {
				return repoerr.ErrNotFound
			}
		}
		return nil
	})
}

func (pr *patRepo) RemoveAllScopeEntry(ctx context.Context, userID, patID string) error {
	return nil
}

func (pr *patRepo) createRetrieveUserBucket(tx *bolt.Tx, userID string) (*bolt.Bucket, error) {
	rootBucket := tx.Bucket([]byte(pr.bucketName))
	if rootBucket == nil {
		return nil, errors.Wrap(repoerr.ErrCreateEntity, fmt.Errorf("bucket %s not found", pr.bucketName))
	}

	userBucket, err := rootBucket.CreateBucketIfNotExists([]byte(userID))
	if err != nil {
		return nil, errors.Wrap(repoerr.ErrCreateEntity, fmt.Errorf("failed to retrieve or create bucket for user %s : %w", userID, err))
	}

	return userBucket, nil
}

func (pr *patRepo) retrieveUserBucket(tx *bolt.Tx, userID string) (*bolt.Bucket, error) {
	rootBucket := tx.Bucket([]byte(pr.bucketName))
	if rootBucket == nil {
		return nil, fmt.Errorf("bucket %s not found", pr.bucketName)
	}

	return rootBucket.Bucket([]byte(userID)), nil
}

func (pr *patRepo) updatePATField(_ context.Context, userID, patID, key string, value []byte) (auth.PAT, error) {
	var pat auth.PAT
	prefix := []byte(patID + keySeparator)
	kv := map[string][]byte{}
	if err := pr.db.Update(func(tx *bolt.Tx) error {
		b, err := pr.retrieveUserBucket(tx, userID)
		if err != nil {
			return errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
		if err := b.Put([]byte(patID+keySeparator+key), value); err != nil {
			return errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
		c := b.Cursor()
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			kv[string(k)] = v
		}
		return nil
	}); err != nil {
		return auth.PAT{}, err
	}
	return pat, nil
}

func (pr *patRepo) scopeKeyBuilder(patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType) (string, error) {
	op, err := operation.ValidString()
	if err != nil {
		return "", errors.Wrap(repoerr.ErrMalformedEntity, err)
	}
	switch platformEntityType {
	case auth.PlatformUsersScope:
		return patID + keySeparator + scopeKey + keySeparator + platformEntityType.String() + keySeparator + op, nil
	case auth.PlatformDomainsScope:
		if optionalDomainID == "" {
			return "", fmt.Errorf("failed to add platform %s scope: invalid domain id", platformEntityType.String())
		}
		odet, err := optionalDomainEntityType.ValidString()
		if err != nil {
			return "", errors.Wrap(repoerr.ErrMalformedEntity, err)
		}

		return patID + keySeparator + scopeKey + keySeparator + platformEntityType.String() + keySeparator + optionalDomainID + keySeparator + odet + keySeparator + op + keySeparator, nil
	default:
		return "", errors.Wrap(repoerr.ErrMalformedEntity, fmt.Errorf("invalid platform entity type %s", platformEntityType.String()))
	}
}

func (pr *patRepo) updateLastUsed(ctx context.Context, userID, patID string) error {
	return nil
}

func (pr *patRepo) retrievePAT(b *bolt.Bucket, patID string) (auth.PAT, error) {
	var pat auth.PAT

	id := b.Get([]byte(patID + ":id"))
	if id == nil {
		return pat, repoerr.ErrNotFound
	}
	pat.ID = string(id)
	pat.User = getString(b, patID, userKey)
	pat.Name = getString(b, patID, nameKey)
	pat.Description = getString(b, patID, descriptionKey)
	pat.IssuedAt = getTime(b, patID, issuedAtKey)
	pat.ExpiresAt = getTime(b, patID, expiresAtKey)
	pat.UpdatedAt = getTime(b, patID, updatedAtKey)
	pat.LastUsedAt = getTime(b, patID, lastUsedAtKey)
	pat.Revoked = getBool(b, patID, revokedKey)
	pat.RevokedAt = getTime(b, patID, revokedAtKey)
	return pat, nil
}

func (pr *patRepo) retrieveScope(b *bolt.Bucket, patID string) (auth.Scope, error) {
	var scope auth.Scope

	c := b.Cursor()
	prefix := []byte(patID + keySeparator + scopeKey + keySeparator)
	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		fmt.Printf("key=%s, value=%v\n", k, v)
		keyWithoutPrefix := k[len(prefix):]
		fmt.Printf("key=%s, value=%v\n", keyWithoutPrefix, v)
	}
	return scope, nil
}

func patToKeyValue(pat auth.PAT) (map[string][]byte, error) {
	kv := map[string][]byte{
		idKey:          []byte(pat.ID),
		userKey:        []byte(pat.User),
		nameKey:        []byte(pat.Name),
		descriptionKey: []byte(pat.Description),
		tokenHashKey:   []byte(pat.Token),
		issuedAtKey:    timeToBytes(pat.IssuedAt),
		expiresAtKey:   timeToBytes(pat.ExpiresAt),
		updatedAtKey:   timeToBytes(pat.UpdatedAt),
		lastUsedAtKey:  timeToBytes(pat.LastUsedAt),
		revokedKey:     booleanToBytes(pat.Revoked),
		revokedAtKey:   timeToBytes(pat.RevokedAt),
	}
	scopeKV, err := scopeToKeyValue(pat.Scope)
	if err != nil {
		return nil, err
	}
	for k, v := range scopeKV {
		kv[k] = v
	}
	return kv, nil
}

func scopeToKeyValue(scope auth.Scope) (map[string][]byte, error) {
	kv := map[string][]byte{}
	for opType, scopeValue := range scope.Users.Operations {
		tempKV, err := scopeEntryToKeyValue(auth.PlatformUsersScope, "", auth.DomainNullScope, opType, scopeValue.Values()...)
		if err != nil {
			return nil, err
		}
		for k, v := range tempKV {
			kv[k] = v
		}
	}
	for domainID, domainScope := range scope.Domains {
		for opType, scopeValue := range domainScope.DomainManagement.Operations {
			tempKV, err := scopeEntryToKeyValue(auth.PlatformDomainsScope, domainID, auth.DomainManagementScope, opType, scopeValue.Values()...)
			if err != nil {
				return nil, errors.Wrap(repoerr.ErrCreateEntity, err)
			}
			for k, v := range tempKV {
				kv[k] = v
			}
		}
		for entityType, scope := range domainScope.Entities {
			for opType, scopeValue := range scope.Operations {
				tempKV, err := scopeEntryToKeyValue(auth.PlatformDomainsScope, domainID, entityType, opType, scopeValue.Values()...)
				if err != nil {
					return nil, errors.Wrap(repoerr.ErrCreateEntity, err)
				}
				for k, v := range tempKV {
					kv[k] = v
				}
			}
		}
	}
	return kv, nil
}
func scopeEntryToKeyValue(platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (map[string][]byte, error) {
	if len(entityIDs) == 0 {
		return nil, repoerr.ErrMalformedEntity
	}

	rootKey, err := scopeRootKey(platformEntityType, optionalDomainID, optionalDomainEntityType, operation)
	if err != nil {
		return nil, err
	}
	if len(entityIDs) == 1 && entityIDs[0] == anyID {
		return map[string][]byte{rootKey: anyIDValue}, nil
	}

	kv := map[string][]byte{rootKey: selectedIDsValue}

	for _, entryID := range entityIDs {
		if entryID == anyID {
			return nil, repoerr.ErrMalformedEntity
		}
		kv[rootKey+keySeparator+entryID] = entityValue
	}

	return kv, nil
}

func scopeRootKey(platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType) (string, error) {

	op, err := operation.ValidString()
	if err != nil {
		return "", errors.Wrap(repoerr.ErrMalformedEntity, err)
	}

	var rootKey strings.Builder

	rootKey.WriteString(scopeKey)
	rootKey.WriteString(keySeparator)
	rootKey.WriteString(platformEntityType.String())
	rootKey.WriteString(keySeparator)

	switch platformEntityType {
	case auth.PlatformUsersScope:
		rootKey.WriteString(op)
	case auth.PlatformDomainsScope:
		if optionalDomainID == "" {
			return "", fmt.Errorf("failed to add platform %s scope: invalid domain id", platformEntityType.String())
		}
		odet, err := optionalDomainEntityType.ValidString()
		if err != nil {
			return "", errors.Wrap(repoerr.ErrMalformedEntity, err)
		}
		rootKey.WriteString(optionalDomainID)
		rootKey.WriteString(keySeparator)
		rootKey.WriteString(odet)
		rootKey.WriteString(keySeparator)
		rootKey.WriteString(op)
	default:
		return "", errors.Wrap(repoerr.ErrMalformedEntity, fmt.Errorf("invalid platform entity type %s", platformEntityType.String()))
	}

	return rootKey.String(), nil
}
func keyValueToPAT(kv map[string][]byte) (auth.PAT, error) {
	var pat auth.PAT
	for k, v := range kv {
		switch {
		case strings.HasSuffix(k, keySeparator+idKey):
			pat.ID = string(v)
		case strings.HasSuffix(k, keySeparator+userKey):
			pat.User = string(v)
		case strings.HasSuffix(k, keySeparator+nameKey):
			pat.Name = string(v)
		case strings.HasSuffix(k, keySeparator+descriptionKey):
			pat.Description = string(v)
		case strings.HasSuffix(k, keySeparator+issuedAtKey):
			pat.IssuedAt = bytesToTime(v)
		case strings.HasSuffix(k, keySeparator+expiresAtKey):
			pat.ExpiresAt = bytesToTime(v)
		case strings.HasSuffix(k, keySeparator+updatedAtKey):
			pat.UpdatedAt = bytesToTime(v)
		case strings.HasSuffix(k, keySeparator+lastUsedAtKey):
			pat.LastUsedAt = bytesToTime(v)
		case strings.HasSuffix(k, keySeparator+revokedKey):
			pat.Revoked = bytesToBoolean(v)
		case strings.HasSuffix(k, keySeparator+revokedAtKey):
			pat.RevokedAt = bytesToTime(v)
		}
	}
	return pat, nil
}

func parseKeyValueToScope(kv map[string][]byte) (auth.Scope, error) {
	scope := auth.Scope{
		Domains: make(map[string]auth.DomainScope),
	}

	for key, value := range kv {
		if strings.Index(key, keySeparator+scopeKey) > 0 {
			parts := strings.Split(key, keySeparator)
			if len(parts) < 4 {
				return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, fmt.Errorf("invalid key format: %s", key))
			}

			platformEntityType, err := auth.ParsePlatformEntityType(parts[2])
			if err != nil {
				return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
			}
			opType, err := auth.ParseOperationType(parts[len(parts)-1])
			if err != nil {
				return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
			}

			switch platformEntityType {
			case auth.PlatformUsersScope:
				if len(parts) != 4 {
					return auth.Scope{}, fmt.Errorf("invalid user scope key format: %s", key)
				}
				if scope.Users.Operations == nil {
					scope.Users.Operations = make(map[auth.OperationType]auth.ScopeValue)
				}
				scope.Users.Operations[opType] = auth.ScopeValue{values: parseValues(value)}
			case auth.PlatformDomainsScope:
				if len(parts) < 5 {
					return auth.Scope{}, fmt.Errorf("invalid domain scope key format: %s", key)
				}
				domainID := parts[2]
				entityType := auth.DomainEntityType(parts[3])
				if scope.Domains[domainID].Entities == nil {
					scope.Domains[domainID] = auth.DomainScope{
						Entities: make(map[auth.DomainEntityType]auth.EntityScope),
					}
				}
				if scope.Domains[domainID].Entities[entityType].Operations == nil {
					scope.Domains[domainID].Entities[entityType] = auth.EntityScope{
						Operations: make(map[auth.OperationType]auth.ScopeValue),
					}
				}
				scope.Domains[domainID].Entities[entityType].Operations[opType] = auth.ScopeValue{values: parseValues(value)}
			default:
				return auth.Scope{}, fmt.Errorf("invalid platform entity type: %s", parts[2])
			}
		}

	}

	return scope, nil
}

func getString(b *bolt.Bucket, patID, key string) string {
	value := b.Get([]byte(patID + keySeparator + key))
	if value != nil {
		return string(value)
	}
	return ""
}

func getTime(b *bolt.Bucket, patID, key string) time.Time {
	value := b.Get([]byte(patID + keySeparator + key))
	if value != nil {
		return bytesToTime(value)
	}
	return time.Time{}
}

func getBool(b *bolt.Bucket, patID, key string) bool {
	value := b.Get([]byte(patID + keySeparator + key))
	if value != nil {
		return bytesToBoolean(value)
	}
	return false
}

func timeToBytes(t time.Time) []byte {
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(t.Nanosecond()))
	return timeBytes
}

func bytesToTime(b []byte) time.Time {
	var timeAtNs uint64
	binary.BigEndian.AppendUint64(b, timeAtNs)
	return time.Unix(0, int64(timeAtNs))
}

func booleanToBytes(b bool) []byte {
	if b {
		return []byte{1}
	}
	return []byte{0}
}

func bytesToBoolean(b []byte) bool {
	if len(b) > 0 && b[0] == 1 {
		return true
	}
	return false
}
