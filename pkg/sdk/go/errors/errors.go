// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	errFailedToReadBody = New("failed to read http response body ")
	errRespBodyNotJSON  = New("response body is not vaild JSON ")
	errJSONKeyNotFound  = New("response body expected error message json key not found ")
)

// Error type for SDK
type Error interface {
	// Error implements the error interface.
	Error() string

	// Msg returns error message
	Msg() string

	// Err returns wrapped error
	Err() Error

	// StatusCode returns error status code
	StatusCode() int

	// Status returns error Status
	Status() string
}

var _ Error = (*customError)(nil)

// customError represents a Mainflux  SDK error
type customError struct {
	statusCode int
	status     string
	msg        string
	err        Error
}

func (ce *customError) Error() string {
	if ce == nil {
		return ""
	}
	if ce.err == nil {
		return ce.msg
	}
	return ce.msg + " : " + ce.err.Error()
}

func (ce *customError) Msg() string {
	return ce.msg
}

func (ce *customError) Err() Error {
	return ce.err
}

func (ce *customError) StatusCode() int {
	return ce.statusCode
}

func (ce *customError) Status() string {
	return ce.status
}

// Contains inspects if e2 error is contained in any layer of e1 error
func Contains(e1 error, e2 error) bool {
	if e1 == nil || e2 == nil {
		return e2 == e1
	}
	ce, ok := e1.(Error)
	if ok {
		if ce.Msg() == e2.Error() {
			return true
		}
		return Contains(ce.Err(), e2)
	}
	return e1.Error() == e2.Error()
}

// Wrap returns an Error that wrap err with wrapper
func Wrap(wrapper error, err error) error {
	if wrapper == nil || err == nil {
		return wrapper
	}
	if w, ok := wrapper.(Error); ok {
		return &customError{
			statusCode: w.StatusCode(),
			status:     w.Status(),
			msg:        w.Msg(),
			err:        cast(err),
		}
	}
	return &customError{
		msg: wrapper.Error(),
		err: cast(err),
	}
}

func cast(err error) Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(Error); ok {
		return e
	}
	return &customError{
		msg: err.Error(),
		err: nil,
	}
}

// New returns an Error that formats as the given text.
func New(text string) Error {
	return &customError{
		msg: text,
		err: nil,
	}
}

// New returns an Error that formats as the given text.
func NewWithStatus(statusCode int, status, text string) Error {
	return &customError{
		statusCode: statusCode,
		status:     status,
		msg:        text,
		err:        nil,
	}
}

func EncodeError(resp *http.Response, expectedStatusCode int, errorMsgJsonKey string) error {
	if resp.StatusCode == expectedStatusCode {
		return nil
	}

	b, bErr := io.ReadAll(resp.Body)
	if bErr != nil {
		e := Wrap(errFailedToReadBody, bErr)
		return Wrap(NewWithStatus(resp.StatusCode, resp.Status, ""), e)
	}

	if errorMsgJsonKey == "" {
		return NewWithStatus(resp.StatusCode, resp.Status, string(b))
	}

	var content map[string]interface{}
	err := json.Unmarshal(b, &content)
	if err != nil {
		e := Wrap(errRespBodyNotJSON, New(string(b)))

		return Wrap(NewWithStatus(resp.StatusCode, resp.Status, ""), e)
	}

	if msg, ok := content[errorMsgJsonKey]; ok {
		return NewWithStatus(resp.StatusCode, resp.Status, fmt.Sprintf("%v", msg))
	}
	e := Wrap(errJSONKeyNotFound, New(string(b)))
	return Wrap(NewWithStatus(resp.StatusCode, resp.Status, ""), e)
}
