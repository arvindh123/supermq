// Code generated by mockery v2.43.2. DO NOT EDIT.

// Copyright (c) Abstract Machines

package mocks

import (
	context "context"

	policies "github.com/absmach/supermq/pkg/policies"
	mock "github.com/stretchr/testify/mock"
)

// CallBack is an autogenerated mock type for the CallBack type
type CallBack struct {
	mock.Mock
}

// Authorize provides a mock function with given fields: ctx, pr
func (_m *CallBack) Authorize(ctx context.Context, pr policies.Policy) error {
	ret := _m.Called(ctx, pr)

	if len(ret) == 0 {
		panic("no return value specified for Authorize")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, policies.Policy) error); ok {
		r0 = rf(ctx, pr)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewCallBack creates a new instance of CallBack. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCallBack(t interface {
	mock.TestingT
	Cleanup(func())
}) *CallBack {
	mock := &CallBack{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
