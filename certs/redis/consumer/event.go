// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package consumer

type removeEvent struct {
	id string
}

type updateChannelEvent struct {
	id       string
	name     string
	metadata map[string]interface{}
}
