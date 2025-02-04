// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	authn "github.com/absmach/supermq/pkg/authn"

	groups "github.com/absmach/supermq/groups"

	mock "github.com/stretchr/testify/mock"

	roles "github.com/absmach/supermq/pkg/roles"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// AddChildrenGroups provides a mock function with given fields: ctx, session, id, childrenGroupIDs
func (_m *Service) AddChildrenGroups(ctx context.Context, session authn.Session, id string, childrenGroupIDs []string) error {
	ret := _m.Called(ctx, session, id, childrenGroupIDs)

	if len(ret) == 0 {
		panic("no return value specified for AddChildrenGroups")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, []string) error); ok {
		r0 = rf(ctx, session, id, childrenGroupIDs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddParentGroup provides a mock function with given fields: ctx, session, id, parentID
func (_m *Service) AddParentGroup(ctx context.Context, session authn.Session, id string, parentID string) error {
	ret := _m.Called(ctx, session, id, parentID)

	if len(ret) == 0 {
		panic("no return value specified for AddParentGroup")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) error); ok {
		r0 = rf(ctx, session, id, parentID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddRole provides a mock function with given fields: ctx, session, entityID, roleName, optionalActions, optionalMembers
func (_m *Service) AddRole(ctx context.Context, session authn.Session, entityID string, roleName string, optionalActions []string, optionalMembers []string) (roles.RoleProvision, error) {
	ret := _m.Called(ctx, session, entityID, roleName, optionalActions, optionalMembers)

	if len(ret) == 0 {
		panic("no return value specified for AddRole")
	}

	var r0 roles.RoleProvision
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string, []string) (roles.RoleProvision, error)); ok {
		return rf(ctx, session, entityID, roleName, optionalActions, optionalMembers)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string, []string) roles.RoleProvision); ok {
		r0 = rf(ctx, session, entityID, roleName, optionalActions, optionalMembers)
	} else {
		r0 = ret.Get(0).(roles.RoleProvision)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, []string, []string) error); ok {
		r1 = rf(ctx, session, entityID, roleName, optionalActions, optionalMembers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateGroup provides a mock function with given fields: ctx, session, g
func (_m *Service) CreateGroup(ctx context.Context, session authn.Session, g groups.Group) (groups.Group, []roles.RoleProvision, error) {
	ret := _m.Called(ctx, session, g)

	if len(ret) == 0 {
		panic("no return value specified for CreateGroup")
	}

	var r0 groups.Group
	var r1 []roles.RoleProvision
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, groups.Group) (groups.Group, []roles.RoleProvision, error)); ok {
		return rf(ctx, session, g)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, groups.Group) groups.Group); ok {
		r0 = rf(ctx, session, g)
	} else {
		r0 = ret.Get(0).(groups.Group)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, groups.Group) []roles.RoleProvision); ok {
		r1 = rf(ctx, session, g)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]roles.RoleProvision)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, authn.Session, groups.Group) error); ok {
		r2 = rf(ctx, session, g)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DeleteGroup provides a mock function with given fields: ctx, session, id
func (_m *Service) DeleteGroup(ctx context.Context, session authn.Session, id string) error {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteGroup")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) error); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DisableGroup provides a mock function with given fields: ctx, session, id
func (_m *Service) DisableGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for DisableGroup")
	}

	var r0 groups.Group
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (groups.Group, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) groups.Group); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(groups.Group)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EnableGroup provides a mock function with given fields: ctx, session, id
func (_m *Service) EnableGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for EnableGroup")
	}

	var r0 groups.Group
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (groups.Group, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) groups.Group); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(groups.Group)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAvailableActions provides a mock function with given fields: ctx, session
func (_m *Service) ListAvailableActions(ctx context.Context, session authn.Session) ([]string, error) {
	ret := _m.Called(ctx, session)

	if len(ret) == 0 {
		panic("no return value specified for ListAvailableActions")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session) ([]string, error)); ok {
		return rf(ctx, session)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session) []string); ok {
		r0 = rf(ctx, session)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session) error); ok {
		r1 = rf(ctx, session)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListChildrenGroups provides a mock function with given fields: ctx, session, id, startLevel, endLevel, pm
func (_m *Service) ListChildrenGroups(ctx context.Context, session authn.Session, id string, startLevel int64, endLevel int64, pm groups.PageMeta) (groups.Page, error) {
	ret := _m.Called(ctx, session, id, startLevel, endLevel, pm)

	if len(ret) == 0 {
		panic("no return value specified for ListChildrenGroups")
	}

	var r0 groups.Page
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, int64, int64, groups.PageMeta) (groups.Page, error)); ok {
		return rf(ctx, session, id, startLevel, endLevel, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, int64, int64, groups.PageMeta) groups.Page); ok {
		r0 = rf(ctx, session, id, startLevel, endLevel, pm)
	} else {
		r0 = ret.Get(0).(groups.Page)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, int64, int64, groups.PageMeta) error); ok {
		r1 = rf(ctx, session, id, startLevel, endLevel, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListEntityMembers provides a mock function with given fields: ctx, session, entityID, pq
func (_m *Service) ListEntityMembers(ctx context.Context, session authn.Session, entityID string, pq roles.MembersRolePageQuery) (roles.MembersRolePage, error) {
	ret := _m.Called(ctx, session, entityID, pq)

	if len(ret) == 0 {
		panic("no return value specified for ListEntityMembers")
	}

	var r0 roles.MembersRolePage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, roles.MembersRolePageQuery) (roles.MembersRolePage, error)); ok {
		return rf(ctx, session, entityID, pq)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, roles.MembersRolePageQuery) roles.MembersRolePage); ok {
		r0 = rf(ctx, session, entityID, pq)
	} else {
		r0 = ret.Get(0).(roles.MembersRolePage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, roles.MembersRolePageQuery) error); ok {
		r1 = rf(ctx, session, entityID, pq)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListGroups provides a mock function with given fields: ctx, session, pm
func (_m *Service) ListGroups(ctx context.Context, session authn.Session, pm groups.PageMeta) (groups.Page, error) {
	ret := _m.Called(ctx, session, pm)

	if len(ret) == 0 {
		panic("no return value specified for ListGroups")
	}

	var r0 groups.Page
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, groups.PageMeta) (groups.Page, error)); ok {
		return rf(ctx, session, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, groups.PageMeta) groups.Page); ok {
		r0 = rf(ctx, session, pm)
	} else {
		r0 = ret.Get(0).(groups.Page)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, groups.PageMeta) error); ok {
		r1 = rf(ctx, session, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListUserGroups provides a mock function with given fields: ctx, session, userID, pm
func (_m *Service) ListUserGroups(ctx context.Context, session authn.Session, userID string, pm groups.PageMeta) (groups.Page, error) {
	ret := _m.Called(ctx, session, userID, pm)

	if len(ret) == 0 {
		panic("no return value specified for ListUserGroups")
	}

	var r0 groups.Page
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, groups.PageMeta) (groups.Page, error)); ok {
		return rf(ctx, session, userID, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, groups.PageMeta) groups.Page); ok {
		r0 = rf(ctx, session, userID, pm)
	} else {
		r0 = ret.Get(0).(groups.Page)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, groups.PageMeta) error); ok {
		r1 = rf(ctx, session, userID, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveAllChildrenGroups provides a mock function with given fields: ctx, session, id
func (_m *Service) RemoveAllChildrenGroups(ctx context.Context, session authn.Session, id string) error {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for RemoveAllChildrenGroups")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) error); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveChildrenGroups provides a mock function with given fields: ctx, session, id, childrenGroupIDs
func (_m *Service) RemoveChildrenGroups(ctx context.Context, session authn.Session, id string, childrenGroupIDs []string) error {
	ret := _m.Called(ctx, session, id, childrenGroupIDs)

	if len(ret) == 0 {
		panic("no return value specified for RemoveChildrenGroups")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, []string) error); ok {
		r0 = rf(ctx, session, id, childrenGroupIDs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveMemberFromAllRoles provides a mock function with given fields: ctx, session, memberID
func (_m *Service) RemoveMemberFromAllRoles(ctx context.Context, session authn.Session, memberID string) error {
	ret := _m.Called(ctx, session, memberID)

	if len(ret) == 0 {
		panic("no return value specified for RemoveMemberFromAllRoles")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) error); ok {
		r0 = rf(ctx, session, memberID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveMemberFromEntity provides a mock function with given fields: ctx, session, entityID, memberID
func (_m *Service) RemoveMemberFromEntity(ctx context.Context, session authn.Session, entityID string, memberID string) error {
	ret := _m.Called(ctx, session, entityID, memberID)

	if len(ret) == 0 {
		panic("no return value specified for RemoveMemberFromEntity")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) error); ok {
		r0 = rf(ctx, session, entityID, memberID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveParentGroup provides a mock function with given fields: ctx, session, id
func (_m *Service) RemoveParentGroup(ctx context.Context, session authn.Session, id string) error {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for RemoveParentGroup")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) error); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveRole provides a mock function with given fields: ctx, session, entityID, roleID
func (_m *Service) RemoveRole(ctx context.Context, session authn.Session, entityID string, roleID string) error {
	ret := _m.Called(ctx, session, entityID, roleID)

	if len(ret) == 0 {
		panic("no return value specified for RemoveRole")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) error); ok {
		r0 = rf(ctx, session, entityID, roleID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RetrieveAllRoles provides a mock function with given fields: ctx, session, entityID, limit, offset
func (_m *Service) RetrieveAllRoles(ctx context.Context, session authn.Session, entityID string, limit uint64, offset uint64) (roles.RolePage, error) {
	ret := _m.Called(ctx, session, entityID, limit, offset)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveAllRoles")
	}

	var r0 roles.RolePage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, uint64, uint64) (roles.RolePage, error)); ok {
		return rf(ctx, session, entityID, limit, offset)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, uint64, uint64) roles.RolePage); ok {
		r0 = rf(ctx, session, entityID, limit, offset)
	} else {
		r0 = ret.Get(0).(roles.RolePage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, uint64, uint64) error); ok {
		r1 = rf(ctx, session, entityID, limit, offset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveGroupHierarchy provides a mock function with given fields: ctx, session, id, hm
func (_m *Service) RetrieveGroupHierarchy(ctx context.Context, session authn.Session, id string, hm groups.HierarchyPageMeta) (groups.HierarchyPage, error) {
	ret := _m.Called(ctx, session, id, hm)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveGroupHierarchy")
	}

	var r0 groups.HierarchyPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, groups.HierarchyPageMeta) (groups.HierarchyPage, error)); ok {
		return rf(ctx, session, id, hm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, groups.HierarchyPageMeta) groups.HierarchyPage); ok {
		r0 = rf(ctx, session, id, hm)
	} else {
		r0 = ret.Get(0).(groups.HierarchyPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, groups.HierarchyPageMeta) error); ok {
		r1 = rf(ctx, session, id, hm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveRole provides a mock function with given fields: ctx, session, entityID, roleID
func (_m *Service) RetrieveRole(ctx context.Context, session authn.Session, entityID string, roleID string) (roles.Role, error) {
	ret := _m.Called(ctx, session, entityID, roleID)

	if len(ret) == 0 {
		panic("no return value specified for RetrieveRole")
	}

	var r0 roles.Role
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) (roles.Role, error)); ok {
		return rf(ctx, session, entityID, roleID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) roles.Role); ok {
		r0 = rf(ctx, session, entityID, roleID)
	} else {
		r0 = ret.Get(0).(roles.Role)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string) error); ok {
		r1 = rf(ctx, session, entityID, roleID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleAddActions provides a mock function with given fields: ctx, session, entityID, roleID, actions
func (_m *Service) RoleAddActions(ctx context.Context, session authn.Session, entityID string, roleID string, actions []string) ([]string, error) {
	ret := _m.Called(ctx, session, entityID, roleID, actions)

	if len(ret) == 0 {
		panic("no return value specified for RoleAddActions")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) ([]string, error)); ok {
		return rf(ctx, session, entityID, roleID, actions)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) []string); ok {
		r0 = rf(ctx, session, entityID, roleID, actions)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r1 = rf(ctx, session, entityID, roleID, actions)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleAddMembers provides a mock function with given fields: ctx, session, entityID, roleID, members
func (_m *Service) RoleAddMembers(ctx context.Context, session authn.Session, entityID string, roleID string, members []string) ([]string, error) {
	ret := _m.Called(ctx, session, entityID, roleID, members)

	if len(ret) == 0 {
		panic("no return value specified for RoleAddMembers")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) ([]string, error)); ok {
		return rf(ctx, session, entityID, roleID, members)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) []string); ok {
		r0 = rf(ctx, session, entityID, roleID, members)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r1 = rf(ctx, session, entityID, roleID, members)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleCheckActionsExists provides a mock function with given fields: ctx, session, entityID, roleID, actions
func (_m *Service) RoleCheckActionsExists(ctx context.Context, session authn.Session, entityID string, roleID string, actions []string) (bool, error) {
	ret := _m.Called(ctx, session, entityID, roleID, actions)

	if len(ret) == 0 {
		panic("no return value specified for RoleCheckActionsExists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) (bool, error)); ok {
		return rf(ctx, session, entityID, roleID, actions)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) bool); ok {
		r0 = rf(ctx, session, entityID, roleID, actions)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r1 = rf(ctx, session, entityID, roleID, actions)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleCheckMembersExists provides a mock function with given fields: ctx, session, entityID, roleID, members
func (_m *Service) RoleCheckMembersExists(ctx context.Context, session authn.Session, entityID string, roleID string, members []string) (bool, error) {
	ret := _m.Called(ctx, session, entityID, roleID, members)

	if len(ret) == 0 {
		panic("no return value specified for RoleCheckMembersExists")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) (bool, error)); ok {
		return rf(ctx, session, entityID, roleID, members)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) bool); ok {
		r0 = rf(ctx, session, entityID, roleID, members)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r1 = rf(ctx, session, entityID, roleID, members)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleListActions provides a mock function with given fields: ctx, session, entityID, roleID
func (_m *Service) RoleListActions(ctx context.Context, session authn.Session, entityID string, roleID string) ([]string, error) {
	ret := _m.Called(ctx, session, entityID, roleID)

	if len(ret) == 0 {
		panic("no return value specified for RoleListActions")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) ([]string, error)); ok {
		return rf(ctx, session, entityID, roleID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) []string); ok {
		r0 = rf(ctx, session, entityID, roleID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string) error); ok {
		r1 = rf(ctx, session, entityID, roleID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleListMembers provides a mock function with given fields: ctx, session, entityID, roleID, limit, offset
func (_m *Service) RoleListMembers(ctx context.Context, session authn.Session, entityID string, roleID string, limit uint64, offset uint64) (roles.MembersPage, error) {
	ret := _m.Called(ctx, session, entityID, roleID, limit, offset)

	if len(ret) == 0 {
		panic("no return value specified for RoleListMembers")
	}

	var r0 roles.MembersPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, uint64, uint64) (roles.MembersPage, error)); ok {
		return rf(ctx, session, entityID, roleID, limit, offset)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, uint64, uint64) roles.MembersPage); ok {
		r0 = rf(ctx, session, entityID, roleID, limit, offset)
	} else {
		r0 = ret.Get(0).(roles.MembersPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, uint64, uint64) error); ok {
		r1 = rf(ctx, session, entityID, roleID, limit, offset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RoleRemoveActions provides a mock function with given fields: ctx, session, entityID, roleID, actions
func (_m *Service) RoleRemoveActions(ctx context.Context, session authn.Session, entityID string, roleID string, actions []string) error {
	ret := _m.Called(ctx, session, entityID, roleID, actions)

	if len(ret) == 0 {
		panic("no return value specified for RoleRemoveActions")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r0 = rf(ctx, session, entityID, roleID, actions)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RoleRemoveAllActions provides a mock function with given fields: ctx, session, entityID, roleID
func (_m *Service) RoleRemoveAllActions(ctx context.Context, session authn.Session, entityID string, roleID string) error {
	ret := _m.Called(ctx, session, entityID, roleID)

	if len(ret) == 0 {
		panic("no return value specified for RoleRemoveAllActions")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) error); ok {
		r0 = rf(ctx, session, entityID, roleID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RoleRemoveAllMembers provides a mock function with given fields: ctx, session, entityID, roleID
func (_m *Service) RoleRemoveAllMembers(ctx context.Context, session authn.Session, entityID string, roleID string) error {
	ret := _m.Called(ctx, session, entityID, roleID)

	if len(ret) == 0 {
		panic("no return value specified for RoleRemoveAllMembers")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) error); ok {
		r0 = rf(ctx, session, entityID, roleID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RoleRemoveMembers provides a mock function with given fields: ctx, session, entityID, roleID, members
func (_m *Service) RoleRemoveMembers(ctx context.Context, session authn.Session, entityID string, roleID string, members []string) error {
	ret := _m.Called(ctx, session, entityID, roleID, members)

	if len(ret) == 0 {
		panic("no return value specified for RoleRemoveMembers")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, []string) error); ok {
		r0 = rf(ctx, session, entityID, roleID, members)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateGroup provides a mock function with given fields: ctx, session, g
func (_m *Service) UpdateGroup(ctx context.Context, session authn.Session, g groups.Group) (groups.Group, error) {
	ret := _m.Called(ctx, session, g)

	if len(ret) == 0 {
		panic("no return value specified for UpdateGroup")
	}

	var r0 groups.Group
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, groups.Group) (groups.Group, error)); ok {
		return rf(ctx, session, g)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, groups.Group) groups.Group); ok {
		r0 = rf(ctx, session, g)
	} else {
		r0 = ret.Get(0).(groups.Group)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, groups.Group) error); ok {
		r1 = rf(ctx, session, g)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateRoleName provides a mock function with given fields: ctx, session, entityID, roleID, newRoleName
func (_m *Service) UpdateRoleName(ctx context.Context, session authn.Session, entityID string, roleID string, newRoleName string) (roles.Role, error) {
	ret := _m.Called(ctx, session, entityID, roleID, newRoleName)

	if len(ret) == 0 {
		panic("no return value specified for UpdateRoleName")
	}

	var r0 roles.Role
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, string) (roles.Role, error)); ok {
		return rf(ctx, session, entityID, roleID, newRoleName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, string) roles.Role); ok {
		r0 = rf(ctx, session, entityID, roleID, newRoleName)
	} else {
		r0 = ret.Get(0).(roles.Role)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, string) error); ok {
		r1 = rf(ctx, session, entityID, roleID, newRoleName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ViewGroup provides a mock function with given fields: ctx, session, id
func (_m *Service) ViewGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for ViewGroup")
	}

	var r0 groups.Group
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (groups.Group, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) groups.Group); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(groups.Group)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService(t interface {
	mock.TestingT
	Cleanup(func())
}) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
