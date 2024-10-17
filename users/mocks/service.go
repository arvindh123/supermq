// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	authn "github.com/absmach/magistrala/pkg/authn"

	magistrala "github.com/absmach/magistrala"

	mock "github.com/stretchr/testify/mock"

	users "github.com/absmach/magistrala/users"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// DeleteUser provides a mock function with given fields: ctx, session, id
func (_m *Service) DeleteUser(ctx context.Context, session authn.Session, id string) error {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for DeleteUser")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) error); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DisableUser provides a mock function with given fields: ctx, session, id
func (_m *Service) DisableUser(ctx context.Context, session authn.Session, id string) (users.User, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for DisableUser")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (users.User, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) users.User); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EnableUser provides a mock function with given fields: ctx, session, id
func (_m *Service) EnableUser(ctx context.Context, session authn.Session, id string) (users.User, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for EnableUser")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (users.User, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) users.User); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GenerateResetToken provides a mock function with given fields: ctx, email, host
func (_m *Service) GenerateResetToken(ctx context.Context, email string, host string) error {
	ret := _m.Called(ctx, email, host)

	if len(ret) == 0 {
		panic("no return value specified for GenerateResetToken")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, email, host)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Identify provides a mock function with given fields: ctx, session
func (_m *Service) Identify(ctx context.Context, session authn.Session) (string, error) {
	ret := _m.Called(ctx, session)

	if len(ret) == 0 {
		panic("no return value specified for Identify")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session) (string, error)); ok {
		return rf(ctx, session)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session) string); ok {
		r0 = rf(ctx, session)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session) error); ok {
		r1 = rf(ctx, session)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IssueToken provides a mock function with given fields: ctx, identity, secret, domainID
func (_m *Service) IssueToken(ctx context.Context, identity string, secret string, domainID string) (*magistrala.Token, error) {
	ret := _m.Called(ctx, identity, secret, domainID)

	if len(ret) == 0 {
		panic("no return value specified for IssueToken")
	}

	var r0 *magistrala.Token
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) (*magistrala.Token, error)); ok {
		return rf(ctx, identity, secret, domainID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *magistrala.Token); ok {
		r0 = rf(ctx, identity, secret, domainID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*magistrala.Token)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, identity, secret, domainID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListMembers provides a mock function with given fields: ctx, session, objectKind, objectID, pm
func (_m *Service) ListMembers(ctx context.Context, session authn.Session, objectKind string, objectID string, pm users.Page) (users.MembersPage, error) {
	ret := _m.Called(ctx, session, objectKind, objectID, pm)

	if len(ret) == 0 {
		panic("no return value specified for ListMembers")
	}

	var r0 users.MembersPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, users.Page) (users.MembersPage, error)); ok {
		return rf(ctx, session, objectKind, objectID, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string, users.Page) users.MembersPage); ok {
		r0 = rf(ctx, session, objectKind, objectID, pm)
	} else {
		r0 = ret.Get(0).(users.MembersPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string, users.Page) error); ok {
		r1 = rf(ctx, session, objectKind, objectID, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListUsers provides a mock function with given fields: ctx, session, pm
func (_m *Service) ListUsers(ctx context.Context, session authn.Session, pm users.Page) (users.UsersPage, error) {
	ret := _m.Called(ctx, session, pm)

	if len(ret) == 0 {
		panic("no return value specified for ListUsers")
	}

	var r0 users.UsersPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.Page) (users.UsersPage, error)); ok {
		return rf(ctx, session, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.Page) users.UsersPage); ok {
		r0 = rf(ctx, session, pm)
	} else {
		r0 = ret.Get(0).(users.UsersPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, users.Page) error); ok {
		r1 = rf(ctx, session, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OAuthAddUserPolicy provides a mock function with given fields: ctx, user
func (_m *Service) OAuthAddUserPolicy(ctx context.Context, user users.User) error {
	ret := _m.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for OAuthAddUserPolicy")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, users.User) error); ok {
		r0 = rf(ctx, user)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// OAuthCallback provides a mock function with given fields: ctx, user
func (_m *Service) OAuthCallback(ctx context.Context, user users.User) (users.User, error) {
	ret := _m.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for OAuthCallback")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, users.User) (users.User, error)); ok {
		return rf(ctx, user)
	}
	if rf, ok := ret.Get(0).(func(context.Context, users.User) users.User); ok {
		r0 = rf(ctx, user)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, users.User) error); ok {
		r1 = rf(ctx, user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RefreshToken provides a mock function with given fields: ctx, session, refreshToken, domainID
func (_m *Service) RefreshToken(ctx context.Context, session authn.Session, refreshToken string, domainID string) (*magistrala.Token, error) {
	ret := _m.Called(ctx, session, refreshToken, domainID)

	if len(ret) == 0 {
		panic("no return value specified for RefreshToken")
	}

	var r0 *magistrala.Token
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) (*magistrala.Token, error)); ok {
		return rf(ctx, session, refreshToken, domainID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) *magistrala.Token); ok {
		r0 = rf(ctx, session, refreshToken, domainID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*magistrala.Token)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string) error); ok {
		r1 = rf(ctx, session, refreshToken, domainID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RegisterUser provides a mock function with given fields: ctx, session, user, selfRegister
func (_m *Service) RegisterUser(ctx context.Context, session authn.Session, user users.User, selfRegister bool) (users.User, error) {
	ret := _m.Called(ctx, session, user, selfRegister)

	if len(ret) == 0 {
		panic("no return value specified for RegisterUser")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User, bool) (users.User, error)); ok {
		return rf(ctx, session, user, selfRegister)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User, bool) users.User); ok {
		r0 = rf(ctx, session, user, selfRegister)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, users.User, bool) error); ok {
		r1 = rf(ctx, session, user, selfRegister)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResetSecret provides a mock function with given fields: ctx, session, secret
func (_m *Service) ResetSecret(ctx context.Context, session authn.Session, secret string) error {
	ret := _m.Called(ctx, session, secret)

	if len(ret) == 0 {
		panic("no return value specified for ResetSecret")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) error); ok {
		r0 = rf(ctx, session, secret)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SearchUsers provides a mock function with given fields: ctx, pm
func (_m *Service) SearchUsers(ctx context.Context, pm users.Page) (users.UsersPage, error) {
	ret := _m.Called(ctx, pm)

	if len(ret) == 0 {
		panic("no return value specified for SearchUsers")
	}

	var r0 users.UsersPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, users.Page) (users.UsersPage, error)); ok {
		return rf(ctx, pm)
	}
	if rf, ok := ret.Get(0).(func(context.Context, users.Page) users.UsersPage); ok {
		r0 = rf(ctx, pm)
	} else {
		r0 = ret.Get(0).(users.UsersPage)
	}

	if rf, ok := ret.Get(1).(func(context.Context, users.Page) error); ok {
		r1 = rf(ctx, pm)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendPasswordReset provides a mock function with given fields: ctx, host, email, user, token
func (_m *Service) SendPasswordReset(ctx context.Context, host string, email string, user string, token string) error {
	ret := _m.Called(ctx, host, email, user, token)

	if len(ret) == 0 {
		panic("no return value specified for SendPasswordReset")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string) error); ok {
		r0 = rf(ctx, host, email, user, token)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateProfilePicture provides a mock function with given fields: ctx, session, user
func (_m *Service) UpdateProfilePicture(ctx context.Context, session authn.Session, user users.User) (users.User, error) {
	ret := _m.Called(ctx, session, user)

	if len(ret) == 0 {
		panic("no return value specified for UpdateProfilePicture")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User) (users.User, error)); ok {
		return rf(ctx, session, user)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User) users.User); ok {
		r0 = rf(ctx, session, user)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, users.User) error); ok {
		r1 = rf(ctx, session, user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateUser provides a mock function with given fields: ctx, session, user
func (_m *Service) UpdateUser(ctx context.Context, session authn.Session, user users.User) (users.User, error) {
	ret := _m.Called(ctx, session, user)

	if len(ret) == 0 {
		panic("no return value specified for UpdateUser")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User) (users.User, error)); ok {
		return rf(ctx, session, user)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User) users.User); ok {
		r0 = rf(ctx, session, user)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, users.User) error); ok {
		r1 = rf(ctx, session, user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateUserIdentity provides a mock function with given fields: ctx, session, id, identity
func (_m *Service) UpdateUserIdentity(ctx context.Context, session authn.Session, id string, identity string) (users.User, error) {
	ret := _m.Called(ctx, session, id, identity)

	if len(ret) == 0 {
		panic("no return value specified for UpdateUserIdentity")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) (users.User, error)); ok {
		return rf(ctx, session, id, identity)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) users.User); ok {
		r0 = rf(ctx, session, id, identity)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string) error); ok {
		r1 = rf(ctx, session, id, identity)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateUserNames provides a mock function with given fields: ctx, session, usr
func (_m *Service) UpdateUserNames(ctx context.Context, session authn.Session, usr users.User) (users.User, error) {
	ret := _m.Called(ctx, session, usr)

	if len(ret) == 0 {
		panic("no return value specified for UpdateUserNames")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User) (users.User, error)); ok {
		return rf(ctx, session, usr)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User) users.User); ok {
		r0 = rf(ctx, session, usr)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, users.User) error); ok {
		r1 = rf(ctx, session, usr)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateUserRole provides a mock function with given fields: ctx, session, user
func (_m *Service) UpdateUserRole(ctx context.Context, session authn.Session, user users.User) (users.User, error) {
	ret := _m.Called(ctx, session, user)

	if len(ret) == 0 {
		panic("no return value specified for UpdateUserRole")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User) (users.User, error)); ok {
		return rf(ctx, session, user)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User) users.User); ok {
		r0 = rf(ctx, session, user)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, users.User) error); ok {
		r1 = rf(ctx, session, user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateUserSecret provides a mock function with given fields: ctx, session, oldSecret, newSecret
func (_m *Service) UpdateUserSecret(ctx context.Context, session authn.Session, oldSecret string, newSecret string) (users.User, error) {
	ret := _m.Called(ctx, session, oldSecret, newSecret)

	if len(ret) == 0 {
		panic("no return value specified for UpdateUserSecret")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) (users.User, error)); ok {
		return rf(ctx, session, oldSecret, newSecret)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string, string) users.User); ok {
		r0 = rf(ctx, session, oldSecret, newSecret)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string, string) error); ok {
		r1 = rf(ctx, session, oldSecret, newSecret)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateUserTags provides a mock function with given fields: ctx, session, user
func (_m *Service) UpdateUserTags(ctx context.Context, session authn.Session, user users.User) (users.User, error) {
	ret := _m.Called(ctx, session, user)

	if len(ret) == 0 {
		panic("no return value specified for UpdateUserTags")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User) (users.User, error)); ok {
		return rf(ctx, session, user)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, users.User) users.User); ok {
		r0 = rf(ctx, session, user)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, users.User) error); ok {
		r1 = rf(ctx, session, user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ViewProfile provides a mock function with given fields: ctx, session
func (_m *Service) ViewProfile(ctx context.Context, session authn.Session) (users.User, error) {
	ret := _m.Called(ctx, session)

	if len(ret) == 0 {
		panic("no return value specified for ViewProfile")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session) (users.User, error)); ok {
		return rf(ctx, session)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session) users.User); ok {
		r0 = rf(ctx, session)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session) error); ok {
		r1 = rf(ctx, session)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ViewUser provides a mock function with given fields: ctx, session, id
func (_m *Service) ViewUser(ctx context.Context, session authn.Session, id string) (users.User, error) {
	ret := _m.Called(ctx, session, id)

	if len(ret) == 0 {
		panic("no return value specified for ViewUser")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (users.User, error)); ok {
		return rf(ctx, session, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) users.User); ok {
		r0 = rf(ctx, session, id)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ViewUserByUserName provides a mock function with given fields: ctx, session, userName
func (_m *Service) ViewUserByUserName(ctx context.Context, session authn.Session, userName string) (users.User, error) {
	ret := _m.Called(ctx, session, userName)

	if len(ret) == 0 {
		panic("no return value specified for ViewUserByUserName")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) (users.User, error)); ok {
		return rf(ctx, session, userName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, authn.Session, string) users.User); ok {
		r0 = rf(ctx, session, userName)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(context.Context, authn.Session, string) error); ok {
		r1 = rf(ctx, session, userName)
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
