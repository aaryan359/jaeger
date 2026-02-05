// Copyright (c) 2025 The Jaeger Authors.
// SPDX-License-Identifier: Apache-2.0

package tenancy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type mockConsumer struct {
	lastTenant string
	lastCtx    context.Context
}

func (m *mockConsumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	m.lastCtx = ctx
	m.lastTenant = GetTenant(ctx)
	return nil
}

func (m *mockConsumer) ConsumeMetrics(ctx context.Context, md interface{}) error {
	m.lastCtx = ctx
	m.lastTenant = GetTenant(ctx)
	return nil
}

func (m *mockConsumer) ConsumeLogs(ctx context.Context, ld interface{}) error {
	m.lastCtx = ctx
	m.lastTenant = GetTenant(ctx)
	return nil
}

func (m *mockConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{}
}

func TestReceiverMiddleware_ExtractTenantFromRequest_Disabled(t *testing.T) {
	tm := NewManager(&Options{Enabled: false})
	mock := &mockConsumer{}
	rm := NewReceiverMiddleware(tm, mock)

	req := httptest.NewRequest("POST", "/v1/traces", nil)
	tenant, err := rm.ExtractTenantFromRequest(req)

	assert.NoError(t, err)
	assert.Empty(t, tenant)
}

func TestReceiverMiddleware_ExtractTenantFromRequest_MissingHeader(t *testing.T) {
	tm := NewManager(&Options{Enabled: true})
	mock := &mockConsumer{}
	rm := NewReceiverMiddleware(tm, mock)

	req := httptest.NewRequest("POST", "/v1/traces", nil)
	tenant, err := rm.ExtractTenantFromRequest(req)

	require.Error(t, err)
	assert.Empty(t, tenant)
	assert.Equal(t, "missing tenant header", err.Error())
	tenantErr, ok := err.(*TenantError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, tenantErr.StatusCode())
}

func TestReceiverMiddleware_ExtractTenantFromRequest_ValidTenant(t *testing.T) {
	tm := NewManager(&Options{
		Enabled: true,
		Header:  "x-tenant",
	})
	mock := &mockConsumer{}
	rm := NewReceiverMiddleware(tm, mock)

	req := httptest.NewRequest("POST", "/v1/traces", nil)
	req.Header.Set("x-tenant", "acme")
	tenant, err := rm.ExtractTenantFromRequest(req)

	assert.NoError(t, err)
	assert.Equal(t, "acme", tenant)
}

func TestReceiverMiddleware_ExtractTenantFromRequest_InvalidTenant(t *testing.T) {
	tm := NewManager(&Options{
		Enabled: true,
		Header:  "x-tenant",
		Tenants: []string{"acme", "megacorp"},
	})
	mock := &mockConsumer{}
	rm := NewReceiverMiddleware(tm, mock)

	req := httptest.NewRequest("POST", "/v1/traces", nil)
	req.Header.Set("x-tenant", "invalid")
	tenant, err := rm.ExtractTenantFromRequest(req)

	require.Error(t, err)
	assert.Empty(t, tenant)
	assert.Equal(t, "unknown tenant", err.Error())
	tenantErr, ok := err.(*TenantError)
	require.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, tenantErr.StatusCode())
}

func TestReceiverMiddleware_ConsumeTracesWithTenant(t *testing.T) {
	tm := NewManager(&Options{Enabled: true})
	mock := &mockConsumer{}
	rm := NewReceiverMiddleware(tm, mock)

	ctx := context.Background()
	td := ptrace.NewTraces()

	err := rm.ConsumeTracesWithTenant(ctx, "acme", td)

	assert.NoError(t, err)
	assert.Equal(t, "acme", mock.lastTenant)
	assert.Equal(t, "acme", GetTenant(mock.lastCtx))
}

func TestReceiverMiddleware_ConsumeTracesWithTenant_EmptyTenant(t *testing.T) {
	tm := NewManager(&Options{Enabled: false})
	mock := &mockConsumer{}
	rm := NewReceiverMiddleware(tm, mock)

	ctx := context.Background()
	td := ptrace.NewTraces()

	err := rm.ConsumeTracesWithTenant(ctx, "", td)

	assert.NoError(t, err)
	assert.Empty(t, mock.lastTenant)
}

func TestTenantError(t *testing.T) {
	err := NewTenantError("test error", http.StatusForbidden)

	assert.Equal(t, "test error", err.Error())
	assert.Equal(t, http.StatusForbidden, err.StatusCode())
}
