// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/mainflux/mainflux"
	bsmocks "github.com/mainflux/mainflux/bootstrap/mocks"
	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/certs/mocks"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/things"
	httpapi "github.com/mainflux/mainflux/things/api/things/http"
	thmocks "github.com/mainflux/mainflux/things/mocks"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	wrongValue = "wrong-value"
	email      = "user@example.com"
	token      = "token"
	thingsNum  = 1
	certID     = "1"
	name       = "certificate name"
	thingKey   = "thingKey"
	thingID    = "1"
	ttl        = "1h"
	keyBits    = 2048
	key        = "rsa"
	certNum    = 10

	cfgLogLevel    = "error"
	cfgClientTLS   = false
	cfgServerCert  = ""
	cfgServerKey   = ""
	cfgCertsURL    = "http://localhost"
	cfgJaegerURL   = ""
	cfgAuthURL     = "localhost:8181"
	cfgAuthTimeout = "1s"

	cfgBootStrapURL = "http://localhost"
	cfgMfUser       = "admin@example.com"
	cfgMfPass       = "12345678"

	caPath            = "../docker/ssl/certs/ca.crt"
	caKeyPath         = "../docker/ssl/certs/ca.key"
	cfgSignHoursValid = "24h"
	cfgSignRSABits    = 2048
)

func newService(tokens map[string]string) (certs.Service, error) {

	policies := []thmocks.MockSubjectSet{{Object: "users", Relation: "member"}}
	auth := thmocks.NewAuthService(tokens, map[string][]thmocks.MockSubjectSet{email: policies})

	repo := mocks.NewCertsRepository()

	tlsCert, caCert, err := loadCertificates(caPath, caKeyPath)
	if err != nil {
		return nil, err
	}

	pki := mocks.NewPkiAgent(tlsCert, caCert, cfgSignRSABits, cfgSignHoursValid)

	idProvider := uuid.NewMock()
	return certs.New(auth, repo, idProvider, pki), nil
}

func newThingsService(auth mainflux.AuthServiceClient) things.Service {
	ths := make(map[string]things.Thing, thingsNum)
	for i := 0; i < thingsNum; i++ {
		id := strconv.Itoa(i + 1)
		ths[id] = things.Thing{
			ID:    id,
			Key:   thingKey,
			Owner: email,
		}
	}

	return bsmocks.NewThingsService(ths, map[string]things.Channel{}, auth)
}

func TestIssueCert(t *testing.T) {
	svc, err := newService(map[string]string{token: email})
	require.Nil(t, err, fmt.Sprintf("unexpected service creation error: %s\n", err))

	cases := []struct {
		name    string
		token   string
		desc    string
		thingID string
		ttl     string
		key     string
		keyBits int
		err     error
	}{
		{
			desc:    "issue new cert",
			name:    name,
			token:   token,
			thingID: thingID,
			ttl:     ttl,
			err:     nil,
		},
		{
			desc:    "issue new cert for invalid token",
			name:    name,
			token:   "",
			thingID: thingID,
			ttl:     ttl,
			err:     certs.ErrThingRetrieve,
		},
		{
			desc:    "issue new cert for non existing thing id",
			name:    name,
			token:   token,
			thingID: "2",
			ttl:     ttl,
			err:     certs.ErrThingRetrieve,
		},
		{
			desc:    "issue new cert for non existing thing id",
			name:    name,
			token:   wrongValue,
			thingID: thingID,
			ttl:     ttl,
			err:     errors.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		c, err := svc.IssueCert(context.Background(), tc.token, tc.thingID, tc.name, tc.ttl)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		cert, _ := readCert([]byte(c.Certificate))
		if cert != nil {
			assert.True(t, strings.Contains(cert.Subject.CommonName, thingKey), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		}
	}

}

func TestRevokeCert(t *testing.T) {
	svc, err := newService(map[string]string{token: email})
	require.Nil(t, err, fmt.Sprintf("unexpected service creation error: %s\n", err))

	_, err = svc.IssueCert(context.Background(), token, thingID, name, ttl)
	require.Nil(t, err, fmt.Sprintf("unexpected service creation error: %s\n", err))

	cases := []struct {
		token string
		desc  string
		id    string
		err   error
	}{
		{
			desc:  "revoke cert",
			token: token,
			id:    certID,
			err:   nil,
		},
		{
			desc:  "revoke cert for invalid token",
			token: wrongValue,
			id:    certID,
			err:   errors.ErrAuthentication,
		},
		{
			desc:  "revoke cert for unknown id",
			token: token,
			id:    "2",
			err:   nil,
		},
	}

	for _, tc := range cases {
		err := svc.RevokeCert(context.Background(), tc.token, tc.id)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestViewCert(t *testing.T) {
	svc, err := newService(map[string]string{token: email})
	require.Nil(t, err, fmt.Sprintf("unexpected service creation error: %s\n", err))

	ic, err := svc.IssueCert(context.Background(), token, thingID, name, ttl, keyBits, key)
	require.Nil(t, err, fmt.Sprintf("unexpected cert creation error: %s\n", err))

	cert := certs.Cert{
		ID:          ic.ID,
		ThingID:     thingID,
		Certificate: ic.Certificate,
		Serial:      ic.Serial,
		Expire:      ic.Expire,
	}

	cases := []struct {
		token string
		desc  string
		id    string
		cert  certs.Cert
		err   error
	}{
		{
			desc:  "list cert with valid token and id",
			token: token,
			id:    cert.ID,
			cert:  cert,
			err:   nil,
		},
		{
			desc:  "list cert with invalid token and id",
			token: wrongValue,
			id:    cert.ID,
			cert:  certs.Cert{},
			err:   errors.ErrAuthentication,
		},
		{
			desc:  "list cert with invalid id",
			token: token,
			id:    cert.ID,
			cert:  certs.Cert{},
			err:   errors.ErrNotFound,
		},
	}

	for _, tc := range cases {
		cert, err := svc.ViewCert(context.Background(), tc.token, tc.id)
		assert.Equal(t, tc.cert, cert, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.cert, cert))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func newThingsServer(svc things.Service) *httptest.Server {
	logger := logger.NewMock()
	mux := httpapi.MakeHandler(mocktracer.New(), svc, logger)
	return httptest.NewServer(mux)
}

func loadCertificates(caPath, caKeyPath string) (tls.Certificate, *x509.Certificate, error) {
	var tlsCert tls.Certificate
	var caCert *x509.Certificate

	if caPath == "" || caKeyPath == "" {
		return tlsCert, caCert, nil
	}

	if _, err := os.Stat(caPath); os.IsNotExist(err) {
		return tlsCert, caCert, err
	}

	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		return tlsCert, caCert, err
	}

	tlsCert, err := tls.LoadX509KeyPair(caPath, caKeyPath)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(err, err)
	}

	b, err := ioutil.ReadFile(caPath)
	if err != nil {
		return tlsCert, caCert, err
	}

	caCert, err = readCert(b)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(err, err)
	}

	return tlsCert, caCert, nil
}

func readCert(b []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("failed to decode PEM data")
	}

	return x509.ParseCertificate(block.Bytes)
}
