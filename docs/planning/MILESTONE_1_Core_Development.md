# Milestone 1: Core Development
Status: Not Started

This milestone establishes the foundation of the Kubernetes mutating admission controller for automatic pod labeling. It includes project setup, core webhook server implementation, admission request handling, label management logic, certificate management, and basic testing framework.

## Success Criteria
- Go project properly initialized with module structure
- Basic HTTPS webhook server running and accepting admission requests
- Admission handler successfully processing and mutating pod specs
- Label generation working with basic rules
- TLS certificates properly handled
- Unit tests achieving >70% coverage for core components

## Task 1: Project Setup and Structure
Status: Completed

Initialize the Go project with proper module structure, development environment setup, and initial configuration files.

### Success Criteria
- Go module initialized with correct dependencies
- Project directory structure created following Go best practices
- Development tools configured (linting, formatting, etc.)
- Initial Makefile with common commands

### Subtask 1.1: Initialize Go Module
Status: Completed

Create Go module and basic project structure.

Commit SHA: (pending commit) 

#### TDD Test Cases
- Test that go.mod file exists with correct module name
- Test that required directories exist (cmd/, pkg/, config/, deploy/, test/)
- Test that .gitignore properly excludes binaries and temporary files

#### Success Criteria
- go.mod created with module github.com/jjshanks/pod-labeler
- Basic directory structure established
- .gitignore configured for Go projects

### Subtask 1.2: Add Core Dependencies
Status: Completed

Add essential Kubernetes and controller-runtime dependencies to go.mod.

Commit SHA: (pending commit) 

#### TDD Test Cases
- Test that go.mod contains k8s.io/api dependency
- Test that go.mod contains k8s.io/apimachinery dependency
- Test that go.mod contains sigs.k8s.io/controller-runtime dependency

#### Success Criteria
- Core Kubernetes dependencies added
- controller-runtime dependency added
- go.sum file generated

### Subtask 1.3: Create Makefile
Status: Completed

Create Makefile with common development commands.

Commit SHA: (pending commit) 

#### TDD Test Cases
- Test that make build creates binary in bin/ directory
- Test that make test runs go test
- Test that make lint runs golangci-lint

#### Success Criteria
- Makefile created with build, test, lint, and clean targets
- Binary output configured to bin/ directory
- Formatting and linting commands included

### Subtask 1.4: Setup Development Configuration
Status: Completed

Create development configuration files and tool settings.

Commit SHA: (pending commit) 

#### TDD Test Cases
- Test that .golangci.yml exists with proper linter configuration
- Test that .editorconfig exists with Go formatting rules
- Test that development config directory exists

#### Success Criteria
- .golangci.yml configured for project standards
- .editorconfig set up for consistent formatting
- Development scripts directory created

## Task 2: Core Webhook Server Implementation
Status: Not Started

Implement the basic HTTPS webhook server with proper TLS configuration and request routing.

### Success Criteria
- HTTPS server starting and listening on configured port
- TLS properly configured with certificate loading
- Basic request routing implemented
- Graceful shutdown handling

### Subtask 2.1: Create Main Entry Point
Status: Not Started

Create cmd/webhook/main.go with basic application structure.

Commit SHA: 

#### TDD Test Cases
- Test that main function initializes without panic
- Test that command-line flags are properly defined
- Test that logger is initialized correctly

#### Success Criteria
- main.go created with proper package structure
- Command-line flag parsing implemented
- Basic logging setup completed

### Subtask 2.2: Implement HTTP Server Structure
Status: Not Started

Create pkg/webhook/server.go with WebhookServer struct and methods.

Commit SHA: 

#### TDD Test Cases
- Test WebhookServer struct creation
- Test that server configuration is properly set
- Test that server methods exist (Start, Stop)

#### Success Criteria
- WebhookServer struct defined with necessary fields
- Constructor function implemented
- Basic server lifecycle methods created

### Subtask 2.3: Add TLS Configuration
Status: Not Started

Implement TLS configuration for HTTPS server.

Commit SHA: 

#### TDD Test Cases
- Test TLS config creation with proper cipher suites
- Test minimum TLS version is set to 1.2
- Test certificate loading from file paths

#### Success Criteria
- TLS configuration with secure defaults
- Certificate and key loading implemented
- TLS 1.2+ enforced

### Subtask 2.4: Implement Server Start Method
Status: Not Started

Create server Start method with proper error handling.

Commit SHA: 

#### TDD Test Cases
- Test server starts on configured port
- Test server returns error on invalid port
- Test server handles TLS errors properly

#### Success Criteria
- Start method creates and configures http.Server
- Proper error handling for startup failures
- Server listens on configured address

### Subtask 2.5: Add Graceful Shutdown
Status: Not Started

Implement graceful shutdown handling for the server.

Commit SHA: 

#### TDD Test Cases
- Test server shuts down on SIGTERM
- Test server waits for active connections
- Test shutdown timeout is respected

#### Success Criteria
- Signal handling for SIGTERM and SIGINT
- Graceful shutdown with configurable timeout
- Proper cleanup of resources

### Subtask 2.6: Create Health Check Endpoints
Status: Not Started

Add /healthz and /readyz endpoints for Kubernetes probes.

Commit SHA: 

#### TDD Test Cases
- Test /healthz returns 200 OK
- Test /readyz returns 200 when ready
- Test /readyz returns 503 when not ready

#### Success Criteria
- Health check endpoints responding correctly
- Readiness logic implemented
- Proper HTTP status codes returned

## Task 3: Admission Handler Development
Status: Not Started

Implement the core admission webhook handler that processes admission requests and generates responses.

### Success Criteria
- Admission requests properly decoded and validated
- Handler correctly processes CREATE and UPDATE operations
- Responses include proper patches when mutations occur
- Error handling returns appropriate admission responses

### Subtask 3.1: Create Handler Interface
Status: Not Started

Define admission handler interface and basic structure.

Commit SHA: 

#### TDD Test Cases
- Test Handler interface has Handle method
- Test PodLabelMutator implements Handler interface
- Test handler struct has required dependencies

#### Success Criteria
- Handler interface defined in pkg/webhook/handler.go
- PodLabelMutator struct created
- Constructor function implemented

### Subtask 3.2: Implement Request Decoding
Status: Not Started

Add admission request decoding logic.

Commit SHA: 

#### TDD Test Cases
- Test successful decoding of valid AdmissionReview
- Test error handling for malformed requests
- Test pod extraction from request object

#### Success Criteria
- AdmissionReview v1 decoding implemented
- Pod object extraction working
- Proper error responses for invalid requests

### Subtask 3.3: Add Operation Filtering
Status: Not Started

Filter requests to only process CREATE and UPDATE operations.

Commit SHA: 

#### TDD Test Cases
- Test CREATE operations are processed
- Test UPDATE operations are processed
- Test DELETE operations are allowed without processing

#### Success Criteria
- Operation type checking implemented
- Non-mutation operations return early
- Proper allowed responses for skipped operations

### Subtask 3.4: Implement System Namespace Skip
Status: Not Started

Add logic to skip system namespaces from mutation.

Commit SHA: 

#### TDD Test Cases
- Test kube-system namespace is skipped
- Test kube-public namespace is skipped
- Test kube-node-lease namespace is skipped

#### Success Criteria
- System namespace list defined
- Skip logic implemented in handler
- Allowed response returned for system namespaces

### Subtask 3.5: Create Basic Response Builder
Status: Not Started

Implement admission response construction.

Commit SHA: 

#### TDD Test Cases
- Test allowed response creation
- Test patch response creation
- Test error response creation

#### Success Criteria
- Response builder methods created
- Proper UID handling from request
- AdmissionResponse struct properly populated

## Task 4: Label Management System
Status: Not Started

Design and implement the label generation and management system with validation and rules engine.

### Success Criteria
- Label generator interface defined and implemented
- Dynamic label generation based on pod/namespace context
- Label validation ensures Kubernetes constraints
- Extensible rules system for future enhancements

### Subtask 4.1: Define Label Generator Interface
Status: Not Started

Create label generator interface and types in pkg/labels/generator.go.

Commit SHA: 

#### TDD Test Cases
- Test Generator interface has GenerateLabels method
- Test LabelRule interface is defined
- Test label map type handling

#### Success Criteria
- Generator interface with context-aware generation
- LabelRule interface for extensibility
- Type definitions for label operations

### Subtask 4.2: Implement Label Validation
Status: Not Started

Add label key/value validation per Kubernetes specs.

Commit SHA: 

#### TDD Test Cases
- Test valid label keys are accepted
- Test invalid label keys are rejected (>63 chars, invalid chars)
- Test label values validation (empty allowed, >63 chars rejected)

#### Success Criteria
- Label key regex validation implemented
- Label value validation implemented
- Error messages clearly indicate validation failures

### Subtask 4.3: Create Static Label Generator
Status: Not Started

Implement basic static label generator for managed-by labels.

Commit SHA: 

#### TDD Test Cases
- Test static labels are always added
- Test managed-by label has correct value
- Test version label is included

#### Success Criteria
- StaticLabelGenerator struct implemented
- Always adds admission.pod-labeler/managed=true
- Always adds admission.pod-labeler/version=v1.0.0

### Subtask 4.4: Add Environment Label Rule
Status: Not Started

Implement environment labeling based on namespace prefix.

Commit SHA: 

#### TDD Test Cases
- Test prod- prefix results in environment=production
- Test staging- prefix results in environment=staging
- Test other namespaces get environment=development

#### Success Criteria
- EnvironmentRule implements LabelRule interface
- Namespace prefix matching logic works
- Correct environment labels generated

### Subtask 4.5: Create Label Chain Processor
Status: Not Started

Implement chain of responsibility for label rules.

Commit SHA: 

#### TDD Test Cases
- Test rules are applied in order
- Test multiple rules can add different labels
- Test label conflicts are handled (last wins)

#### Success Criteria
- DynamicLabelGenerator with rule chain
- Rules applied sequentially
- Label map properly merged

## Task 5: JSON Patch Generation
Status: Not Started

Implement JSON patch generation for pod mutations following RFC 6902.

### Success Criteria
- JSON patches correctly generated for label additions
- Patches handle cases where labels map doesn't exist
- Path escaping handles special characters properly
- Patch operations are idempotent

### Subtask 5.1: Create Patch Builder Interface
Status: Not Started

Define patch builder interface and types in pkg/patch/builder.go.

Commit SHA: 

#### TDD Test Cases
- Test Builder interface has BuildLabelPatches method
- Test patch operation types are defined
- Test JSON patch structure matches RFC 6902

#### Success Criteria
- Builder interface defined
- Patch operation types created
- RFC 6902 compliance ensured

### Subtask 5.2: Implement Label Path Detection
Status: Not Started

Add logic to detect if pod already has labels.

Commit SHA: 

#### TDD Test Cases
- Test detection when pod.metadata.labels exists
- Test detection when pod.metadata.labels is nil
- Test detection when pod.metadata doesn't exist

#### Success Criteria
- Correctly identifies label map presence
- Returns appropriate patch operations
- Handles edge cases safely

### Subtask 5.3: Create Add Labels Patch
Status: Not Started

Generate patch to add labels map if missing.

Commit SHA: 

#### TDD Test Cases
- Test patch adds /metadata/labels when missing
- Test patch operation has correct structure
- Test empty map is properly initialized

#### Success Criteria
- Add operation for missing labels map
- Correct JSON patch path
- Empty map as initial value

### Subtask 5.4: Implement Individual Label Patches
Status: Not Started

Generate patches for individual label additions.

Commit SHA: 

#### TDD Test Cases
- Test each label generates separate patch operation
- Test label keys with / are properly escaped
- Test special characters in keys are handled

#### Success Criteria
- One patch operation per label
- JSON pointer escaping implemented
- RFC 6902 path format followed

### Subtask 5.5: Create Patch Marshaling
Status: Not Started

Marshal patch operations to JSON bytes.

Commit SHA: 

#### TDD Test Cases
- Test patch array marshals to valid JSON
- Test empty patch array returns empty array
- Test malformed patches cause errors

#### Success Criteria
- Patches marshal to JSON successfully
- Proper error handling for marshal failures
- Empty patches return []

## Task 6: Certificate Management Foundation
Status: Not Started

Set up basic certificate management for TLS communication.

### Success Criteria
- Certificate loading from files implemented
- Basic certificate validation
- Certificate manager interface defined
- Preparation for cert-manager integration

### Subtask 6.1: Define Certificate Manager Interface
Status: Not Started

Create certificate manager interface in pkg/cert/manager.go.

Commit SHA: 

#### TDD Test Cases
- Test Manager interface has GetCertificate method
- Test Manager interface has certificate validation
- Test error types are properly defined

#### Success Criteria
- CertificateManager interface defined
- Methods for cert operations specified
- Error types for cert failures created

### Subtask 6.2: Implement File-based Certificate Manager
Status: Not Started

Create basic file-based certificate loading.

Commit SHA: 

#### TDD Test Cases
- Test loading certificate from valid file
- Test loading key from valid file
- Test error on missing or invalid files

#### Success Criteria
- FileCertManager struct implemented
- Certificate and key loading working
- Proper error messages for failures

### Subtask 6.3: Add Certificate Validation
Status: Not Started

Implement basic certificate validation checks.

Commit SHA: 

#### TDD Test Cases
- Test expired certificates are rejected
- Test certificates not yet valid are rejected
- Test valid certificates pass validation

#### Success Criteria
- Expiry checking implemented
- Not-before date validation
- Clear error messages for validation failures

### Subtask 6.4: Create Certificate Configuration
Status: Not Started

Add certificate-related configuration options.

Commit SHA: 

#### TDD Test Cases
- Test cert path configuration parsing
- Test key path configuration parsing
- Test default paths are sensible

#### Success Criteria
- Certificate configuration struct defined
- Command-line flags for cert paths
- Environment variable support

## Task 7: Basic Testing Framework
Status: Not Started

Establish testing framework and utilities for the project.

### Success Criteria
- Test utilities for common operations
- Mock implementations for interfaces
- Test data builders for complex objects
- Initial unit tests for core components

### Subtask 7.1: Create Test Utilities Package
Status: Not Started

Set up pkg/testutil with common test helpers.

Commit SHA: 

#### TDD Test Cases
- Test admission request builder works
- Test pod builder creates valid pods
- Test namespace builder works correctly

#### Success Criteria
- Test builder pattern implementations
- Common test data generators
- Assertion helpers created

### Subtask 7.2: Implement Mock Client
Status: Not Started

Create mock Kubernetes client for testing.

Commit SHA: 

#### TDD Test Cases
- Test mock client implements client.Client interface
- Test Get method returns configured objects
- Test List method works with options

#### Success Criteria
- Mock client with configurable responses
- Supports Get and List operations
- Error injection capability

### Subtask 7.3: Add Admission Request Builders
Status: Not Started

Create builders for test admission requests.

Commit SHA: 

#### TDD Test Cases
- Test CREATE request builder
- Test UPDATE request builder
- Test request includes all required fields

#### Success Criteria
- Fluent builder API for requests
- Supports all admission operations
- Generates valid AdmissionReview objects

### Subtask 7.4: Create Handler Test Suite
Status: Not Started

Implement comprehensive handler tests.

Commit SHA: 

#### TDD Test Cases
- Test handler with valid CREATE request
- Test handler with system namespace
- Test handler with invalid request

#### Success Criteria
- Table-driven tests for handler
- Edge cases covered
- >80% handler code coverage

### Subtask 7.5: Add Integration Test Structure
Status: Not Started

Set up integration test framework.

Commit SHA: 

#### TDD Test Cases
- Test webhook server starts in test mode
- Test admission requests are processed
- Test patches are correctly applied

#### Success Criteria
- Integration test directory structure
- Test server startup/shutdown
- End-to-end request flow tested

### Subtask 7.6: Create Benchmark Tests
Status: Not Started

Add performance benchmark tests.

Commit SHA: 

#### TDD Test Cases
- Test webhook latency benchmark runs
- Test memory allocation tracking
- Test concurrent request handling

#### Success Criteria
- Benchmark for admission handler
- Latency measurements implemented
- Performance baseline established