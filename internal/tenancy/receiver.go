// Copyright (c) 2025 The Jaeger Authors.
// SPDX-License-Identifier: Apache-2.0

package tenancy

import (
	"net/http"
)

// ReceiverMiddleware provides utilities for extracting and validating tenant from HTTP requests.
// This middleware should be applied to receivers that accept HTTP requests.
type ReceiverMiddleware struct {
	tm *Manager
}

// NewReceiverMiddleware creates a new receiver middleware that enforces tenancy.
func NewReceiverMiddleware(tm *Manager) *ReceiverMiddleware {
	return &ReceiverMiddleware{
		tm: tm,
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
