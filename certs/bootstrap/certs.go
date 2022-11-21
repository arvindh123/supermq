// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/pkg/errors"

	"os"
	"sync"

	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
)

const authHeaderPrefix = "Bearer"

var (
	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("unauthorized access to Bootstrap service")

	// ErrUnexpectedBSResponse indicates unexpected response from Bootstrap service.
	ErrUnexpectedBSResponse = errors.New("unexpected Bootstrap service response")

	ErrUnableToAccess = errors.New("unable to access bootstrap service")

	ErrFailedToLogin = errors.New("Failed to login")

	ErrFailedToReadResponseBody = errors.New("failed to read Bootstrap response body ")
)

type bootstrapClient struct {
	updateURL string
	token     string
	email     string
	pass      string
	sdk       mfsdk.SDK
	mu        sync.Mutex
}

// New returns new Bootstrap service client
func New(updateURL, email, pass string, sdk mfsdk.SDK) certs.BootstrapClient {
	token := os.Getenv("MF_USERS_TOKEN")

	return &bootstrapClient{
		updateURL: updateURL,
		email:     email,
		pass:      pass,
		sdk:       sdk,
		token:     token,
		mu:        sync.Mutex{},
	}
}

func (c *bootstrapClient) fetchToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.token
}

func (c *bootstrapClient) login() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	user := mfsdk.User{Email: c.email, Password: c.pass}
	token, err := c.sdk.CreateToken(user)
	if err != nil {
		return "", err
	}
	c.token = token

	return c.token, nil
}

func (c *bootstrapClient) UpdateCerts(ctx context.Context, thingID, clientCert, clientKey, caCert string) error {
	url := fmt.Sprintf("%s/%s", c.updateURL, thingID)
	r := cert{
		ClientCert: clientCert,
		ClientKey:  clientKey,
		CACert:     caCert,
	}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("%s %s", authHeaderPrefix, c.fetchToken()),
	}

	res, err := request(ctx, http.MethodPatch, url, data, headers)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	bsResponseErrorType(res)
	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusForbidden, http.StatusUnauthorized:
		token, err := c.login()
		if err != nil {
			return errors.Wrap(ErrFailedToLogin, err)
		}
		headers["Authorization"] = fmt.Sprintf("%s %s", authHeaderPrefix, token)
	case http.StatusNotFound:
		return bsResponseErrorType(res)
	default:
		return errors.Wrap(ErrUnexpectedBSResponse, errDetailsBSResp(res))
	}

	res1, err := request(ctx, http.MethodPatch, url, data, headers)
	if err != nil {
		return err
	}

	defer res1.Body.Close()

	switch res1.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusForbidden, http.StatusUnauthorized:
		return errors.Wrap(ErrUnauthorizedAccess, errDetailsBSResp(res1))
	case http.StatusNotFound:
		return bsResponseErrorType(res)
	default:
		return errors.Wrap(ErrUnexpectedBSResponse, errDetailsBSResp(res1))
	}
}

func errDetailsBSResp(res *http.Response) error {
	err := fmt.Errorf("Bootstrap response http status code %d", res.StatusCode)
	b, bErr := io.ReadAll(res.Body)
	if bErr != nil {
		err = errors.Wrap(err, ErrFailedToReadResponseBody)
		err = errors.Wrap(err, bErr)
	}
	err = fmt.Errorf("%w, response body: %s", err, b)
	return err

}

func bsResponseErrorType(res *http.Response) error {
	b, bErr := io.ReadAll(res.Body)
	if bErr != nil {
		return errors.Wrap(ErrFailedToReadResponseBody, bErr)
	}
	var content map[string]string
	err := json.Unmarshal(b, &content)
	if err != nil {
		return errors.ErrNotFound
	}
	if msg, ok := content["error"]; ok {
		return errors.New(msg)
	}
	return errors.New(string(b))
}
func request(ctx context.Context, method, url string, data []byte, header map[string]string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(data))
	req = req.WithContext(ctx)

	if err != nil {
		return nil, err
	}

	for k, v := range header {
		req.Header.Add(k, v)
	}
	defer req.Body.Close()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

type cert struct {
	ClientCert string `json:"client_cert"`
	ClientKey  string `json:"client_key"`
	CACert     string `json:"ca_cert"`
}
