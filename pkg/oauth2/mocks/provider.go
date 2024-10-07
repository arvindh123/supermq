// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	users "github.com/absmach/magistrala/users"

	xoauth2 "golang.org/x/oauth2"
)

// Provider is an autogenerated mock type for the Provider type
type Provider struct {
	mock.Mock
}

// ErrorURL provides a mock function with given fields:
func (_m *Provider) ErrorURL() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for ErrorURL")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Exchange provides a mock function with given fields: ctx, code
func (_m *Provider) Exchange(ctx context.Context, code string) (xoauth2.Token, error) {
	ret := _m.Called(ctx, code)

	if len(ret) == 0 {
		panic("no return value specified for Exchange")
	}

	var r0 xoauth2.Token
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (xoauth2.Token, error)); ok {
		return rf(ctx, code)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) xoauth2.Token); ok {
		r0 = rf(ctx, code)
	} else {
		r0 = ret.Get(0).(xoauth2.Token)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, code)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsEnabled provides a mock function with given fields:
func (_m *Provider) IsEnabled() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsEnabled")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Name provides a mock function with given fields:
func (_m *Provider) Name() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Name")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// RedirectURL provides a mock function with given fields:
func (_m *Provider) RedirectURL() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for RedirectURL")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// State provides a mock function with given fields:
func (_m *Provider) State() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for State")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// UserInfo provides a mock function with given fields: accessToken
func (_m *Provider) UserInfo(accessToken string) (users.User, error) {
	ret := _m.Called(accessToken)

	if len(ret) == 0 {
		panic("no return value specified for UserInfo")
	}

	var r0 users.User
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (users.User, error)); ok {
		return rf(accessToken)
	}
	if rf, ok := ret.Get(0).(func(string) users.User); ok {
		r0 = rf(accessToken)
	} else {
		r0 = ret.Get(0).(users.User)
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(accessToken)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewProvider creates a new instance of Provider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *Provider {
	mock := &Provider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
