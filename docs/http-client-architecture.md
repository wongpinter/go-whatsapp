# HTTP Client Architecture Analysis & Recommendations

## Current State Analysis

### Issues Identified

1. **Code Duplication**: Each package (`cloudapi`, `bm`, `flows`) initializes its own Resty client with similar configurations
2. **Resource Inefficiency**: Multiple HTTP clients for the same API endpoint create unnecessary connection pools
3. **Inconsistent Configuration**: Different packages use different API versions and retry settings
4. **Incomplete Implementation**: `WithHTTPClient` option was not fully implemented
5. **No Connection Pooling Strategy**: Each client manages its own connections independently

### Current Implementation Pattern

```go
// Each package does this independently:
c.restyClient = resty.New().
    SetBaseURL("https://graph.facebook.com/v19.0").
    SetAuthToken(c.accessToken).
    SetHeader("Content-Type", "application/json").
    SetRetryCount(3).
    SetRetryWaitTime(1 * time.Second).
    SetRetryMaxWaitTime(10 * time.Second)
```

## Recommended Solution

### 1. Shared HTTP Client Factory (`internal/httpclient/factory.go`)

**Purpose**: Centralized HTTP client creation with consistent configuration

**Key Features**:
- Standardized client configuration across all packages
- Connection pooling optimization
- Configurable retry strategies
- Built-in rate limiting support
- Comprehensive logging and error handling

**Benefits**:
- Eliminates code duplication
- Ensures consistent HTTP client behavior
- Optimizes resource usage through connection pooling
- Provides centralized configuration management

### 2. HTTP Client Manager (`internal/httpclient/manager.go`)

**Purpose**: Manages and reuses HTTP clients to avoid creating multiple instances

**Key Features**:
- Client caching and reuse based on configuration
- Thread-safe client management
- Automatic cleanup and resource management
- Support for different client types (CloudAPI, Business, Flows)

**Benefits**:
- Reduces memory usage by reusing clients
- Improves performance through client caching
- Prevents connection pool exhaustion
- Simplifies client lifecycle management

### 3. Client Type Differentiation

```go
type ClientType string

const (
    CloudAPIClient    ClientType = "cloudapi"
    BusinessAPIClient ClientType = "business"
    FlowsAPIClient    ClientType = "flows"
)
```

**Purpose**: Allows different configurations for different API types while maintaining efficiency

## Implementation Strategy

### Phase 1: Core Infrastructure ✅

1. **Created `internal/httpclient/factory.go`**:
   - HTTP client factory with standardized configuration
   - Support for custom HTTP clients
   - Built-in retry logic and rate limiting
   - Comprehensive logging

2. **Created `internal/httpclient/manager.go`**:
   - Client caching and reuse mechanism
   - Thread-safe operations
   - Resource cleanup capabilities
   - Default configuration helpers

### Phase 2: Main Client Integration ✅

1. **Updated `client.go`**:
   - Integrated HTTP client manager
   - Improved `WithHTTPClient` option implementation
   - Added support for shared client configuration

### Phase 3: Package Client Updates (Recommended Next Steps)

1. **Update `cloudapi/client.go`**:
   ```go
   // Replace direct Resty initialization with factory usage
   manager := httpclient.NewManager(config.DefaultConfig(), logger)
   clientConfig := &httpclient.ClientConfig{
       AccessToken: accessToken,
       APIVersion:  "v19.0",
       Logger:      logger,
   }
   c.restyClient, err = manager.GetOrCreateClient(httpclient.CloudAPIClient, clientConfig)
   ```

2. **Update `bm/client.go`** and `flows/client.go` similarly

## Usage Examples

### Basic Usage with Shared Client

```go
// Main client automatically uses shared HTTP client manager
client, err := whatsapp.NewClient(
    phoneNumberID,
    accessToken,
    whatsapp.WithRateLimiting(80.0),
    whatsapp.WithTimeout(30*time.Second),
)
```

### Advanced Configuration

```go
// Create manager with custom configuration
manager := httpclient.NewManager(config.DefaultConfig(), logger)

// Configure client with advanced options
config := &httpclient.ClientConfig{
    AccessToken:   "token",
    APIVersion:    "v19.0",
    Timeout:       30 * time.Second,
    RetryCount:    5,
    CustomHeaders: map[string]string{
        "X-App-Version": "1.0.0",
    },
}

// Add rate limiting
config = httpclient.WithRateLimiting(config, 100.0, 20)

// Create client
client, err := manager.GetOrCreateClient(httpclient.CloudAPIClient, config)
```

### Client Reuse Demonstration

```go
// First client creation
client1, _ := manager.GetOrCreateClient(httpclient.CloudAPIClient, config)

// Second client with same config - reuses the first one
client2, _ := manager.GetOrCreateClient(httpclient.CloudAPIClient, config)

// client1 == client2 (same instance)
```

## Performance Benefits

### Before (Current State)
- **Memory Usage**: Each package creates separate HTTP clients
- **Connection Pools**: Multiple pools for the same endpoint
- **Configuration**: Duplicated across packages
- **Maintenance**: Changes needed in multiple places

### After (Recommended Implementation)
- **Memory Usage**: Shared clients reduce memory footprint
- **Connection Pools**: Optimized pooling with reuse
- **Configuration**: Centralized and consistent
- **Maintenance**: Single point of configuration

## Migration Strategy

### Immediate Benefits (Already Implemented)
1. Main `whatsapp.Client` now uses shared HTTP client manager
2. Improved `WithHTTPClient` option for custom HTTP clients
3. Foundation for package client updates

### Next Steps (Recommended)
1. Update `cloudapi` package to use shared factory
2. Update `bm` package to use shared factory  
3. Update `flows` package to use shared factory
4. Add comprehensive tests for client reuse
5. Add metrics for monitoring client usage

### Backward Compatibility
- All existing APIs remain unchanged
- Functional options continue to work as expected
- Performance improvements are transparent to users

## Configuration Options

### Default Configuration
```go
cfg := config.DefaultConfig()
// BaseURL: "https://graph.facebook.com"
// APIVersion: "v19.0"
// RequestTimeout: 30s
// RetryCount: 3
// RetryWaitTime: 1s
// RetryMaxWait: 10s
```

### Custom Configuration
```go
config := &httpclient.ClientConfig{
    AccessToken:   "your-token",
    APIVersion:    "v19.0",
    Timeout:       45 * time.Second,
    RetryCount:    5,
    RetryWaitTime: 2 * time.Second,
    RetryMaxWait:  20 * time.Second,
    UserAgent:     "MyApp/1.0.0",
    CustomHeaders: map[string]string{
        "X-Custom-Header": "value",
    },
}
```

## Testing Strategy

### Unit Tests
- Test client factory creation with various configurations
- Test client manager caching and reuse logic
- Test error handling and edge cases

### Integration Tests
- Test actual HTTP requests with shared clients
- Test rate limiting functionality
- Test client cleanup and resource management

### Performance Tests
- Benchmark client creation vs reuse
- Memory usage comparison before/after
- Connection pool efficiency testing

## Monitoring and Observability

### Metrics to Track
- Number of cached clients
- Client reuse rate
- Connection pool utilization
- Request latency and success rates

### Logging
- Client creation and reuse events
- Configuration changes
- Error conditions and retries

## Security Considerations

### Access Token Management
- Tokens are not logged or exposed
- Secure client key generation for caching
- Proper cleanup of sensitive data

### Connection Security
- TLS configuration consistency
- Certificate validation
- Secure header handling

## Conclusion

The recommended HTTP client architecture provides:

1. **Efficiency**: Reduced resource usage through client reuse
2. **Consistency**: Standardized configuration across packages
3. **Maintainability**: Centralized HTTP client management
4. **Performance**: Optimized connection pooling
5. **Flexibility**: Support for custom configurations
6. **Observability**: Built-in logging and monitoring

This architecture maintains backward compatibility while providing significant performance and maintainability improvements.
