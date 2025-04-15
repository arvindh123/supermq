// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/ws"
	"github.com/go-chi/chi/v5"
)

func handshake(ctx context.Context, svc ws.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := decodeRequest(r)
		if err != nil {
			encodeError(w, err)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logger.Warn(fmt.Sprintf("Failed to upgrade connection to websocket: %s", err.Error()))
			return
		}
		req.conn = conn
		client := ws.NewClient(conn)

		if err := svc.Subscribe(ctx, req.clientKey, req.domainID, req.chanID, req.subtopic, client); err != nil {
			req.conn.Close()
			return
		}

		logger.Debug(fmt.Sprintf("Successfully upgraded communication to WS on channel %s", req.chanID))
	}
}

func decodeRequest(r *http.Request) (connReq, error) {
	authKey := r.Header.Get("Authorization")
	if authKey == "" {
		authKeys := r.URL.Query()["authorization"]
		if len(authKeys) == 0 {
			logger.Debug("Missing authorization key.")
			return connReq{}, errUnauthorizedAccess
		}
		authKey = authKeys[0]
	}

	domainID := chi.URLParam(r, "domainID")
	chanID := chi.URLParam(r, "chanID")

	req := connReq{
		clientKey: authKey,
		chanID:    chanID,
		domainID:  domainID,
	}

	subTopic := chi.URLParam(r, "subTopic")

	if subTopic != "" {
		subTopic, err := messaging.ParseSubtopic(subTopic)
		if err != nil {
			return connReq{}, err
		}
		req.subtopic = subTopic
	}

	return req, nil
}

func encodeError(w http.ResponseWriter, err error) {
	var statusCode int

	switch err {
	case ws.ErrEmptyTopic:
		statusCode = http.StatusBadRequest
	case errUnauthorizedAccess:
		statusCode = http.StatusForbidden
	case errMalformedSubtopic, errors.ErrMalformedEntity:
		statusCode = http.StatusBadRequest
	default:
		statusCode = http.StatusNotFound
	}
	logger.Warn(fmt.Sprintf("Failed to authorize: %s", err.Error()))
	w.WriteHeader(statusCode)
}
