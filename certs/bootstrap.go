// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package certs

import "context"

// Bootstrap client struct
type BootstrapClient interface {
	UpdateCerts(context.Context, string, string, string, string) error
}
