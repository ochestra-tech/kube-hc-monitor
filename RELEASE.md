ochestra.ai is a comprehensive Kubernetes health and cost management tool with the following components:

Main Application (main.go): Core application that provides:

Health monitoring with metrics and alerts
Cost tracking with pricing configuration
Prometheus metrics integration
Continuous monitoring or one-shot execution
Multiple output formats (JSON, HTML, text)

Health Utilities (pkg/health/utils.go): Monitors:

Node health (ready status, pressure conditions)
Pod health (running, failed, crash-looping pods)
Control plane status (API server, etcd, scheduler)
Network health (CNI, DNS, services)
Resource utilization (CPU, memory, storage)
Health scoring system

Cost Management (pkg/cost/utils.go): Tracks:

Node costs by instance type
Pod costs by resource usage
Namespace cost aggregation
Cost optimization recommendations
Resource efficiency metrics

Report Generator (pkg/reports/generator.go): Provides:

JSON, HTML, and text report formats
Health and cost report generation
Combined reports with optimization recommendations
Visual dashboards

Usage Example (example/main.go): Shows:

Configuration and initialization
Cost forecasting
Resource optimization
Cleanup automation
Advanced monitoring features

Key features include:

Real-time health monitoring with health scores
Cost tracking with pricing models
Optimization recommendations
Prometheus metrics integration
Multiple output formats
Flexible configuration
Resource cleanup automation
Cost forecasting
Comprehensive reporting

To use this tool:

Build the application:

bashgo build -o k8s-health-cost main.go

Run a health check:

bash./k8s-health-cost --type health --format text

Generate a cost report:

bash./k8s-health-cost --type cost --format html --output cost-report.html

Monitor continuously:

bash./k8s-health-cost --type combined --interval 5m --metrics-port 8080
The tool provides comprehensive monitoring and cost management for Kubernetes clusters, helping optimize resource usage and identify issues proactively.
