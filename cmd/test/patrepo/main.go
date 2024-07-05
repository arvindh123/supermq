package main

import (
	"context"
	"fmt"
	"time"

	"github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/auth/bolt"
	boltclient "github.com/absmach/magistrala/internal/clients/bolt"
	"github.com/caarlos0/env/v10"
	"github.com/google/uuid"
)

func main() {

	boltDBConfig := boltclient.Config{}
	if err := env.ParseWithOptions(&boltDBConfig, env.Options{}); err != nil {
		panic(err)
	}

	client, err := boltclient.Connect(boltDBConfig, bolt.Init)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	patrepo := bolt.NewPATSRepository(client, boltDBConfig.Bucket)

	pat := auth.PAT{
		ID:        uuid.New().String(),
		User:      "user123",
		Name:      "user 123",
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Scope: auth.Scope{
			Users: auth.OperationScope{
				Operations: map[auth.OperationType]auth.ScopeValue{
					auth.ReadOp: &auth.AnyIDs{},
				},
			},
			Domains: map[string]auth.DomainScope{
				"domain_1": {
					DomainManagement: auth.OperationScope{
						Operations: map[auth.OperationType]auth.ScopeValue{
							auth.ReadOp: &auth.AnyIDs{},
						},
					},
					Entities: map[auth.DomainEntityType]auth.OperationScope{
						auth.DomainGroupsScope: {
							Operations: map[auth.OperationType]auth.ScopeValue{
								auth.ReadOp: &auth.SelectedIDs{"group_1": {}, "group_2": {}},
							},
						},
					},
				},
			},
		},
	}

	if err := patrepo.Save(context.Background(), pat); err != nil {
		panic(err)
	}

	rPAT, err := patrepo.Retrieve(context.Background(), pat.User, pat.ID)
	if err != nil {
		panic(err)
	}
	fmt.Println(rPAT.String())

	fmt.Println(patrepo.CheckScopeEntry(context.Background(), pat.User, pat.ID, auth.PlatformDomainsScope, "domain_1", auth.DomainGroupsScope, auth.ReadOp, "group_1"))
	fmt.Println(patrepo.CheckScopeEntry(context.Background(), pat.User, pat.ID, auth.PlatformDomainsScope, "domain_1", auth.DomainGroupsScope, auth.ReadOp, "group_2"))
	fmt.Println(patrepo.CheckScopeEntry(context.Background(), pat.User, pat.ID, auth.PlatformDomainsScope, "domain_1", auth.DomainGroupsScope, auth.UpdateOp, "group_1"))
	fmt.Println(patrepo.CheckScopeEntry(context.Background(), pat.User, pat.ID, auth.PlatformDomainsScope, "domain_1", auth.DomainGroupsScope, auth.UpdateOp, "group_2"))
	fmt.Println(patrepo.CheckScopeEntry(context.Background(), pat.User, pat.ID, auth.PlatformDomainsScope, "domain_1", auth.DomainThingsScope, auth.UpdateOp, "group_2"))
	fmt.Println(patrepo.CheckScopeEntry(context.Background(), pat.User, pat.ID, auth.PlatformUsersScope, "", auth.DomainNullScope, auth.ReadOp, "user_123"))
	fmt.Println(patrepo.CheckScopeEntry(context.Background(), pat.User, pat.ID, auth.PlatformUsersScope, "", auth.DomainNullScope, auth.UpdateOp, "user_123"))

}
