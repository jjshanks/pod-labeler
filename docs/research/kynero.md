# Kyverno Mutating Admission Webhook Technical Guide

## Overview

Kyverno implements a sophisticated mutating admission webhook system that intercepts and modifies Kubernetes resources before they are persisted to etcd. This guide provides an in-depth technical analysis of the architecture, implementation, and operation of Kyverno's mutating webhook.

## Architecture Overview

### Core Components

1. **Webhook Server** (`pkg/webhooks/server.go`)
   - HTTP server using httprouter for efficient routing
   - TLS-enabled for secure communication
   - Handles multiple webhook endpoints for different policy types

2. **Webhook Controller** (`pkg/controllers/webhook/controller.go`)
   - Dynamically manages webhook configurations
   - Creates and updates MutatingWebhookConfiguration resources
   - Implements automatic cleanup and lifecycle management

3. **Mutation Handler** (`pkg/webhooks/resource/mutation/mutation.go`)
   - Core logic for processing mutation requests
   - Applies policies in order and generates JSON patches
   - Handles failure policies and error conditions

4. **Mutation Engine** (`pkg/engine/mutation.go`)
   - Executes mutation rules defined in policies
   - Supports both overlay and strategic merge patches
   - Handles conditional mutations based on context

## Request Flow

### 1. Webhook Registration

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────────┐
│ Webhook         │────▶│ Kubernetes API   │────▶│ MutatingWebhook     │
│ Controller      │     │ Server           │     │ Configuration       │
└─────────────────┘     └──────────────────┘     └─────────────────────┘
```

The webhook controller dynamically creates webhook configurations:
- Monitors policy changes
- Generates webhook rules based on active policies
- Updates configurations when policies change

### 2. Mutation Request Processing

```
┌──────────────┐     ┌─────────────┐     ┌──────────────┐     ┌────────────┐
│ API Server   │────▶│ Webhook     │────▶│ Admission    │────▶│ Mutation   │
│              │     │ Server      │     │ Handler      │     │ Handler    │
└──────────────┘     └─────────────┘     └──────────────┘     └────────────┘
                                                                       │
                                                                       ▼
┌──────────────┐     ┌─────────────┐     ┌──────────────┐     ┌────────────┐
│ API Server   │◀────│ JSON Patch  │◀────│ Mutation     │◀────│ Policy     │
│ (patched)    │     │ Response    │     │ Engine       │     │ Processor  │
└──────────────┘     └─────────────┘     └──────────────┘     └────────────┘
```

## Webhook Endpoints

The server exposes multiple endpoints for different policy types:

### Resource Mutation Endpoints
- `/mutate/*` - General resource mutation
- `/mpol/*` - Mutating policy-specific endpoint
- `/ivpol/mutate/*` - Image verification policy mutations

### Policy-Specific Routes
```go
// Mutating policies route
mux.HandlerFunc("POST", "/mpol/*policies", 
    handlerFunc("MUTATE", resourceHandlers.MutatingPolicies, "")...)

// Image verification mutations
mux.HandlerFunc("POST", "/ivpol/mutate/*policies",
    handlerFunc("IVPOL-MUTATE", resourceHandlers.ImageVerificationPoliciesMutation, "")...)
```

## Mutation Processing Pipeline

### 1. Request Validation
- Validates admission request format
- Checks content type and body
- Extracts resource information

### 2. Policy Selection
```go
func (h *mutationHandler) applyMutations(
    ctx context.Context,
    request handlers.AdmissionRequest,
    policies []kyvernov1.PolicyInterface,
    policyContext *engine.PolicyContext,
    cfg config.Configuration,
) ([]byte, []engineapi.EngineResponse, error)
```

### 3. Mutation Application
For each applicable policy:
1. Check if policy has mutation rules
2. Apply mutations in order
3. Accumulate JSON patches
4. Update resource state for next policy

### 4. Patch Generation
```go
var patches []jsonpatch.JsonPatchOperation
for _, policy := range policies {
    engineResponse, policyPatches, err := v.applyMutation(ctx, request, currentContext, failurePolicy, policy)
    if len(policyPatches) > 0 {
        patches = append(patches, policyPatches...)
    }
}
```

## Key Design Patterns

### 1. Handler Chain Pattern
```go
handlerFunc("MUTATE", resourceHandlers.Mutation, "").
    WithFilter(configuration).
    WithProtection(toggle.FromContext(ctx).ProtectManagedResources()).
    WithDump(debugModeOpts.DumpPayload).
    WithTopLevelGVK(discovery).
    WithRoles(rbLister, crbLister).
    WithMetrics(resourceLogger, metricsConfig.Config(), metrics.WebhookMutating).
    WithAdmission(resourceLogger.WithName("mutate"))
```

### 2. Policy Context Management
- Maintains state across policy evaluations
- Updates resource after each mutation
- Preserves original resource for comparison

### 3. Failure Policy Handling
```go
if policy.GetSpec().GetFailurePolicy(ctx) == kyvernov1.Fail {
    failurePolicy = kyvernov1.Fail
}
```

## Dynamic Webhook Configuration

### Controller Logic
The webhook controller (`pkg/controllers/webhook/controller.go`) implements:

1. **Policy Watching**: Monitors policy changes
2. **Rule Generation**: Creates webhook rules from policies
3. **Configuration Updates**: Updates webhook configurations
4. **Cleanup**: Removes configurations when policies are deleted

### Configuration Structure
```go
type controller struct {
    client           dclient.Interface
    kyvernoClient    versioned.Interface
    policyLister     kyvernov1listers.ClusterPolicyLister
    mwcClient        controllerutils.CreateUpdateDeleteClient
    vwcClient        controllerutils.CreateUpdateDeleteClient
    // ... other fields
}
```

## Mutation Engine Details

### Rule Processing
```go
func (e *engine) mutate(
    ctx context.Context,
    logger logr.Logger,
    policyContext engineapi.PolicyContext,
) (engineapi.PolicyResponse, unstructured.Unstructured) {
    for _, rule := range autogen.Default.ComputeRules(policy, "") {
        if !rule.HasMutate() {
            continue
        }
        // Apply mutation...
    }
}
```

### Mutation Types
1. **Overlay Mutations**: Direct field replacement
2. **Strategic Merge Patches**: Kubernetes-aware merging
3. **JSON Patches**: RFC 6902 operations
4. **Conditional Mutations**: Based on context variables

## Performance Considerations

### 1. Efficient Routing
- Uses httprouter for O(1) route matching
- Minimal allocations in request path

### 2. Policy Caching
- In-memory policy cache
- Indexed by resource type for fast lookup

### 3. Parallel Processing
- Independent policies can be evaluated concurrently
- Results aggregated after processing

### 4. Resource Optimization
```go
const (
    Workers                   = 2
    DefaultWebhookTimeout     = 10
    IdleDeadline              = tickerInterval * 10
    maxRetries                = 10
)
```

## Security Features

### 1. TLS Configuration
- Certificate rotation support
- Root CA validation
- Secure communication with API server

### 2. RBAC Integration
- Role and ClusterRole validation
- User permission checks
- Service account verification

### 3. Resource Protection
```go
WithProtection(toggle.FromContext(ctx).ProtectManagedResources())
```

## Monitoring and Observability

### 1. Metrics Collection
- Request duration
- Policy execution time
- Success/failure rates
- Patch generation statistics

### 2. Logging
```go
logger := logger.WithValues(
    "gvk", admissionReview.Request.Kind,
    "gvr", admissionReview.Request.Resource.String(),
    "namespace", admissionReview.Request.Namespace,
    "name", admissionReview.Request.Name,
    "operation", admissionReview.Request.Operation,
)
```

### 3. Tracing
- OpenTelemetry integration
- Distributed trace support
- Performance profiling

## Error Handling

### 1. Admission Errors
- Proper HTTP status codes
- Detailed error messages
- Admission response with reasons

### 2. Failure Policies
- **Ignore**: Continue on error
- **Fail**: Block resource creation/update

### 3. Timeout Handling
```go
timeout := time.Duration(cfg.GetWebhookTimeout()) * time.Second
ctx, cancel := context.WithTimeout(ctx, timeout)
defer cancel()
```

## Best Practices

### 1. Policy Design
- Keep mutations atomic
- Use specific match criteria
- Test policies thoroughly
- Consider performance impact

### 2. Webhook Configuration
- Set appropriate timeouts
- Configure failure policies
- Monitor webhook health
- Plan for upgrades

### 3. Debugging
- Enable debug logging
- Use admission review dumps
- Monitor metrics
- Test in staging first

## Conclusion

Kyverno's mutating admission webhook provides a powerful and flexible system for modifying Kubernetes resources. Its design emphasizes:
- Dynamic configuration management
- Efficient request processing
- Comprehensive error handling
- Strong security practices
- Excellent observability

The architecture supports complex mutation scenarios while maintaining performance and reliability, making it suitable for production Kubernetes environments.