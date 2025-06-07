# Kubenetes Health and Cost Management Project!

Ochestra AI is an cloud native K8's management tool that leverages artificial inteligence to simplify and automate the management of cloud native workloads on Kubernetes. Its main purpose is to help solve the challenges associated with operating cloud - native applications at scale - Complexity, Cost and Performance.


# Kubernetes Health and Cost Management Tool

A comprehensive Go-based tool for monitoring Kubernetes cluster health and managing costs. This tool provides real-time health assessments, cost tracking, optimization recommendations, and detailed reporting for Kubernetes environments.

## Features

### 🏥 Health Monitoring
- **Node Health**: Monitor node status, resource pressure, and availability
- **Pod Health**: Track pod states, restart counts, and crash loops
- **Control Plane**: Monitor API server, etcd, scheduler, and controller manager
- **Network Health**: Check CNI, DNS resolution, service endpoints, and ingress
- **Resource Usage**: Track CPU, memory, and storage utilization
- **Health Scoring**: Overall cluster health score (0-100)

### 💰 Cost Management
- **Node Costs**: Calculate costs by instance type and region
- **Pod Costs**: Track resource consumption and costs per workload
- **Namespace Costs**: Aggregate costs by namespace
- **Cost Forecasting**: Project future costs based on usage trends
- **Optimization**: Identify over-provisioned resources and cost savings

### 📊 Reporting
- **Multiple Formats**: JSON, HTML, and text output
- **Interactive Dashboards**: Visual HTML reports with charts
- **Prometheus Metrics**: Export metrics for monitoring systems
- **Combined Reports**: Health and cost analysis in one view

### 🔧 Automation
- **Resource Cleanup**: Automated cleanup of unused resources
- **Cost Alerts**: Monitor cost changes and send notifications
- **Continuous Monitoring**: Run as a service with configurable intervals
- **Optimization Recommendations**: Automated suggestions for improvements

## Installation

### Prerequisites
- Go 1.19 or later
- Access to a Kubernetes cluster
- `kubectl` configured with cluster access
- (Optional) Metrics Server deployed in the cluster for detailed resource usage

### Build from Source

```bash
# Clone the repository
git clone https://github.com/ochestra-tech/ochestra-ai
cd ochestra-ai

# Download dependencies
go mod tidy

# Build the application
go build -o ochestra-ai ./cmd/main.go
```

### Dependencies

The tool requires the following Go modules:

```bash
go get k8s.io/client-go@latest
go get k8s.io/api@latest
go get k8s.io/apimachinery@latest
go get k8s.io/metrics@latest
go get github.com/prometheus/client_golang@latest
go get github.com/olekukonko/tablewriter@latest
```

## Configuration

### Kubeconfig
The tool uses your existing kubeconfig file. By default, it looks for `~/.kube/config`, but you can specify a different path:

```bash
./ochestra-ai --kubeconfig /path/to/kubeconfig
```

### Pricing Configuration
Create a `pricing-config.json` file to define your cloud pricing:

```json
{
  "defaults": {
    "cpu": 0.03,
    "memory": 0.004,
    "storage": 0.00012,
    "network": 0.08,
    "gpuPricing": {
      "nvidia-tesla-v100": 1.2,
      "nvidia-tesla-k80": 0.6
    }
  },
  "instanceTypes": {
    "m5.large": {
      "cpu": 0.032,
      "memory": 0.0045,
      "storage": 0.00015,
      "network": 0.09
    },
    "c5.large": {
      "cpu": 0.035,
      "memory": 0.0035,
      "storage": 0.00018,
      "network": 0.095
    }
  },
  "regionMultipliers": {
    "us-east-1": 1.0,
    "us-west-2": 1.05,
    "eu-west-1": 1.1,
    "ap-southeast-1": 1.15
  }
}
```

## Usage

### Basic Commands

#### Health Check
```bash
# Quick health check
./ochestra-ai --type health --format text

# Detailed health report in HTML
./ochestra-ai --type health --format html --output health-report.html
```

#### Cost Analysis
```bash
# Cost report in JSON format
./ochestra-ai --type cost --format json --output cost-report.json

# Monthly cost breakdown
./ochestra-ai --type cost --format text
```

#### Combined Report
```bash
# Complete health and cost analysis
./ochestra-ai --type combined --format html --output cluster-report.html
```

### Continuous Monitoring

```bash
# Monitor every 5 minutes with Prometheus metrics
./ochestra-ai --interval 5m --metrics-port 8080

# Custom configuration
./ochestra-ai \
  --kubeconfig ~/.kube/config \
  --pricing-config ./my-pricing.json \
  --interval 10m \
  --metrics-port 9090 \
  --type combined \
  --format json \
  --output /var/log/k8s-reports/report.json
```

### Command Line Options

| Option | Description | Default |
|--------|-------------|---------|
| `--kubeconfig` | Path to kubeconfig file | `~/.kube/config` |
| `--pricing-config` | Path to pricing configuration | `pricing-config.json` |
| `--type` | Report type (health, cost, combined) | `combined` |
| `--format` | Output format (text, json, html) | `text` |
| `--output` | Output file path (empty for stdout) | `` |
| `--interval` | Check interval for continuous monitoring | `60s` |
| `--metrics-port` | Prometheus metrics port | `8080` |
| `--one-shot` | Run once and exit | `false` |

## API and Programming Interface

### Health Check API

```go
package main

import (
    "context"
    "fmt"
    "github.com/ochestra-tech/ochestra-ai/pkg/health"
)

func main() {
    clientset, metricsClient := initKubernetesClients()
    
    healthData, err := health.GetClusterHealth(
        context.Background(), 
        clientset, 
        metricsClient,
    )
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Cluster Health Score: %d/100\n", healthData.HealthScore)
}
```

### Cost Analysis API

```go
package main

import (
    "context"
    "github.com/ochestra-tech/ochestra-ai/pkg/cost"
)

func main() {
    clientset, metricsClient := initKubernetesClients()
    pricing := loadPricingConfig()
    
    nodeCosts, err := cost.GetNodeCosts(
        context.Background(),
        clientset,
        metricsClient,
        pricing,
    )
    if err != nil {
        panic(err)
    }
    
    for _, node := range nodeCosts {
        fmt.Printf("Node %s: $%.2f/hour\n", node.Name, node.TotalCost)
    }
}
```

### Report Generation API

```go
package main

import (
    "context"
    "os"
    "github.com/ochestra-tech/ochestra-ai/pkg/reports"
)

func main() {
    clientset, metricsClient := initKubernetesClients()
    pricing := loadPricingConfig()
    
    generator := reports.NewReportGenerator(
        clientset,
        metricsClient,
        reports.FormatHTML,
        os.Stdout,
    )
    
    err := generator.GenerateCombinedReport(context.Background(), pricing)
    if err != nil {
        panic(err)
    }
}
```

## Prometheus Metrics

The tool exports the following Prometheus metrics:

| Metric | Type | Description |
|--------|------|-------------|
| `k8s_health_manager_node_status` | Gauge | Node readiness status |
| `k8s_health_manager_pod_status` | Gauge | Pod status by namespace |
| `k8s_health_manager_namespace_resource_usage` | Gauge | Resource usage by namespace |
| `k8s_health_manager_namespace_cost` | Gauge | Cost per namespace per hour |
| `k8s_health_manager_resource_efficiency` | Gauge | Resource efficiency ratio |

### Grafana Dashboard

You can create Grafana dashboards using these metrics:

```promql
# Cluster health score
k8s_health_manager_cluster_health_score

# Cost per namespace
k8s_health_manager_namespace_cost

# Resource efficiency
k8s_health_manager_resource_efficiency
```


## Examples

### Example Output

#### Health Report (Text)
```
=== Kubernetes Cluster Health Report ===
Generated at: 2024-01-15T10:30:00Z

Overall Health Score: 85/100

--- Node Health ---
Total Nodes:                    3
Ready Nodes:                    3
Memory Pressure Nodes:          0
Disk Pressure Nodes:            0
PID Pressure Nodes:             0
Network Unavailable Nodes:      0
Average Node Load:              45.2

--- Pod Health ---
Total Pods:                     48
Running Pods:                   45
Pending Pods:                   2
Failed Pods:                    1
Restarting Pods:                0
Crash Looping Pods:             0

--- Control Plane Status ---
API Server Healthy:             true
Controller Manager Healthy:     true
Scheduler Healthy:              true
Etcd Healthy:                   true
CoreDNS Healthy:                true
API Server Latency:             12.5 ms

--- Resource Usage ---
Cluster CPU Usage:              65.2%
Cluster Memory Usage:           72.8%
Cluster Storage Usage:          45.1%
```

#### Cost Report (Text)
```
=== Kubernetes Cluster Cost Report ===
Generated at: 2024-01-15T10:30:00Z

Total Hourly Cost:              $12.45
Total Monthly Cost:             $8,964.00

--- Node Cost Summary ---
┌──────────────────┬──────────────┬─────────────┬───────────┬─────────────┬─────────────┐
│ Node             │ Instance Type │ Hourly Cost │ CPU Cost  │ Memory Cost │ Utilization │
├──────────────────┼──────────────┼─────────────┼───────────┼─────────────┼─────────────┤
│ node-1           │ m5.large     │ $4.15       │ $2.88     │ $1.27       │ 68.5%       │
│ node-2           │ m5.large     │ $4.15       │ $2.88     │ $1.27       │ 71.2%       │
│ node-3           │ c5.large     │ $4.15       │ $3.15     │ $1.00       │ 59.8%       │
└──────────────────┴──────────────┴─────────────┴───────────┴─────────────┴─────────────┘

--- Namespace Cost Summary ---
┌─────────────────┬──────────────┬───────────┬─────────────┬───────────┐
│ Namespace       │ Monthly Cost │ CPU Cost  │ Memory Cost │ Pod Count │
├─────────────────┼──────────────┼───────────┼─────────────┼───────────┤
│ production      │ $4,234.80    │ $2,876.40 │ $1,358.40   │ 24        │
│ staging         │ $2,156.40    │ $1,438.20 │ $718.20     │ 12        │
│ monitoring      │ $1,892.80    │ $1,254.60 │ $638.20     │ 8         │
└─────────────────┴──────────────┴───────────┴─────────────┴───────────┘
```

## Contributing

### Development Setup

1. Fork the repository
2. Clone your fork: `git clone https://github.com/your-username/ochestra-ai.git`
3. Create a feature branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Add tests for new functionality
6. Run tests: `go test ./...`
7. Create a pull request

### Code Structure

```
.
├── cmd/
│   └── main.go                # Application entry point
├── pkg/
│   ├── health/
│   │   └── health-checker.go  # Health monitoring utilities
│   ├── cost/
│   │   └── cost-tracker.go    # Cost calculation utilities
│   └── reports/
│       └── generator.go       # Report generation
├── examples/
│   └── main.go                # Usage examples
├── configs/
│   └── pricing-config.json    # Default pricing configuration
├── deployments/
│   └── kubernetes.yaml        # Kubernetes deployment manifests
└── README.md
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/health/
```

## Troubleshooting

### Common Issues

#### 1. Permission Denied
```
Error: failed to list nodes: nodes is forbidden
```
**Solution**: Ensure your service account has the required RBAC permissions (see Kubernetes Deployment section).

#### 2. Metrics Server Not Found
```
Error: failed to get pod metrics: the server could not find the requested resource
```
**Solution**: Install metrics-server in your cluster:
```bash
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
```

#### 3. Invalid Pricing Configuration
```
Error: failed to parse pricing config
```
**Solution**: Validate your `pricing-config.json` file format against the example provided.

### Debug Mode

Enable debug logging:
```bash
./ochestra-ai --debug --type health
```

### Log Analysis

Check application logs for detailed error information:
```bash
# For container deployment
kubectl logs -n monitoring deployment/ochestra-ai

# For local deployment
./ochestra-ai 2>&1 | tee app.log
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/ochestra-tech/ochestra-ai/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ochestra-tech/ochestra-ai/discussions)
- **Documentation**: [Wiki](https://github.com/ochestra-tech/ochestra-ai/wiki)

## Roadmap

- [ ] **Multi-cluster support**: Monitor multiple clusters from a single instance (KubeCostGuard Project)
- [ ] **Historical data storage**: Store metrics in time-series database (KubeOpera Project)
- [ ] **Advanced forecasting**: ML-based cost prediction
- [ ] **Cloud provider integration**: Direct billing API integration (KubeCostGuard Project)
- [ ] **Slack/Teams notifications**: Real-time alerts
- [ ] **Helm chart**: Easy deployment with Helm (KubeCostGuard Project)
- [ ] **Web UI**: Built-in web interface for centralized multi-cluster monitoring & observability (KubeCostOpera Project)

## Acknowledgments

- [Kubernetes client-go](https://github.com/kubernetes/client-go) - Kubernetes API client
- [Prometheus client](https://github.com/prometheus/client_golang) - Metrics collection
- [TableWriter](https://github.com/olekukonko/tablewriter) - Beautiful table output

---

*Built with ❤️ for the Kubernetes community*
