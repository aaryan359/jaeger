// Copyright (c) 2025 The Jaeger Authors.
// SPDX-License-Identifier: Apache-2.0

package tenancy

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReceiverMiddleware_ExtractTenantFromRequest_Disabled(t *testing.T) {
	tm := NewManager(&Options{Enabled: false})
	rm := NewReceiverMiddleware(tm)

	req := httptest.NewRequest(http.MethodPost, "/v1/traces", http.NoBody)
	tenant, err := rm.ExtractTenantFromRequest(req)

	require.NoError(t, err)
	assert.Empty(t, tenant)
}

func TestReceiverMiddleware_ExtractTenantFromRequest_MissingHeader(t *testing.T) {
	tm := NewManager(&Options{Enabled: true})
	rm := NewReceiverMiddleware(tm)

	req := httptest.NewRequest(http.MethodPost, "/v1/traces", http.NoBody)
	tenant, err := rm.ExtractTenantFromRequest(req)

	require.Error(t, err)
	assert.Empty(t, tenant)
	assert.Equal(t, "missing tenant header", err.Error())
	var tenantErr *TenantError
	require.ErrorAs(t, err, &tenantErr)
	assert.Equal(t, http.StatusUnauthorized, tenantErr.StatusCode())
}

func TestReceiverMiddleware_ExtractTenantFromRequest_ValidTenant(t *testing.T) {
	tm := NewManager(&Options{
		Enabled: true,
		Header:  "x-tenant",
	})
	rm := NewReceiverMiddleware(tm)

	req := httptest.NewRequest(http.MethodPost, "/v1/traces", http.NoBody)
	req.Header.Set("x-tenant", "acme")
	tenant, err := rm.ExtractTenantFromRequest(req)

	require.NoError(t, err)
	assert.Equal(t, "acme", tenant)
}

func TestReceiverMiddleware_ExtractTenantFromRequest_InvalidTenant(t *testing.T) {
	tm := NewManager(&Options{
		Enabled: true,
		Header:  "x-tenant",
		Tenants: []string{"acme", "megacorp"},
	})
	rm := NewReceiverMiddleware(tm)

	req := httptest.NewRequest(http.MethodPost, "/v1/traces", http.NoBody)
	req.Header.Set("x-tenant", "invalid")
	tenant, err := rm.ExtractTenantFromRequest(req)

	require.Error(t, err)
	assert.Empty(t, tenant)
	assert.Equal(t, "unknown tenant", err.Error())
	var tenantErr *TenantError
	require.ErrorAs(t, err, &tenantErr)
	assert.Equal(t, http.StatusUnauthorized, tenantErr.StatusCode())
}

func TestTenantError(t *testing.T) {
	err := NewTenantError("test error", http.StatusForbidden)

	assert.Equal(t, "test error", err.Error())
	assert.Equal(t, http.StatusForbidden, err.StatusCode())
}
