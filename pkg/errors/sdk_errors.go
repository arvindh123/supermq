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
type SDKError interface {
	baseError

	Err() SDKError
	// StatusCode returns error status code
	StatusCode() int
}

var _ SDKError = (*customSDKError)(nil)

// customError represents a Mainflux  SDK error
type customSDKError struct {
	statusCode int
	msg        string
	err        SDKError
}

func (ce *customSDKError) Error() string {
	if ce == nil {
		return ""
	}
	if ce.err == nil {
		return ce.msg
	}
	return ce.msg + " : " + ce.err.Error()
}

func (ce *customSDKError) Msg() string {
	return ce.msg
}

func (ce *customSDKError) Err() SDKError {
	return ce.err
}

func (ce *customSDKError) StatusCode() int {
	if err := ce.Err(); err != nil && ce.statusCode == 0 {
		return err.StatusCode()
	}
	return ce.statusCode
}

func castSDKError(err error) SDKError {
	if err == nil {
		return nil
	}
	if e, ok := err.(SDKError); ok {
		return e
	}
	return &customSDKError{
		msg: err.Error(),
		err: nil,
	}
}

// New returns an Error that formats as the given text.
func NewSDKError(text string) SDKError {
	return &customSDKError{
		msg: text,
		err: nil,
	}
}

// New returns an Error that formats as the given text.
func NewSDKErrorWithStatus(statusCode int, text string) SDKError {
	return &customSDKError{
		statusCode: statusCode,
		msg:        text,
		err:        nil,
	}
}

// CheckError will check for error in *http.Response
func CheckError(resp *http.Response, expectedStatusCodes ...int) error {
	for expectedStatusCode := range expectedStatusCodes {
		if resp.StatusCode == expectedStatusCode {
			return nil
		}
	}

	b, bErr := io.ReadAll(resp.Body)
	if bErr != nil {
		e := Wrap(errFailedToReadBody, bErr)
		return Wrap(NewSDKErrorWithStatus(resp.StatusCode, ""), e)
	}

	var content map[string]interface{}
	err := json.Unmarshal(b, &content)
	if err != nil {
		e := Wrap(errRespBodyNotJSON, New(string(b)))

		return Wrap(NewSDKErrorWithStatus(resp.StatusCode, ""), e)
	}

	if msg, ok := content["error"]; ok {
		return NewSDKErrorWithStatus(resp.StatusCode, fmt.Sprintf("%v", msg))
	}
	e := Wrap(errJSONKeyNotFound, New(string(b)))
	return Wrap(NewSDKErrorWithStatus(resp.StatusCode, ""), e)
}
