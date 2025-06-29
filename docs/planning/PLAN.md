# Kubernetes Mutating Admission Controller Implementation Plan

## Executive Summary

This comprehensive implementation plan outlines the development of a production-ready Kubernetes mutating admission controller for automatic pod labeling. The controller will support Kubernetes v1.32+ environments, implement modern security practices for 2025, and include extensive testing coverage. This plan synthesizes research from multiple sources to create a robust, secure, and performant solution.

## Project Overview

### Objectives
- Create a mutating admission controller that automatically adds labels to pods
- Support Kubernetes v1.32+ with exploration of new MutatingAdmissionPolicy features
- Implement comprehensive security measures aligned with 2025 standards
- Achieve sub-100ms latency at p99 for webhook processing
- Provide extensive testing coverage (>90% unit tests)
- Deploy with production-ready monitoring and operational excellence

### Architecture Decision

After careful analysis, we will implement a **hybrid approach**:

1. **Primary Implementation**: Traditional mutating webhook using Go and controller-runtime
   - Provides full programmatic control
   - Supports complex labeling logic
   - Enables integration with external systems
   - Mature and battle-tested approach

2. **Secondary Implementation**: MutatingAdmissionPolicy with CEL (v1.32 alpha)
   - For simple, declarative labeling rules
   - In-process execution (no network latency)
   - Simplified deployment for basic scenarios
   - Future-proofing for when feature becomes stable

## Technical Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                   Kubernetes API Server                      │
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────┐   │
│  │Authentication│  │Authorization │  │Schema Validation│   │
│  └──────┬──────┘  └──────┬───────┘  └────────┬────────┘   │
│         │                 │                    │            │
│         ▼                 ▼                    ▼            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │          Mutating Admission Phase                    │   │
│  │  ┌─────────────────┐  ┌────────────────────────┐   │   │
│  │  │MutatingAdmission│  │  Webhook Controller    │   │   │
│  │  │Policy (CEL)     │  │  (External HTTPS)     │   │   │
│  │  └─────────────────┘  └────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────┘   │
│                              │                              │
│                              ▼                              │
│  ┌─────────────────────────────────────────────────────┐   │
│  │          Validating Admission Phase                  │   │
│  └─────────────────────────────────────────────────────┘   │
│                              │                              │
│                              ▼                              │
│                         etcd Storage                        │
└─────────────────────────────────────────────────────────────┘
```

### Webhook Server Architecture

```go
type WebhookServer struct {
    server         *http.Server
    certManager    CertificateManager
    labelHandler   LabelMutationHandler
    metrics        *prometheus.Registry
    cache          *cache.LRU
    healthChecker  HealthChecker
}
```

Key components:
- **HTTPS Server**: TLS 1.2+ with mutual authentication
- **Certificate Manager**: Automated rotation via cert-manager
- **Label Handler**: Core mutation logic with chain of responsibility pattern
- **Metrics Collection**: Prometheus integration
- **LRU Cache**: Performance optimization for repeated requests
- **Health Endpoints**: Liveness and readiness probes

## Implementation Phases

### Phase 1: Core Development (Week 1-2)

#### 1.1 Project Setup
```bash
# Initialize Go module
go mod init github.com/jjshanks/pod-labeler

# Project structure
pod-labeler/
├── cmd/
│   └── webhook/
│       └── main.go              # Entry point
├── pkg/
│   ├── webhook/
│   │   ├── server.go           # HTTPS server setup
│   │   ├── handler.go          # Admission handler
│   │   └── handler_test.go
│   ├── mutation/
│   │   ├── mutator.go          # Core mutation logic
│   │   ├── mutator_test.go
│   │   └── types.go
│   ├── labels/
│   │   ├── generator.go        # Label generation logic
│   │   ├── validator.go        # Label validation
│   │   └── rules.go            # Label rules engine
│   └── patch/
│       ├── json.go             # JSON patch generation
│       └── json_test.go
├── config/
│   ├── webhook/
│   │   └── config.yaml         # Webhook configuration
│   └── crd/
│       └── policy.yaml         # MutatingAdmissionPolicy
├── deploy/
│   ├── helm/                   # Helm chart
│   └── manifests/              # Raw Kubernetes manifests
└── test/
    ├── unit/
    ├── integration/
    └── e2e/
```

#### 1.2 Core Webhook Implementation

```go
// pkg/webhook/handler.go
package webhook

import (
    "context"
    "encoding/json"
    "net/http"
    
    admissionv1 "k8s.io/api/admission/v1"
    corev1 "k8s.io/api/core/v1"
    "sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type PodLabelMutator struct {
    Client       client.Client
    decoder      *admission.Decoder
    labelGen     labels.Generator
    patchBuilder patch.Builder
}

func (m *PodLabelMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
    pod := &corev1.Pod{}
    if err := m.decoder.Decode(req, pod); err != nil {
        return admission.Errored(http.StatusBadRequest, err)
    }
    
    // Skip system namespaces
    if isSystemNamespace(req.Namespace) {
        return admission.Allowed("system namespace")
    }
    
    // Generate labels based on context
    labels, err := m.labelGen.GenerateLabels(ctx, pod, req.Namespace)
    if err != nil {
        return admission.Errored(http.StatusInternalServerError, err)
    }
    
    // Generate JSON patches
    patches := m.patchBuilder.BuildLabelPatches(pod, labels)
    patchBytes, err := json.Marshal(patches)
    if err != nil {
        return admission.Errored(http.StatusInternalServerError, err)
    }
    
    return admission.PatchResponseFromRaw(req.Object.Raw, patchBytes)
}
```

#### 1.3 Label Generation Strategy

```go
// pkg/labels/generator.go
package labels

type Generator interface {
    GenerateLabels(ctx context.Context, pod *corev1.Pod, namespace string) (map[string]string, error)
}

type DynamicLabelGenerator struct {
    client    client.Client
    ruleChain []LabelRule
}

func (g *DynamicLabelGenerator) GenerateLabels(ctx context.Context, pod *corev1.Pod, namespaceName string) (map[string]string, error) {
    labels := make(map[string]string)
    
    // Fetch namespace for context
    namespace := &corev1.Namespace{}
    if err := g.client.Get(ctx, types.NamespacedName{Name: namespaceName}, namespace); err != nil {
        return nil, err
    }
    
    // Apply rule chain
    for _, rule := range g.ruleChain {
        if rule.Matches(pod, namespace) {
            ruleLables := rule.GenerateLabels(pod, namespace)
            for k, v := range ruleLabels {
                labels[k] = v
            }
        }
    }
    
    // Always add managed-by label
    labels["admission.pod-labeler/managed"] = "true"
    labels["admission.pod-labeler/version"] = "v1.0.0"
    
    return labels, nil
}
```

### Phase 2: Security Implementation (Week 2-3)

#### 2.1 mTLS Configuration

```yaml
# config/webhook/tls-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: webhook-tls-config
  namespace: pod-labeler
data:
  tls.yaml: |
    server:
      minVersion: "1.2"
      cipherSuites:
        - TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
        - TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
      clientAuth: RequireAndVerifyClientCert
      clientCAs:
        - /etc/webhook/ca/ca.crt
```

#### 2.2 Container Security

```yaml
# deploy/manifests/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pod-labeler-webhook
  namespace: pod-labeler
spec:
  replicas: 3
  selector:
    matchLabels:
      app: pod-labeler-webhook
  template:
    metadata:
      labels:
        app: pod-labeler-webhook
    spec:
      serviceAccountName: pod-labeler-webhook
      automountServiceAccountToken: false
      securityContext:
        runAsNonRoot: true
        runAsUser: 65534
        fsGroup: 65534
        seccompProfile:
          type: RuntimeDefault
      containers:
      - name: webhook
        image: pod-labeler:v1.0.0
        ports:
        - containerPort: 8443
          name: webhook
        - containerPort: 8080
          name: metrics
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop: ["ALL"]
          readOnlyRootFilesystem: true
          runAsNonRoot: true
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 256Mi
        env:
        - name: GOMEMLIMIT
          value: "240MiB"
        volumeMounts:
        - name: webhook-tls
          mountPath: /etc/webhook/tls
          readOnly: true
        - name: webhook-token
          mountPath: /var/run/secrets/kubernetes.io/serviceaccount
          readOnly: true
        - name: tmp
          mountPath: /tmp
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: webhook-tls
        secret:
          secretName: pod-labeler-webhook-tls
      - name: webhook-token
        projected:
          sources:
          - serviceAccountToken:
              path: token
              expirationSeconds: 3600
              audience: webhook-server
      - name: tmp
        emptyDir: {}
```

#### 2.3 RBAC Configuration

```yaml
# deploy/manifests/rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: pod-labeler-webhook
  namespace: pod-labeler
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: pod-labeler-webhook
rules:
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get", "list"]
- apiGroups: ["admissionregistration.k8s.io"]
  resources: ["mutatingwebhookconfigurations"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: pod-labeler-webhook
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: pod-labeler-webhook
subjects:
- kind: ServiceAccount
  name: pod-labeler-webhook
  namespace: pod-labeler
```

#### 2.4 Network Policies

```yaml
# deploy/manifests/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: pod-labeler-webhook
  namespace: pod-labeler
spec:
  podSelector:
    matchLabels:
      app: pod-labeler-webhook
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: TCP
      port: 8443
  - from:
    - podSelector: {}
    ports:
    - protocol: TCP
      port: 8080  # Metrics
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: TCP
      port: 443  # Kubernetes API
  - to:
    - podSelector: {}
    ports:
    - protocol: TCP
      port: 53   # DNS
```

### Phase 3: Certificate Management (Week 3)

#### 3.1 cert-manager Integration

```yaml
# deploy/manifests/certificate.yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: pod-labeler-webhook-tls
  namespace: pod-labeler
spec:
  secretName: pod-labeler-webhook-tls
  dnsNames:
  - pod-labeler-webhook.pod-labeler.svc
  - pod-labeler-webhook.pod-labeler.svc.cluster.local
  issuerRef:
    name: pod-labeler-ca-issuer
    kind: ClusterIssuer
  duration: 8760h   # 1 year
  renewBefore: 240h # 10 days
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: pod-labeler-ca-issuer
spec:
  ca:
    secretName: pod-labeler-ca-secret
```

#### 3.2 Webhook Configuration

```yaml
# deploy/manifests/webhook-configuration.yaml
apiVersion: admissionregistration.k8s.io/v1
kind: MutatingWebhookConfiguration
metadata:
  name: pod-labeler-webhook
  annotations:
    cert-manager.io/inject-ca-from: pod-labeler/pod-labeler-webhook-tls
webhooks:
- name: pod-labeler.admission.k8s.io
  clientConfig:
    service:
      name: pod-labeler-webhook
      namespace: pod-labeler
      path: /mutate
    caBundle: <WILL BE INJECTED BY CERT-MANAGER>
  rules:
  - operations: ["CREATE", "UPDATE"]
    apiGroups: [""]
    apiVersions: ["v1"]
    resources: ["pods"]
  namespaceSelector:
    matchExpressions:
    - key: kubernetes.io/metadata.name
      operator: NotIn
      values: ["kube-system", "kube-node-lease", "kube-public"]
  objectSelector:
    matchExpressions:
    - key: admission.pod-labeler/skip
      operator: NotIn
      values: ["true"]
  sideEffects: None
  timeoutSeconds: 5
  failurePolicy: Fail
  admissionReviewVersions: ["v1", "v1beta1"]
```

### Phase 4: Testing Strategy (Week 3-4)

#### 4.1 Unit Testing

```go
// pkg/mutation/mutator_test.go
package mutation

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodLabelMutation(t *testing.T) {
    tests := []struct {
        name           string
        pod            *corev1.Pod
        namespace      *corev1.Namespace
        expectedLabels map[string]string
        expectError    bool
    }{
        {
            name: "production namespace labeling",
            pod: &corev1.Pod{
                ObjectMeta: metav1.ObjectMeta{
                    Name: "test-pod",
                },
            },
            namespace: &corev1.Namespace{
                ObjectMeta: metav1.ObjectMeta{
                    Name: "prod-app",
                },
            },
            expectedLabels: map[string]string{
                "environment":         "production",
                "compliance-required": "true",
                "admission.pod-labeler/managed": "true",
            },
        },
        {
            name: "staging namespace labeling",
            pod: &corev1.Pod{
                ObjectMeta: metav1.ObjectMeta{
                    Name: "test-pod",
                },
            },
            namespace: &corev1.Namespace{
                ObjectMeta: metav1.ObjectMeta{
                    Name: "staging-app",
                },
            },
            expectedLabels: map[string]string{
                "environment": "staging",
                "admission.pod-labeler/managed": "true",
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            generator := NewDynamicLabelGenerator(nil, defaultRuleChain())
            labels, err := generator.GenerateLabels(context.Background(), tt.pod, tt.namespace)
            
            if tt.expectError {
                assert.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            for key, expectedValue := range tt.expectedLabels {
                assert.Equal(t, expectedValue, labels[key])
            }
        })
    }
}
```

#### 4.2 Integration Testing with EnvTest

```go
// test/integration/webhook_test.go
package integration

import (
    "context"
    "path/filepath"
    "testing"
    "time"
    
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/envtest"
)

var _ = Describe("Pod Label Mutation", func() {
    var (
        ctx       context.Context
        k8sClient client.Client
        testEnv   *envtest.Environment
    )
    
    BeforeEach(func() {
        ctx = context.Background()
        testEnv = &envtest.Environment{
            CRDDirectoryPaths: []string{filepath.Join("..", "..", "config", "crd")},
            WebhookInstallOptions: envtest.WebhookInstallOptions{
                Paths: []string{filepath.Join("..", "..", "config", "webhook")},
            },
        }
        
        cfg, err := testEnv.Start()
        Expect(err).NotTo(HaveOccurred())
        
        k8sClient, err = client.New(cfg, client.Options{})
        Expect(err).NotTo(HaveOccurred())
    })
    
    AfterEach(func() {
        err := testEnv.Stop()
        Expect(err).NotTo(HaveOccurred())
    })
    
    Context("when creating pods in different namespaces", func() {
        It("should add production labels to pods in prod namespace", func() {
            namespace := &corev1.Namespace{
                ObjectMeta: metav1.ObjectMeta{
                    Name: "prod-test",
                },
            }
            Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
            
            pod := &corev1.Pod{
                ObjectMeta: metav1.ObjectMeta{
                    Name:      "test-pod",
                    Namespace: namespace.Name,
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{{
                        Name:  "test",
                        Image: "nginx:latest",
                    }},
                },
            }
            
            Expect(k8sClient.Create(ctx, pod)).To(Succeed())
            
            Eventually(func(g Gomega) {
                var createdPod corev1.Pod
                g.Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(pod), &createdPod)).To(Succeed())
                g.Expect(createdPod.Labels).To(HaveKeyWithValue("environment", "production"))
                g.Expect(createdPod.Labels).To(HaveKeyWithValue("compliance-required", "true"))
            }, 10*time.Second).Should(Succeed())
        })
    })
})
```

#### 4.3 Performance Testing

```go
// test/performance/benchmark_test.go
package performance

import (
    "context"
    "testing"
    "time"
    
    "github.com/jjshanks/pod-labeler/pkg/webhook"
)

func BenchmarkWebhookLatency(b *testing.B) {
    handler := webhook.NewPodLabelMutator()
    req := createBenchmarkAdmissionRequest()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        start := time.Now()
        resp := handler.Handle(context.Background(), req)
        duration := time.Since(start)
        
        if !resp.Allowed {
            b.Fatalf("unexpected rejection: %v", resp.Result)
        }
        
        // Target: <100ms latency
        if duration > 100*time.Millisecond {
            b.Errorf("latency %v exceeds 100ms target", duration)
        }
    }
}
```

#### 4.4 Local Cluster Testing

```yaml
# test/e2e/kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 30443
    hostPort: 30443
    protocol: TCP
- role: worker
- role: worker
networking:
  podSubnet: "10.244.0.0/16"
  serviceSubnet: "10.96.0.0/12"
```

```bash
#!/bin/bash
# test/e2e/run-e2e-tests.sh

# Create Kind cluster
kind create cluster --config kind-config.yaml --name pod-labeler-test

# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# Wait for cert-manager
kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=cert-manager -n cert-manager --timeout=60s

# Deploy webhook
kubectl apply -f ../../deploy/manifests/

# Run e2e tests
go test ./... -tags=e2e -v

# Cleanup
kind delete cluster --name pod-labeler-test
```

### Phase 5: MutatingAdmissionPolicy Implementation (Week 4)

```yaml
# config/crd/mutating-admission-policy.yaml
apiVersion: admissionregistration.k8s.io/v1alpha1
kind: MutatingAdmissionPolicy
metadata:
  name: pod-environment-labels
spec:
  matchConstraints:
    resourceRules:
    - apiGroups: [""]
      apiVersions: ["v1"]
      operations: ["CREATE", "UPDATE"]
      resources: ["pods"]
  mutations:
  - patchType: "ApplyConfiguration"
    applyConfiguration:
      expression: |
        Object{
          metadata: Object.metadata{
            labels: object.metadata.labels + {
              "environment": has(object.metadata.namespace) && object.metadata.namespace.startsWith("prod-") ? "production" : 
                            has(object.metadata.namespace) && object.metadata.namespace.startsWith("staging-") ? "staging" : "development",
              "managed-by": "admission-policy"
            }
          }
        }
---
apiVersion: admissionregistration.k8s.io/v1alpha1
kind: MutatingAdmissionPolicyBinding
metadata:
  name: pod-environment-labels-binding
spec:
  policyName: pod-environment-labels
  validationActions: ["Deny"]
  matchResources:
    namespaceSelector:
      matchExpressions:
      - key: kubernetes.io/metadata.name
        operator: NotIn
        values: ["kube-system", "kube-node-lease", "kube-public"]
```

### Phase 6: Deployment and Operations (Week 5)

#### 6.1 Helm Chart

```yaml
# deploy/helm/pod-labeler/values.yaml
replicaCount: 3

image:
  repository: ghcr.io/jjshanks/pod-labeler
  pullPolicy: IfNotPresent
  tag: "v1.0.0"

webhook:
  failurePolicy: Fail
  timeoutSeconds: 5
  namespaceSelector:
    matchExpressions:
    - key: kubernetes.io/metadata.name
      operator: NotIn
      values: ["kube-system", "kube-node-lease", "kube-public"]

service:
  type: ClusterIP
  port: 443
  targetPort: 8443

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 50
  targetMemoryUtilizationPercentage: 80

certificates:
  useCertManager: true
  duration: 8760h
  renewBefore: 240h

monitoring:
  enabled: true
  serviceMonitor:
    enabled: true
    interval: 30s
  prometheusRule:
    enabled: true
    rules:
    - alert: WebhookHighLatency
      expr: histogram_quantile(0.99, webhook_request_duration_seconds_bucket) > 0.1
      for: 5m
      labels:
        severity: warning
      annotations:
        summary: "Webhook latency is high"
        description: "99th percentile latency {{ $value }}s exceeds 100ms target"
    - alert: WebhookCertificateExpiringSoon
      expr: webhook_certificate_expiry_seconds < 604800
      for: 1h
      labels:
        severity: warning
      annotations:
        summary: "Webhook certificate expiring soon"
        description: "Certificate expires in {{ $value | humanizeDuration }}"

securityContext:
  runAsNonRoot: true
  runAsUser: 65534
  fsGroup: 65534
  seccompProfile:
    type: RuntimeDefault

podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65534
  fsGroup: 65534

containerSecurityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop: ["ALL"]
  readOnlyRootFilesystem: true
  runAsNonRoot: true
```

#### 6.2 Monitoring and Observability

```go
// pkg/metrics/metrics.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    WebhookRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "webhook_request_duration_seconds",
            Help: "Duration of admission webhook requests in seconds",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
        },
        []string{"operation", "namespace", "result"},
    )
    
    WebhookRequestTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "webhook_request_total",
            Help: "Total number of admission webhook requests",
        },
        []string{"operation", "namespace", "result"},
    )
    
    LabelsAddedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "webhook_labels_added_total",
            Help: "Total number of labels added by the webhook",
        },
        []string{"label_key", "namespace"},
    )
    
    CertificateExpirySeconds = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "webhook_certificate_expiry_seconds",
            Help: "Time until webhook certificate expires in seconds",
        },
    )
)
```

```yaml
# deploy/manifests/service-monitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: pod-labeler-webhook
  namespace: pod-labeler
spec:
  selector:
    matchLabels:
      app: pod-labeler-webhook
  endpoints:
  - port: metrics
    interval: 30s
    path: /metrics
```

#### 6.3 Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o webhook cmd/webhook/main.go

# Production stage
FROM gcr.io/distroless/static:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/webhook /webhook

USER 65532:65532
EXPOSE 8443 8080

ENTRYPOINT ["/webhook"]
```

### Phase 7: Documentation (Week 5-6)

#### 7.1 API Documentation

```markdown
# Pod Labeler Webhook API

## Endpoints

### POST /mutate
Admission webhook endpoint for pod mutations.

**Request**: AdmissionReview v1
**Response**: AdmissionReview v1 with patches

### GET /healthz
Liveness probe endpoint.

**Response**: 200 OK

### GET /readyz
Readiness probe endpoint.

**Response**: 200 OK when ready

### GET /metrics
Prometheus metrics endpoint.

**Response**: Prometheus text format
```

#### 7.2 Operational Runbook

```markdown
# Pod Labeler Operational Runbook

## Deployment

### Prerequisites
- Kubernetes 1.32+
- cert-manager 1.14+
- Prometheus Operator (optional)

### Installation
```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# Install webhook using Helm
helm install pod-labeler ./deploy/helm/pod-labeler \
  --namespace pod-labeler \
  --create-namespace
```

## Troubleshooting

### Certificate Issues
```bash
# Check certificate status
kubectl get certificate -n pod-labeler

# Check webhook configuration
kubectl get mutatingwebhookconfiguration pod-labeler-webhook -o yaml

# Manually refresh certificate
kubectl delete secret pod-labeler-webhook-tls -n pod-labeler
```

### Performance Issues
```bash
# Check webhook metrics
kubectl port-forward -n pod-labeler svc/pod-labeler-webhook 8080:8080
curl http://localhost:8080/metrics | grep webhook_request_duration

# Check pod resources
kubectl top pods -n pod-labeler
```

### Webhook Failures
```bash
# Check webhook logs
kubectl logs -n pod-labeler -l app=pod-labeler-webhook

# Test webhook connectivity
kubectl run test-pod --image=nginx --dry-run=server -o yaml
```
```

## Success Metrics and Monitoring

### Key Performance Indicators (KPIs)

1. **Latency Metrics**
   - p50 latency: <20ms
   - p95 latency: <50ms
   - p99 latency: <100ms

2. **Availability Metrics**
   - Uptime: 99.9%
   - Error rate: <0.1%
   - Certificate renewal success: 100%

3. **Operational Metrics**
   - Pod mutation success rate: >99.9%
   - Label validation errors: <0.01%
   - Resource utilization: <50% of limits

### Monitoring Dashboard

```yaml
# deploy/manifests/grafana-dashboard.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: pod-labeler-dashboard
  namespace: monitoring
data:
  dashboard.json: |
    {
      "dashboard": {
        "title": "Pod Labeler Webhook",
        "panels": [
          {
            "title": "Request Latency (p99)",
            "targets": [{
              "expr": "histogram_quantile(0.99, webhook_request_duration_seconds_bucket)"
            }]
          },
          {
            "title": "Request Rate",
            "targets": [{
              "expr": "rate(webhook_request_total[5m])"
            }]
          },
          {
            "title": "Error Rate",
            "targets": [{
              "expr": "rate(webhook_request_total{result=\"error\"}[5m])"
            }]
          },
          {
            "title": "Certificate Expiry",
            "targets": [{
              "expr": "webhook_certificate_expiry_seconds / 86400"
            }]
          }
        ]
      }
    }
```

## Risk Analysis and Mitigation

### Identified Risks

1. **Certificate Expiry**
   - **Risk**: Service outage if certificates expire
   - **Mitigation**: Automated renewal via cert-manager, monitoring alerts at 10 days before expiry

2. **Performance Degradation**
   - **Risk**: High latency impacts API server performance
   - **Mitigation**: Aggressive timeouts (5s), horizontal scaling, caching

3. **Webhook Failure**
   - **Risk**: Pod creation blocked if webhook fails
   - **Mitigation**: High availability (3+ replicas), health checks, circuit breakers

4. **Security Breach**
   - **Risk**: Compromised webhook could inject malicious labels
   - **Mitigation**: mTLS, least privilege RBAC, input validation, audit logging

5. **Resource Exhaustion**
   - **Risk**: Memory leaks or CPU spikes
   - **Mitigation**: Resource limits, GOMEMLIMIT, HPA, monitoring

## Timeline and Milestones

### Week 1-2: Core Development
- ✓ Project setup and structure
- ✓ Core webhook implementation
- ✓ Label generation logic
- ✓ Basic unit tests

### Week 2-3: Security Implementation
- ✓ mTLS configuration
- ✓ Container hardening
- ✓ RBAC and network policies
- ✓ Security testing

### Week 3: Certificate Management
- ✓ cert-manager integration
- ✓ Webhook configuration
- ✓ Certificate rotation testing

### Week 3-4: Testing
- ✓ Comprehensive unit tests (>90% coverage)
- ✓ Integration tests with EnvTest
- ✓ Performance benchmarks
- ✓ E2E tests with Kind

### Week 4: Advanced Features
- ✓ MutatingAdmissionPolicy implementation
- ✓ Advanced label rules
- ✓ Caching optimization

### Week 5: Deployment
- ✓ Helm chart creation
- ✓ CI/CD pipeline
- ✓ Monitoring setup
- ✓ Production readiness

### Week 5-6: Documentation
- ✓ API documentation
- ✓ Operational runbooks
- ✓ Security documentation
- ✓ User guides

## Conclusion

This comprehensive implementation plan provides a clear roadmap for developing a production-ready Kubernetes mutating admission controller. By following this plan, we will deliver a secure, performant, and maintainable solution that meets all specified requirements while incorporating modern best practices and future-proofing through support for Kubernetes v1.32+ features.

The hybrid approach of combining traditional webhooks with the new MutatingAdmissionPolicy feature ensures flexibility and positions the solution for long-term success as Kubernetes continues to evolve.