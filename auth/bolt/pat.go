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
	secretKey           = "secret_key"
	scopeKey            = "scope"
	issuedAtKey         = "issued_at"
	expiresAtKey        = "expires_at"
	updatedAtKey        = "updated_at"
	lastUsedAtKey       = "last_used_at"
	revokedKey          = "revoked"
	revokedAtKey        = "revoked_at"
	platformEntitiesKey = "platform_entities"
	patKey              = "pat"

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
		if err := b.Put([]byte(pat.User+keySeparator+patKey+pat.ID), []byte(pat.ID)); err != nil {
			return errors.Wrap(repoerr.ErrCreateEntity, err)
		}
		return nil
	})
}

func (pr *patRepo) Retrieve(ctx context.Context, userID, patID string) (auth.PAT, error) {
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

	return keyValueToPAT(kv)
}

func (pr *patRepo) UpdateName(ctx context.Context, userID, patID, name string) (auth.PAT, error) {
	return pr.updatePATField(ctx, userID, patID, nameKey, []byte(name))
}

func (pr *patRepo) UpdateDescription(ctx context.Context, userID, patID, description string) (auth.PAT, error) {
	return pr.updatePATField(ctx, userID, patID, descriptionKey, []byte(description))
}

func (pr *patRepo) UpdateTokenHash(ctx context.Context, userID, patID, tokenHash string, expiryAt time.Time) (auth.PAT, error) {
	prefix := []byte(patID + keySeparator)
	kv := map[string][]byte{}
	if err := pr.db.Update(func(tx *bolt.Tx) error {
		b, err := pr.retrieveUserBucket(tx, userID)
		if err != nil {
			return errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
		if err := b.Put([]byte(patID+keySeparator+secretKey), []byte(tokenHash)); err != nil {
			return errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
		if err := b.Put([]byte(patID+keySeparator+expiresAtKey), timeToBytes(expiryAt)); err != nil {
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
	return keyValueToPAT(kv)
}

func (pr *patRepo) RetrieveAll(ctx context.Context, userID string, pm auth.PATSPageMeta) (auth.PATSPage, error) {
	prefix := []byte(userID + keySeparator + patKey + keySeparator)

	patIDs := []string{}
	if err := pr.db.Update(func(tx *bolt.Tx) error {
		b, err := pr.retrieveUserBucket(tx, userID)
		if err != nil {
			return errors.Wrap(repoerr.ErrRemoveEntity, err)
		}
		c := b.Cursor()
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			if v != nil {
				patIDs = append(patIDs, string(v))
			}
		}
		return nil
	}); err != nil {
		return auth.PATSPage{}, err
	}

	total := len(patIDs)

	var pats []auth.PAT

	var patsPage = auth.PATSPage{
		Total:  uint64(total),
		Limit:  pm.Limit,
		Offset: pm.Offset,
		PATS:   pats,
	}

	if int(pm.Offset) >= total {
		return patsPage, nil
	}

	aLimit := pm.Limit
	if rLimit := total - int(pm.Offset); int(pm.Limit) > rLimit {
		aLimit = uint64(rLimit)
	}

	for i := pm.Offset; i < aLimit; i++ {
		if total < int(i) {
			pat, err := pr.Retrieve(ctx, userID, patIDs[i])
			if err != nil {
				return patsPage, err
			}
			patsPage.PATS = append(patsPage.PATS, pat)
		}
	}

	return patsPage, nil
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
	prefix := []byte(patID + keySeparator + scopeKey)
	var rKV map[string][]byte
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
		c := b.Cursor()
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			rKV[string(k)] = v
		}
		return nil
	}); err != nil {
		return auth.Scope{}, err
	}

	return parseKeyValueToScope(rKV)
}

func (pr *patRepo) RemoveScopeEntry(ctx context.Context, userID, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (auth.Scope, error) {
	if len(entityIDs) == 0 {
		return auth.Scope{}, repoerr.ErrMalformedEntity
	}
	prefix := []byte(patID + keySeparator + scopeKey)
	var rKV map[string][]byte
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
		c := b.Cursor()
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			rKV[string(k)] = v
		}
		return nil
	}); err != nil {
		return auth.Scope{}, err
	}
	return parseKeyValueToScope(rKV)
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
	return keyValueToPAT(kv)
}

func patToKeyValue(pat auth.PAT) (map[string][]byte, error) {
	kv := map[string][]byte{
		idKey:          []byte(pat.ID),
		userKey:        []byte(pat.User),
		nameKey:        []byte(pat.Name),
		descriptionKey: []byte(pat.Description),
		secretKey:      []byte(pat.Secret),
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
	scope, err := parseKeyValueToScope(kv)
	if err != nil {
		return auth.PAT{}, err
	}
	pat.Scope = scope
	return pat, nil
}

// 129eab11-681b-4c62-89e4-7a9ce0eda29d:scope:users:read
// 129eab11-681b-4c62-89e4-7a9ce0eda29d:scope:domains:domain_1:groups:read
// 129eab11-681b-4c62-89e4-7a9ce0eda29d:scope:domains:domain_1:groups:read:group_2
// 129eab11-681b-4c62-89e4-7a9ce0eda29d:scope:domains:domain_1:domain_management:read
// 129eab11-681b-4c62-89e4-7a9ce0eda29d:scope:domains:domain_1:groups:read:group_1

func parseKeyValueToScopeLegacy(kv map[string][]byte) (auth.Scope, error) {
	scope := auth.Scope{
		Domains: make(map[string]auth.DomainScope),
	}
	for key, value := range kv {
		if strings.Index(key, keySeparator+scopeKey+keySeparator) > 0 {
			keyParts := strings.Split(key, keySeparator)

			platformEntityType, err := auth.ParsePlatformEntityType(keyParts[2])
			if err != nil {
				return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
			}

			switch platformEntityType {
			case auth.PlatformUsersScope:
				switch string(value) {
				case string(entityValue):
					if len(keyParts) != 5 {
						return auth.Scope{}, fmt.Errorf("invalid scope key format: %s", key)
					}
					opType, err := auth.ParseOperationType(keyParts[len(keyParts)-2])
					if err != nil {
						return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
					}
					entityID := keyParts[len(keyParts)-1]
					if scope.Users.Operations == nil {
						scope.Users.Operations = make(map[auth.OperationType]auth.ScopeValue)
					}

					if _, oValueExists := scope.Users.Operations[opType]; !oValueExists {
						scope.Users.Operations[opType] = &auth.SelectedIDs{}
					}
					oValue := scope.Users.Operations[opType]
					if err := oValue.AddValues(entityID); err != nil {
						return auth.Scope{}, fmt.Errorf("failed to add scope key %s with entity value %v : %w", key, entityID, err)
					}
					scope.Users.Operations[opType] = oValue
				case string(anyIDValue):
					if len(keyParts) != 4 {
						return auth.Scope{}, fmt.Errorf("invalid scope key format: %s", key)
					}
					opType, err := auth.ParseOperationType(keyParts[len(keyParts)-1])
					if err != nil {
						return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
					}

					if scope.Users.Operations == nil {
						scope.Users.Operations = make(map[auth.OperationType]auth.ScopeValue)
					}

					if oValue, oValueExists := scope.Users.Operations[opType]; oValueExists && oValue != nil {
						if _, ok := oValue.(*auth.AnyIDs); !ok {
							return auth.Scope{}, fmt.Errorf("failed to add scope key %s with entity anyIDs scope value : key already initialized with different type", key)
						}
					}
					scope.Users.Operations[opType] = &auth.AnyIDs{}
				case string(selectedIDsValue):
					if len(keyParts) != 4 {
						return auth.Scope{}, fmt.Errorf("invalid scope key format: %s", key)
					}
					opType, err := auth.ParseOperationType(keyParts[len(keyParts)-1])
					if err != nil {
						return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
					}

					if scope.Users.Operations == nil {
						scope.Users.Operations = make(map[auth.OperationType]auth.ScopeValue)
					}

					oValue, oValueExists := scope.Users.Operations[opType]
					if oValueExists && oValue != nil {
						if _, ok := oValue.(*auth.SelectedIDs); !ok {
							return auth.Scope{}, fmt.Errorf("failed to add scope key %s with entity selectedIDs scope value : key already initialized with different type", key)
						}
					}
					if !oValueExists {
						scope.Users.Operations[opType] = &auth.SelectedIDs{}
					}
				default:
					return auth.Scope{}, fmt.Errorf("key %s have invalid value %v", key, value)
				}

			case auth.PlatformDomainsScope:
				if len(keyParts) < 6 {
					return auth.Scope{}, fmt.Errorf("invalid scope key format: %s", key)
				}
				domainID := keyParts[3]
				if scope.Domains == nil {
					scope.Domains = make(map[string]auth.DomainScope)
				}
				if _, ok := scope.Domains[domainID]; !ok {
					scope.Domains[domainID] = auth.DomainScope{}
				}
				domainScope := scope.Domains[domainID]

				entityType := keyParts[4]

				switch entityType {
				case auth.DomainManagementScope.String():
					switch string(value) {
					case string(entityValue):
						if len(keyParts) != 7 {
							return auth.Scope{}, fmt.Errorf("invalid scope key format: %s", key)
						}
						opType, err := auth.ParseOperationType(keyParts[len(keyParts)-2])
						if err != nil {
							return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
						}
						entityID := keyParts[len(keyParts)-1]

						if domainScope.DomainManagement.Operations == nil {
							domainScope.DomainManagement.Operations = make(map[auth.OperationType]auth.ScopeValue)
						}

						if _, oValueExists := domainScope.DomainManagement.Operations[opType]; !oValueExists {
							domainScope.DomainManagement.Operations[opType] = &auth.SelectedIDs{}
						}
						oValue := domainScope.DomainManagement.Operations[opType]
						if err := oValue.AddValues(entityID); err != nil {
							return auth.Scope{}, fmt.Errorf("failed to add scope key %s with entity value %v : %w", key, entityID, err)
						}
						domainScope.DomainManagement.Operations[opType] = oValue
					case string(anyIDValue):
						if len(keyParts) != 6 {
							return auth.Scope{}, fmt.Errorf("invalid scope key format: %s", key)
						}
						opType, err := auth.ParseOperationType(keyParts[len(keyParts)-1])
						if err != nil {
							return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
						}

						if domainScope.DomainManagement.Operations == nil {
							domainScope.DomainManagement.Operations = make(map[auth.OperationType]auth.ScopeValue)
						}

						if oValue, oValueExists := domainScope.DomainManagement.Operations[opType]; oValueExists && oValue != nil {
							if _, ok := oValue.(*auth.AnyIDs); !ok {
								return auth.Scope{}, fmt.Errorf("failed to add scope key %s with entity anyIDs scope value : key already initialized with different type", key)
							}
						}
						domainScope.DomainManagement.Operations[opType] = &auth.AnyIDs{}
					case string(selectedIDsValue):
						if len(keyParts) != 6 {
							return auth.Scope{}, fmt.Errorf("invalid scope key format: %s", key)
						}
						opType, err := auth.ParseOperationType(keyParts[len(keyParts)-1])
						if err != nil {
							return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
						}

						if domainScope.DomainManagement.Operations == nil {
							domainScope.DomainManagement.Operations = make(map[auth.OperationType]auth.ScopeValue)
						}

						oValue, oValueExists := domainScope.DomainManagement.Operations[opType]
						if oValueExists && oValue != nil {
							if _, ok := oValue.(*auth.SelectedIDs); !ok {
								return auth.Scope{}, fmt.Errorf("failed to add scope key %s with entity selectedIDs scope value : key already initialized with different type", key)
							}
						}
						if !oValueExists {
							domainScope.DomainManagement.Operations[opType] = &auth.SelectedIDs{}
						}
					default:
						return auth.Scope{}, fmt.Errorf("key %s have invalid value %v", key, value)
					}
				default:
					etype, err := auth.ParseDomainEntityType(entityType)
					if err != nil {
						return auth.Scope{}, fmt.Errorf("key %s invalid entity type %s : %w", key, entityType, err)
					}
					if domainScope.Entities == nil {
						domainScope.Entities = make(map[auth.DomainEntityType]auth.OperationScope)
					}
					if _, ok := domainScope.Entities[etype]; !ok {
						domainScope.Entities[etype] = auth.OperationScope{}
					}
					entityOperationScope := domainScope.Entities[etype]
					switch string(value) {
					case string(entityValue):
						if len(keyParts) != 7 {
							return auth.Scope{}, fmt.Errorf("invalid scope key format: %s", key)
						}
						opType, err := auth.ParseOperationType(keyParts[len(keyParts)-2])
						if err != nil {
							return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
						}
						entityID := keyParts[len(keyParts)-1]
						if entityOperationScope.Operations == nil {
							entityOperationScope.Operations = make(map[auth.OperationType]auth.ScopeValue)
						}

						if _, oValueExists := entityOperationScope.Operations[opType]; !oValueExists {
							entityOperationScope.Operations[opType] = &auth.SelectedIDs{}
						}
						oValue := entityOperationScope.Operations[opType]
						if err := oValue.AddValues(entityID); err != nil {
							return auth.Scope{}, fmt.Errorf("failed to add scope key %s with entity value %v : %w", key, entityID, err)
						}
						entityOperationScope.Operations[opType] = oValue
					case string(anyIDValue):
						if len(keyParts) != 6 {
							return auth.Scope{}, fmt.Errorf("invalid scope key format: %s", key)
						}
						opType, err := auth.ParseOperationType(keyParts[len(keyParts)-1])
						if err != nil {
							return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
						}

						if entityOperationScope.Operations == nil {
							entityOperationScope.Operations = make(map[auth.OperationType]auth.ScopeValue)
						}

						if oValue, oValueExists := entityOperationScope.Operations[opType]; oValueExists && oValue != nil {
							if _, ok := oValue.(*auth.AnyIDs); !ok {
								return auth.Scope{}, fmt.Errorf("failed to add scope key %s with entity anyIDs scope value : key already initialized with different type", key)
							}
						}
						entityOperationScope.Operations[opType] = &auth.AnyIDs{}
					case string(selectedIDsValue):
						if len(keyParts) != 6 {
							return auth.Scope{}, fmt.Errorf("invalid scope key format: %s", key)
						}
						opType, err := auth.ParseOperationType(keyParts[len(keyParts)-1])
						if err != nil {
							return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
						}

						if entityOperationScope.Operations == nil {
							entityOperationScope.Operations = make(map[auth.OperationType]auth.ScopeValue)
						}

						oValue, oValueExists := entityOperationScope.Operations[opType]
						if oValueExists && oValue != nil {
							if _, ok := oValue.(*auth.SelectedIDs); !ok {
								return auth.Scope{}, fmt.Errorf("failed to add scope key %s with entity selectedIDs scope value : key already initialized with different type", key)
							}
						}
						if !oValueExists {
							entityOperationScope.Operations[opType] = &auth.SelectedIDs{}
						}
					default:
						return auth.Scope{}, fmt.Errorf("key %s have invalid value %v", key, value)
					}

					domainScope.Entities[etype] = entityOperationScope
				}

				scope.Domains[domainID] = domainScope
			default:
				return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, fmt.Errorf("invalid platform entity type : %s", platformEntityType.String()))
			}
		}
	}
	return scope, nil
}

func parseKeyValueToScope(kv map[string][]byte) (auth.Scope, error) {
	scope := auth.Scope{
		Domains: make(map[string]auth.DomainScope),
	}
	for key, value := range kv {
		if strings.Index(key, keySeparator+scopeKey+keySeparator) > 0 {
			keyParts := strings.Split(key, keySeparator)

			platformEntityType, err := auth.ParsePlatformEntityType(keyParts[2])
			if err != nil {
				return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
			}

			switch platformEntityType {
			case auth.PlatformUsersScope:
				scope.Users, err = parseOperation(platformEntityType, scope.Users, key, keyParts, value)
				if err != nil {
					return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
				}

			case auth.PlatformDomainsScope:
				if len(keyParts) < 6 {
					return auth.Scope{}, fmt.Errorf("invalid scope key format: %s", key)
				}
				domainID := keyParts[3]
				if scope.Domains == nil {
					scope.Domains = make(map[string]auth.DomainScope)
				}
				if _, ok := scope.Domains[domainID]; !ok {
					scope.Domains[domainID] = auth.DomainScope{}
				}
				domainScope := scope.Domains[domainID]

				entityType := keyParts[4]

				switch entityType {
				case auth.DomainManagementScope.String():
					domainScope.DomainManagement, err = parseOperation(platformEntityType, domainScope.DomainManagement, key, keyParts, value)
					if err != nil {
						return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
					}
				default:
					etype, err := auth.ParseDomainEntityType(entityType)
					if err != nil {
						return auth.Scope{}, fmt.Errorf("key %s invalid entity type %s : %w", key, entityType, err)
					}
					if domainScope.Entities == nil {
						domainScope.Entities = make(map[auth.DomainEntityType]auth.OperationScope)
					}
					if _, ok := domainScope.Entities[etype]; !ok {
						domainScope.Entities[etype] = auth.OperationScope{}
					}
					entityOperationScope := domainScope.Entities[etype]
					entityOperationScope, err = parseOperation(platformEntityType, entityOperationScope, key, keyParts, value)
					if err != nil {
						return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
					}
					domainScope.Entities[etype] = entityOperationScope
				}
				scope.Domains[domainID] = domainScope
			default:
				return auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, fmt.Errorf("invalid platform entity type : %s", platformEntityType.String()))
			}
		}
	}
	return scope, nil
}

func parseOperation(platformEntityType auth.PlatformEntityType, opScope auth.OperationScope, key string, keyParts []string, value []byte) (auth.OperationScope, error) {
	if opScope.Operations == nil {
		opScope.Operations = make(map[auth.OperationType]auth.ScopeValue)
	}

	if err := validateOperation(platformEntityType, opScope, key, keyParts, value); err != nil {
		return auth.OperationScope{}, err
	}

	switch string(value) {
	case string(entityValue):
		opType, err := auth.ParseOperationType(keyParts[len(keyParts)-2])
		if err != nil {
			return auth.OperationScope{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}
		entityID := keyParts[len(keyParts)-1]

		if _, oValueExists := opScope.Operations[opType]; !oValueExists {
			opScope.Operations[opType] = &auth.SelectedIDs{}
		}
		oValue := opScope.Operations[opType]
		if err := oValue.AddValues(entityID); err != nil {
			return auth.OperationScope{}, fmt.Errorf("failed to add scope key %s with entity value %v : %w", key, entityID, err)
		}
		opScope.Operations[opType] = oValue
	case string(anyIDValue):
		opType, err := auth.ParseOperationType(keyParts[len(keyParts)-1])
		if err != nil {
			return auth.OperationScope{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}
		if oValue, oValueExists := opScope.Operations[opType]; oValueExists && oValue != nil {
			if _, ok := oValue.(*auth.AnyIDs); !ok {
				return auth.OperationScope{}, fmt.Errorf("failed to add scope key %s with entity anyIDs scope value : key already initialized with different type", key)
			}
		}
		opScope.Operations[opType] = &auth.AnyIDs{}
	case string(selectedIDsValue):
		opType, err := auth.ParseOperationType(keyParts[len(keyParts)-1])
		if err != nil {
			return auth.OperationScope{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}
		oValue, oValueExists := opScope.Operations[opType]
		if oValueExists && oValue != nil {
			if _, ok := oValue.(*auth.SelectedIDs); !ok {
				return auth.OperationScope{}, fmt.Errorf("failed to add scope key %s with entity selectedIDs scope value : key already initialized with different type", key)
			}
		}
		if !oValueExists {
			opScope.Operations[opType] = &auth.SelectedIDs{}
		}
	default:
		return auth.OperationScope{}, fmt.Errorf("key %s have invalid value %v", key, value)
	}
	return opScope, nil
}

func validateOperation(platformEntityType auth.PlatformEntityType, opScope auth.OperationScope, key string, keyParts []string, value []byte) error {
	expectedKeyPartsLength := 0
	switch string(value) {
	case string(entityValue):
		switch platformEntityType {
		case auth.PlatformDomainsScope:
			expectedKeyPartsLength = 7
		case auth.PlatformUsersScope:
			expectedKeyPartsLength = 5
		default:
			return fmt.Errorf("invalid platform entity type : %s", platformEntityType.String())
		}
	case string(selectedIDsValue), string(anyIDValue):
		switch platformEntityType {
		case auth.PlatformDomainsScope:
			expectedKeyPartsLength = 6
		case auth.PlatformUsersScope:
			expectedKeyPartsLength = 4
		default:
			return fmt.Errorf("invalid platform entity type : %s", platformEntityType.String())
		}
	default:
		return fmt.Errorf("key %s have invalid value %v", key, value)
	}
	if len(keyParts) != expectedKeyPartsLength {
		return fmt.Errorf("invalid scope key format: %s", key)
	}
	return nil
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
