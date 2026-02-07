# Unit Test Architecture Overview

## Project Structure
```
hue-lighter/
├── internal/testutils/helpers.go          # Shared testing utilities (COMPLETE)
├── internal/config/
│   ├── validate_test.go                   # ✅ Config validation tests
│   └── load_test.go                       # ✅ Config loading tests
├── internal/hue_client/
│   ├── client_test.go                     # ✅ HTTP client tests
│   ├── api_key_test.go                    # ✅ API key store tests
│   ├── tls_test.go                        # ❌ TLS configuration tests
│   ├── light_test.go                      # ❌ Light control API tests
│   └── discovery_test.go                  # ❌ Bridge discovery tests
├── internal/services/
│   ├── device_registration/service_test.go    # ❌ Registration service tests
│   ├── light_automation/service_test.go       # ❌ Automation service tests  
│   └── events/event-bus_test.go               # ❌ Event system tests
├── internal/app/
│   ├── bootstrap_test.go                      # ❌ App bootstrap tests
│   └── app_test.go                            # ❌ App lifecycle tests
├── internal/logging/logger_test.go             # ❌ Logger tests
├── internal/sunset/sunset_test.go              # ❌ Sunset calculation tests
└── unit-tests/                                # Documentation
    ├── planning.md                             # Detailed implementation plan
    ├── progress.md                             # Progress tracking
    └── architecture.md                         # This file
```

## Testing Foundation

### Core Dependencies
- **testify v1.9.0**: Assertions, require, suite patterns
- **afero v1.15.0**: File system abstraction and mocking
- **httptest**: Standard library HTTP server mocking
- **Go standard testing**: Table-driven test patterns

### Test Utilities Package

Location: `internal/testutils/helpers.go`

#### HTTP Mocking
```go
// Create mock HTTP responses for Hue Bridge API
func MockHueBridgeResponse(statusCode int, body string) *httptest.Server

// Create generic HTTP responses  
func MockHTTPResponse(handler http.HandlerFunc) *httptest.Server
```

#### File System Mocking
```go
// Create temporary directory with afero.MemMapFs
func CreateTempDir() (afero.Fs, string)

// Create file with content in memory filesystem
func CreateFileWithContent(fs afero.Fs, path, content string) error
```

#### Configuration Helpers
```go
// Valid YAML configuration for testing
func ValidHueConfigYAML() string

// Invalid YAML configurations for error testing
func InvalidHueConfigYAML() string
```

#### Environment Management
```go
// Set environment variable with automatic cleanup
func SetEnv(t *testing.T, key, value string)
```

#### Time Mocking
```go
// Interface for time-dependent testing
type TimeProvider interface {
    Now() time.Time
}

// Fixed time implementation
func NewFixedTimeProvider(fixedTime time.Time) TimeProvider
```

#### Assertion Helpers
```go
// Assert error contains specific message
func AssertErrorContains(t *testing.T, err error, message string)

// Assert no error occurred
func AssertNoError(t *testing.T, err error)
```

## Established Testing Patterns

### 1. Table-Driven Tests (Primary Pattern)

**When to Use**: All validation logic, multiple scenario testing, error condition handling

**Example Structure**:
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name        string
        input       InputType
        expected    OutputType
        wantErr     bool
        errorMsg    string
    }{
        {
            name: "valid input case",
            input: validInput,
            expected: expectedOutput,
            wantErr: false,
        },
        {
            name: "invalid input case", 
            input: invalidInput,
            wantErr: true,
            errorMsg: "expected error message",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionUnderTest(tt.input)
            
            if tt.wantErr {
                require.Error(t, err)
                if tt.errorMsg != "" {
                    assert.Contains(t, err.Error(), tt.errorMsg)
                }
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### 2. HTTP API Testing Pattern

**When to Use**: Testing Hue Bridge API interactions, HTTP client functionality

**Example Structure**:
```go
func TestHTTPFunction(t *testing.T) {
    // Setup mock server
    server := testutils.MockHueBridgeResponse(200, `{
        "data": [{"id": "light1", "type": "light"}]
    }`)
    defer server.Close()
    
    // Create client with mock server URL
    client := &HueClient{
        BaseURL: server.URL,
        APIKey: "test-key",
    }
    
    // Execute function
    result, err := client.GetLights()
    
    // Assertions
    require.NoError(t, err)
    assert.Len(t, result, 1)
    assert.Equal(t, "light1", result[0].ID)
}
```

### 3. File System Testing Pattern

**When to Use**: Configuration loading, certificate handling, file persistence

**Example Structure**:
```go
func TestFileFunction(t *testing.T) {
    // Create memory filesystem
    fs, tempDir := testutils.CreateTempDir()
    
    // Setup test files
    configPath := filepath.Join(tempDir, "config.yaml")
    err := testutils.CreateFileWithContent(fs, configPath, testutils.ValidHueConfigYAML())
    require.NoError(t, err)
    
    // Execute function with filesystem
    result, err := LoadConfigFromPath(fs, configPath)
    
    // Assertions
    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 4. Environment Testing Pattern

**When to Use**: Testing environment variable behavior, configuration overrides

**Example Structure**:
```go
func TestEnvironmentFunction(t *testing.T) {
    // Set environment variables with auto-cleanup
    testutils.SetEnv(t, "HUE_BRIDGE_IP", "192.168.1.100") 
    testutils.SetEnv(t, "LOG_LEVEL", "debug")
    
    // Execute function that reads environment
    config, err := LoadFromEnvironment()
    
    // Assertions
    require.NoError(t, err)
    assert.Equal(t, "192.168.1.100", config.BridgeIP)
}
```

### 5. Time-Dependent Testing Pattern

**When to Use**: Automation services, timestamp handling, time-based triggers

**Example Structure**:
```go
func TestTimeFunction(t *testing.T) {
    // Fixed time for deterministic testing
    fixedTime := time.Date(2026, 2, 6, 18, 30, 0, 0, time.UTC)
    timeProvider := testutils.NewFixedTimeProvider(fixedTime)
    
    // Inject time provider into service
    service := &AutomationService{
        timeProvider: timeProvider,
    }
    
    // Execute time-dependent function
    result := service.ShouldTrigger()
    
    // Assertions based on fixed time
    assert.True(t, result)
}
```

## Code Coverage Strategy

### Current Coverage Status
- `internal/config`: 92.3% ✅ (Target achieved)
- `internal/hue_client`: 35.6% 🔄 (In progress)
- All other packages: 0% ❌ (Pending)

### Coverage Targets by Package
1. **Critical Logic (80%+ coverage)**:
   - `internal/hue_client`: Core API functionality
   - `internal/config`: Configuration validation (achieved)
   - `internal/services/light_automation`: Business logic

2. **Important Logic (70%+ coverage)**:
   - `internal/app`: Application lifecycle
   - `internal/services/device_registration`: Registration flow
   - `internal/services/events`: Event communication

3. **Utility Logic (85%+ coverage)**:
   - `internal/logging`: Simple configuration
   - `internal/sunset`: Mathematical calculations

### Coverage Measurement
```bash
# Current coverage check
make test-coverage

# Detailed package coverage
go test ./internal/config/... -coverprofile=config.out
go tool cover -func=config.out

# HTML coverage report
go tool cover -html=config.out -o coverage.html
```

## Error Testing Strategy

### Error Categories to Test

1. **Network Errors**: Connection failures, timeouts, invalid responses
2. **File System Errors**: Missing files, permission errors, invalid content
3. **Validation Errors**: Invalid configurations, out-of-range values
4. **API Errors**: HTTP error codes, malformed JSON responses
5. **Concurrency Errors**: Race conditions, deadlocks (use `-race` flag)

### Error Testing Pattern
```go
func TestErrorConditions(t *testing.T) {
    tests := []struct {
        name string
        setupFunc func() // Setup error condition
        wantErrType error
        wantErrMessage string
    }{
        {
            name: "network timeout",
            setupFunc: func() { /* setup timeout condition */ },
            wantErrType: &net.OpError{},
            wantErrMessage: "timeout",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setupFunc()
            
            _, err := FunctionUnderTest()
            
            require.Error(t, err)
            assert.IsType(t, tt.wantErrType, err)
            assert.Contains(t, err.Error(), tt.wantErrMessage)
        })
    }
}
```

## Mock Strategy Guidelines

### HTTP Mocking
- Use `httptest.Server` for external API calls
- Prefer `testutils.MockHueBridgeResponse()` for Hue API
- Test both success and error response codes
- Verify request headers, method, and body when relevant

### File System Mocking  
- Use `afero.MemMapFs()` through testutils
- Create realistic directory structures 
- Test file permissions and access patterns
- Avoid touching real file system in unit tests

### Time Mocking
- Use `TimeProvider` interface for injection
- Test specific moments in time (sunset, sunrise)
- Test time zone handling and daylight saving transitions
- Avoid `time.Sleep()` in tests - use mocked time advancement

### External Dependencies
- Create interfaces for testability
- Use `testify/mock` for behavior verification  
- Mock at the boundary of your service
- Prefer dependency injection over global state

## Test Organization Best Practices

### File Naming Convention
- Test files: `*_test.go` in same package
- Match source file names: `client.go` → `client_test.go`
- Keep tests close to implementation

### Test Function Naming
```go
// Good: Descriptive test names
func TestValidateConfig_WithValidInput_ReturnsNoError(t *testing.T)
func TestHueClient_GetLights_WhenBridgeReturnsError_ReturnsError(t *testing.T)

// Avoid: Vague test names  
func TestValidate(t *testing.T)
func TestGetLights(t *testing.T)
```

### Test Structure (AAA Pattern)
```go
func TestFunction(t *testing.T) {
    // Arrange: Setup test data and mocks
    input := TestInput{Value: "test"}
    mockServer := testutils.MockHTTPResponse(/*...*/)
    defer mockServer.Close()
    
    // Act: Execute the function under test
    result, err := FunctionUnderTest(input)
    
    // Assert: Verify results and behavior
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

## Performance Testing Guidelines

### Race Condition Detection
```bash
# Always run race detection
make test-race

# Run with high parallel count for stress testing
go test -race -count=100 ./internal/services/... 
```

### Benchmark Tests (Optional)
```go
func BenchmarkCriticalFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        FunctionUnderTest(input)
    }
}
```

## Integration Testing Considerations

**Note**: Current focus is unit testing. Consider these for future integration testing:

### End-to-End Flows
- Bridge discovery → Registration → Light control
- Configuration loading → Service startup → Automation trigger
- Error recovery and retry mechanisms

### Integration Test Structure  
```go
// Future integration test example
func TestE2E_AutomationFlow(t *testing.T) {
    // Setup real bridge or comprehensive mock
    // Load actual configuration
    // Start services
    // Trigger automation
    // Verify expected behavior
}
```

## Troubleshooting Common Issues

### Import Cycles
- Keep test utilities in separate package
- Use interfaces to break dependencies
- Mock at service boundaries

### Flaky Tests
- Avoid real time dependencies - use mocked time
- Don't rely on file system state
- Use deterministic test data
- Clean up resources in defer statements

### Slow Tests
- Avoid real network calls
- Use in-memory file systems
- Prefer mocking over real external services
- Keep test scope focused

## Summary

This architecture provides a solid foundation for comprehensive unit testing of the hue-lighter project. The established patterns, utilities, and guidelines should enable consistent, maintainable test development across all remaining packages.

**Key Success Factors:**
1. Leverage existing testutils for consistency
2. Follow table-driven test patterns
3. Mock external dependencies appropriately  
4. Focus on error path testing
5. Maintain high coverage standards
6. Use established patterns from completed tests

The foundation is robust and ready for systematic completion of the remaining test coverage.
