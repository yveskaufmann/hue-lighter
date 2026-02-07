# Unit Testing Plan for Hue-Lighter

## Overview

This document provides a comprehensive plan for completing the remaining unit tests for the hue-lighter project. The testing foundation has been established with testify, afero, and custom test utilities. Current coverage: config (92%), hue_client (35.6%).

## Testing Infrastructure (Already Completed ✅)

- **Dependencies**: testify v1.9.0, afero v1.15.0 added to go.mod
- **Test Utilities**: `internal/testutils/helpers.go` with HTTP mocking, file system mocking, environment helpers
- **Makefile Targets**: `test`, `test-coverage`, `test-race`, `test-verbose`
- **Completed Tests**: config package (validate + load), hue_client (client + api_key)

## Remaining Test Implementation Plan

### Phase 1: Complete Hue Client Package (High Priority)

#### 1. Create `internal/hue_client/tls_test.go`
**Complexity**: Medium
**Estimated Time**: 2-3 hours
**Dependencies**: Understanding of TLS certificate verification, x509 package

**Key Functions to Test**:
- `NewBridgeTLSConfig(bridgeID, certPath)` - TLS config creation
- `ResolveCABundlePath()` - CA bundle path resolution
- `createCustomCertVerifier()` - Custom certificate verification logic

**Testing Strategy**:
- Create test CA certificates using crypto/x509/testdata patterns
- Mock file system for certificate files using afero
- Test certificate validation logic with valid/invalid cert chains
- Test environment variable handling for `HUE_CA_CERTS_PATH`
- Test error cases: missing files, invalid certificates, permission errors

**Test Cases**:
```go
func TestNewBridgeTLSConfig(t *testing.T) {
    // Valid certificate bundle
    // Missing certificate file
    // Invalid PEM format
    // System cert pool errors
    // Bridge ID case sensitivity
}

func TestResolveCABundlePath(t *testing.T) {
    // Default path when no env var
    // Custom path from HUE_CA_CERTS_PATH
    // File existence validation
}
```

#### 2. Create `internal/hue_client/light_test.go`
**Complexity**: Medium-High
**Estimated Time**: 3-4 hours
**Dependencies**: Understanding of Hue API responses, light model structures

**Key Functions to Test**:
- `GetAllLights()` - Fetch all lights from bridge
- `TurnOnLightById(lightID, brightness, color)` - Light control
- `TurnOffLightById(lightID)` - Light control
- `GetLightById(lightID)` - Single light retrieval

**Testing Strategy**:
- Use `testutils.MockHueBridgeResponse()` for API responses
- Test with various light states, colors, brightness levels
- Test error responses from Hue API
- Test JSON unmarshaling of complex light model structures
- Verify HTTP request formation (headers, body, method)

**Sample API Responses Needed**:
```json
// Lights collection response
{
  "data": [
    {
      "id": "light-1",
      "type": "light", 
      "on": {"on": true},
      "dimming": {"brightness": 50.0},
      "color": {"xy": {"x": 0.3, "y": 0.3}}
    }
  ]
}

// Error response
{
  "errors": [{
    "description": "resource, /clip/v2/resource/light/invalid, not available"
  }]
}
```

#### 3. Create `internal/hue_client/discovery_test.go`
**Complexity**: High
**Estimated Time**: 4-5 hours
**Dependencies**: mDNS mocking, HTTP discovery mocking

**Key Functions to Test**:
- `DiscoverBridges()` - Network bridge discovery
- mDNS discovery logic
- HTTP-based discovery fallback
- Bridge validation and filtering

**Testing Challenges**:
- mDNS requires mocking `github.com/brutella/dnssd` package
- Network timeouts and error handling
- Multiple discovery methods (mDNS + HTTP)

**Testing Strategy**:
- Mock external discovery API (`discovery.meethue.com`)
- Mock mDNS responses using interface substitution
- Test discovery timeouts and failures
- Test bridge validation logic
- Test discovery result filtering and deduplication

### Phase 2: Services Layer Testing (Medium Priority)

#### 4. Create `internal/services/device_registration/service_test.go`
**Complexity**: Medium
**Estimated Time**: 2-3 hours

**Key Components**:
- Device registration service logic
- Bridge communication for registration
- Device name and type handling
- Registration flow state management

**Testing Strategy**:
- Mock Hue client for registration API calls
- Test registration success/failure flows
- Test device naming and validation
- Test registration timeouts and retries

#### 5. Create `internal/services/light_automation/service_test.go`
**Complexity**: High (Time-dependent logic)
**Estimated Time**: 4-5 hours

**Key Components**:
- Sunrise/sunset calculation integration
- Timer-based automation logic
- Light state management
- Service lifecycle (start/stop)

**Testing Strategy**:
- Mock time using `testutils.FixedTimeProvider`
- Mock sunrise/sunset calculations
- Test automation triggers at specific times
- Test light state caching and updates
- Test service shutdown gracefully

**Time Mocking Example**:
```go
func TestLightAutomation_SunsetTrigger(t *testing.T) {
    fixedTime := time.Date(2026, 2, 6, 18, 30, 0, 0, time.UTC)
    timeProvider := testutils.NewFixedTimeProvider(fixedTime)
    // Test automation triggers at sunset time
}
```

#### 6. Create `internal/services/events/event-bus_test.go`
**Complexity**: Medium-High (Unix sockets)
**Estimated Time**: 3-4 hours

**Key Components**:
- Unix socket communication
- Event broker functionality 
- Inter-process communication
- Event serialization/deserialization

**Testing Strategy**:
- Use temporary unix socket files in `t.TempDir()`
- Test event publishing and subscription
- Test socket creation and cleanup
- Test concurrent event handling
- Test socket permission handling

### Phase 3: Application Layer Testing (Lower Priority)

#### 7. Create `internal/app/bootstrap_test.go`
**Complexity**: Medium
**Estimated Time**: 2-3 hours

**Key Components**:
- Dependency injection setup
- Service initialization order
- Configuration loading integration
- Error handling during bootstrap

**Testing Strategy**:
- Test service dependency resolution
- Test bootstrap with invalid configurations
- Test partial initialization failures
- Mock external dependencies (file system, network)

#### 8. Create `internal/app/app_test.go`
**Complexity**: Medium-High (Signal handling)
**Estimated Time**: 3-4 hours

**Key Components**:
- Application lifecycle management
- Signal handling (SIGTERM, SIGINT)
- Graceful shutdown logic
- Service coordination during shutdown

**Testing Strategy**:
- Mock signal delivery for controlled testing
- Test graceful shutdown timeouts
- Test service cleanup order
- Test application restart scenarios

### Phase 4: Utility Testing (Lowest Priority)

#### 9. Create `internal/logging/logger_test.go`
**Complexity**: Low
**Estimated Time**: 1-2 hours

**Key Components**:
- Logger configuration from environment
- Log level handling
- Log format switching

**Testing Strategy**:
- Use `testutils.SetEnv()` for environment testing
- Test LOG_LEVEL and LOG_FORMAT environment variables
- Capture and verify log output
- Test invalid configuration handling

#### 10. Create `internal/sunset/sunset_test.go`
**Complexity**: Low-Medium
**Estimated Time**: 2 hours

**Key Components**:
- Sunrise/sunset calculation wrapper
- Time zone handling
- Geographic coordinate validation

**Testing Strategy**:
- Test with known sunrise/sunset times for specific dates/locations
- Test edge cases (polar regions, equator)
- Mock time for deterministic testing
- Test coordinate boundary conditions

## Testing Best Practices to Follow

### 1. Table-Driven Tests
Use for all validation logic and multiple scenario testing:
```go
tests := []struct {
    name string
    input InputType
    expected OutputType
    wantErr bool
}{
    // test cases
}
```

### 2. Test Utilities Usage
- Use `testutils.MockHueBridgeResponse()` for HTTP API mocking
- Use `testutils.SetEnv()` for environment variable testing
- Use `testutils.AssertErrorContains()` for error message validation
- Use `testutils.ValidHueConfigYAML()` for valid configuration data

### 3. Mock Strategies
- **HTTP**: Use `httptest.Server` with predefined responses
- **File System**: Use `afero.MemMapFs()` via testutils
- **Time**: Use `testutils.FixedTimeProvider` interface
- **External Dependencies**: Create interface mocks with testify/mock

### 4. Error Testing
- Test happy path first, then error conditions
- Test network failures, timeouts, invalid inputs
- Verify error message content and error type wrapping
- Test error propagation through service layers

### 5. Concurrency Testing
- Use `go test -race` to detect race conditions
- Test service startup/shutdown concurrency
- Test multiple goroutines accessing shared resources

## Integration Testing Considerations

After unit tests are complete, consider adding integration tests for:
- End-to-end bridge registration flow
- Complete automation cycle (sunset trigger → lights on)
- Configuration loading → service startup → automation
- Error recovery and retry mechanisms

## Acceptance Criteria

### Coverage Targets
- **internal/config**: Maintain 92%+ (already achieved)
- **internal/hue_client**: Increase to 80%+ (currently 35.6%)
- **internal/services**: Achieve 75%+ for each service
- **internal/app**: Achieve 70%+ (bootstrap and lifecycle testing)
- **internal/logging**: Achieve 90%+ (simple utility functions)
- **internal/sunset**: Achieve 85%+ (wrapper functions)

### Quality Gates
- All tests must pass with `make test`
- No race conditions detected with `make test-race`  
- Tests must run without external dependencies (no real network, files, or system calls)
- Each test file should have comprehensive table-driven tests
- Error paths must be thoroughly tested

### Documentation Requirements
- Each test file should have package-level documentation
- Complex test setups should have helper function comments
- Mock strategies should be documented within test functions
- Any test-specific configuration should be documented

## Handoff Notes

### Prerequisites
- Golang testing fundamentals
- Understanding of testify assertion library
- Familiarity with HTTP mocking and file system abstraction
- Basic knowledge of Philips Hue API structure

### Getting Started
1. Run `make test-coverage` to see current baseline
2. Review existing test files in `internal/config/` and `internal/hue_client/`
3. Study `internal/testutils/helpers.go` for available utilities
4. Start with Phase 1 (hue_client completion) for highest impact

### Support Resources
- Existing test files serve as patterns and examples
- `testutils` package provides reusable mocking infrastructure
- Hue API documentation: https://developers.meethue.com/develop/hue-api-v2/
- Testify documentation: https://pkg.go.dev/github.com/stretchr/testify

### Expected Timeline
- **Phase 1** (hue_client): 8-12 hours
- **Phase 2** (services): 9-12 hours  
- **Phase 3** (app layer): 5-7 hours
- **Phase 4** (utilities): 3-4 hours
- **Total**: 25-35 hours of development time

This plan provides a structured approach to achieving comprehensive test coverage for the hue-lighter project while maintaining code quality and following Go testing best practices.
