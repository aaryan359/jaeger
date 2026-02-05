# Multi-Tenancy in Jaeger v2

This document describes the multi-tenancy support in Jaeger v2 and how to configure it.

## Overview

Multi-tenancy in Jaeger v2 allows you to isolate traces, metrics, and logs by tenant. Each tenant's data is stored and queried separately, ensuring data isolation and security.

## Architecture

### Components

1. **Storage Extension** (`jaegerstorage`)
   - Centralized configuration for multi-tenancy
   - Manages tenant validation and context propagation
   - Supports all storage backends (Memory, Badger, Elasticsearch, ClickHouse, etc.)

2. **Query Extension** (`jaegerquery`)
   - Enforces tenancy on query requests
   - Validates tenant headers on HTTP and gRPC requests
   - Propagates tenant context to storage

3. **Remote Storage Extension** (`remotestorage`)
   - Enforces tenancy on remote storage requests
   - Validates tenant headers on gRPC requests

4. **Receiver Middleware** (`internal/tenancy/receiver.go`)
   - Extracts tenant from HTTP headers in receivers
   - Validates tenant against allowed list
   - Propagates tenant context through the pipeline

5. **Storage Backends**
   - Memory storage: Full tenant isolation
   - Other backends: Tenant context available for filtering

## Configuration

### Basic Configuration

Enable multi-tenancy in the storage extension:

```yaml
extensions:
  jaegerstorage:
    backends:
      default:
        memory:
          max_traces: 1000000
    multi_tenancy:
      enabled: true
      header: x-tenant
      tenants:
        - acme
        - megacorp
```

### Configuration Options

- `enabled` (bool): Enable/disable multi-tenancy. Default: `false`
- `header` (string): HTTP header name for tenant. Default: `x-tenant`
- `tenants` ([]string): List of allowed tenants. If empty, all tenants are allowed.

### Query Service Configuration

```yaml
extensions:
  jaegerquery:
    multi_tenancy:
      enabled: true
      header: x-tenant
      tenants:
        - acme
        - megacorp
```

### Remote Storage Configuration

```yaml
extensions:
  remotestorage:
    grpc:
      host-port: :17271
    multi_tenancy:
      enabled: true
      header: x-tenant
      tenants:
        - acme
        - megacorp
    storage: default-storage
```

## Usage

### Query with Tenant Header

**HTTP Request:**
```bash
curl -H "x-tenant: acme" http://localhost:16686/api/traces?service=my-service
```

**gRPC Request:**
```go
ctx := metadata.AppendToOutgoingContext(context.Background(), "x-tenant", "acme")
traces, err := client.GetTraces(ctx, &api_v2.GetTraceRequest{...})
```

### Span Ingestion with Tenant Header

**OTLP HTTP:**
```bash
curl -X POST \
  -H "x-tenant: acme" \
  -H "Content-Type: application/protobuf" \
  -d @traces.pb \
  http://localhost:4318/v1/traces
```

**OTLP gRPC:**
```go
ctx := metadata.AppendToOutgoingContext(context.Background(), "x-tenant", "acme")
exporter, _ := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint("localhost:4317"))
```

## Data Isolation

### Query-Side Isolation

- Query requests must include the tenant header
- Requests without the header are rejected with 401 Unauthorized
- Query results are filtered by tenant
- Tenant context is passed to storage backend

### Write-Side Isolation

- Span ingestion requests should include the tenant header
- Tenant context is propagated through the pipeline
- Storage backend stores data in tenant-specific partitions (for memory storage)
- Other backends should implement tenant-aware filtering

### Storage Backend Support

| Backend | Tenant Isolation | Notes |
|---------|------------------|-------|
| Memory | ✅ Full | Spans stored in tenant-specific maps |
| Badger | ⚠️ Partial | Tenant context available, filtering needed |
| Elasticsearch | ⚠️ Partial | Tenant context available, filtering needed |
| ClickHouse | ⚠️ Partial | Tenant context available, filtering needed |
| Cassandra | ⚠️ Partial | Tenant context available, filtering needed |
| gRPC | ✅ Full | Depends on remote storage backend |

## Implementation Details

### Tenant Context Propagation

1. **HTTP Receivers**: Extract tenant from `x-tenant` header
2. **gRPC Receivers**: Extract tenant from metadata
3. **Processors**: Preserve tenant in context
4. **Exporters**: Pass tenant context to storage
5. **Storage**: Filter/partition by tenant

### Tenant Validation

- Tenant header is required when multi-tenancy is enabled
- Tenant must be in the allowed list (if configured)
- Invalid tenants are rejected with 401 Unauthorized

### Error Handling

- Missing tenant header: 401 Unauthorized
- Invalid tenant: 401 Unauthorized
- Tenant not in allowed list: 401 Unauthorized

## Best Practices

1. **Always enable multi-tenancy** if you have multiple tenants
2. **Use a consistent header name** across all components
3. **Restrict allowed tenants** to prevent unauthorized access
4. **Monitor tenant isolation** to ensure data is properly separated
5. **Test tenant isolation** with integration tests

## Troubleshooting

### Spans from different tenants are mixed

- Verify tenant header is being sent with span ingestion requests
- Check that storage backend supports tenant isolation
- Review storage configuration for tenant filtering

### Query returns data from wrong tenant

- Verify tenant header is being sent with query requests
- Check that query service has multi-tenancy enabled
- Review storage backend for tenant filtering

### 401 Unauthorized errors

- Verify tenant header is being sent
- Check that tenant is in the allowed list
- Verify header name matches configuration

## Migration from v1

In v1, tenancy was configured at the collector level. In v2:

1. Move tenancy configuration to storage extension
2. Update query service configuration
3. Update receiver configurations to send tenant headers
4. Test end-to-end tenant isolation

## Future Improvements

- [ ] Automatic tenant header extraction from trace attributes
- [ ] Tenant-aware batch processor
- [ ] Tenant-based rate limiting
- [ ] Tenant-based retention policies
- [ ] Tenant-based sampling strategies
