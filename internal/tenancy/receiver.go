// Copyright (c) 2025 The Jaeger Authors.
// SPDX-License-Identifier: Apache-2.0

package tenancy

import (
	"context"
	"net/http"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// ReceiverMiddleware wraps a consumer to extract and validate tenant from HTTP headers.
// This middleware should be applied to receivers that accept HTTP requests.
type ReceiverMiddleware struct {
	tm       *Manager
	consumer consumer.Consumer
}

// NewReceiverMiddleware creates a new receiver middleware that enforces tenancy.
func NewReceiverMiddleware(tm *Manager, consumer consumer.Consumer) *ReceiverMiddleware {
	return &ReceiverMiddleware{
		tm:       tm,
		consumer: consumer,
	}
}

// ExtractTenantFromRequest extracts the tenant from an HTTP request.
// Returns the tenant string and an error if tenancy is enabled but the header is missing or invalid.
func (rm *ReceiverMiddleware) ExtractTenantFromRequest(r *http.Request) (string, error) {
	if !rm.tm.Enabled {
		return "", nil
	}

	tenant := r.Header.Get(rm.tm.Header)
	if tenant == "" {
		return "", NewTenantError("missing tenant header", http.StatusUnauthorized)
	}

	if !rm.tm.Valid(tenant) {
		return "", NewTenantError("unknown tenant", http.StatusUnauthorized)
	}

	return tenant, nil
}

// TenantError represents a tenancy-related error with an HTTP status code.
type TenantError struct {
	message    string
	statusCode int
}

// NewTenantError creates a new TenantError.
func NewTenantError(message string, statusCode int) *TenantError {
	return &TenantError{
		message:    message,
		statusCode: statusCode,
	}
}

// Error implements the error interface.
func (te *TenantError) Error() string {
	return te.message
}

// StatusCode returns the HTTP status code for this error.
func (te *TenantError) StatusCode() int {
	return te.statusCode
}

// ConsumeTracesWithTenant wraps ConsumeTraces to add tenant context.
func (rm *ReceiverMiddleware) ConsumeTracesWithTenant(ctx context.Context, tenant string, td ptrace.Traces) error {
	if tenant != "" {
		ctx = WithTenant(ctx, tenant)
	}
	return rm.consumer.ConsumeTraces(ctx, td)
}

// ConsumeMetricsWithTenant wraps ConsumeMetrics to add tenant context.
func (rm *ReceiverMiddleware) ConsumeMetricsWithTenant(ctx context.Context, tenant string, md pmetric.Metrics) error {
	if tenant != "" {
		ctx = WithTenant(ctx, tenant)
	}
	return rm.consumer.ConsumeMetrics(ctx, md)
}

// ConsumeLogsWithTenant wraps ConsumeLogs to add tenant context.
func (rm *ReceiverMiddleware) ConsumeLogsWithTenant(ctx context.Context, tenant string, ld plog.Logs) error {
	if tenant != "" {
		ctx = WithTenant(ctx, tenant)
	}
	return rm.consumer.ConsumeLogs(ctx, ld)
}
