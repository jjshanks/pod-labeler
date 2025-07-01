# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Kubernetes mutating admission controller that automatically adds labels to pods. The project implements both a traditional webhook approach and explores the new MutatingAdmissionPolicy feature (Kubernetes v1.32+).

## Key Commands

### Building and Running
```bash
# Build the webhook binary
make build

# Run tests with coverage
make test

# Generate coverage report
make coverage

# Run linters
make lint

# Format code
make fmt

# Run the webhook locally (after building)
make run

# Build Docker image
make docker-build

# Clean build artifacts
make clean
```

### Development Workflow
```bash
# Run a single test
go test -v -run TestPodLabelMutation ./pkg/mutation/

# Run tests for a specific package
go test -v ./pkg/webhook/...

# Run benchmarks
go test -bench=. ./test/performance/

# Check code before committing
make fmt lint test
```

## Architecture Overview

The codebase follows a hybrid approach with two implementations:

1. **Traditional Webhook** (Primary)
   - Entry point: `cmd/webhook/main.go`
   - Core server: `pkg/webhook/server.go` - HTTPS server with TLS 1.2+
   - Admission handler: `pkg/webhook/handler.go` - Processes admission requests
   - Label generation: `pkg/labels/generator.go` - Chain of responsibility pattern for rules
   - JSON patches: `pkg/patch/json.go` - RFC 6902 compliant patch generation

2. **MutatingAdmissionPolicy** (Secondary, Alpha in v1.32)
   - Located in `config/crd/policy.yaml`
   - Uses CEL expressions for simple labeling rules

### Key Design Patterns
- **Chain of Responsibility**: Label rules are applied sequentially via `LabelRule` interface
- **Builder Pattern**: JSON patches are constructed via `patch.Builder`
- **Interface-based Design**: Core components (`Generator`, `Handler`, `CertificateManager`) are interfaces for testability

### Label Generation Logic
- System namespaces (kube-system, kube-public, kube-node-lease) are skipped
- Environment labels based on namespace prefix:
  - `prod-*` → `environment=production`
  - `staging-*` → `environment=staging`
  - Others → `environment=development`
- All mutated pods get `admission.pod-labeler/managed=true`

## Planning and Progress Tracking

- Planning documents are in `docs/planning/`
- Current milestone: `docs/planning/MILESTONE_1_Core_Development.md`
- Update milestone docs as tasks are completed
- Mark subtasks with commit SHAs after implementation

## Testing Strategy

The project uses multiple testing approaches:
- **Unit Tests**: Standard Go tests with testify assertions
- **Integration Tests**: Using controller-runtime's EnvTest
- **E2E Tests**: Kind cluster testing (see `test/e2e/`)
- **Performance Tests**: Benchmarks targeting <100ms p99 latency

## Security Considerations

- mTLS with client certificate verification
- Container runs as non-root user (65534)
- Read-only root filesystem
- No privilege escalation allowed
- Resource limits enforced
- Network policies restrict traffic

## Important Notes

- All commits must be signed with DCO (use `git commit -s` for Signed-off-by)
- Dependencies are managed via go.mod - run `go mod tidy` after adding new ones
- The webhook must respond within 5 seconds (configured timeout)
- Certificate rotation is handled by cert-manager in production
