package health

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ClusterHealth represents overall cluster health status
type ClusterHealth struct {
	Timestamp          time.Time                  `json:"timestamp"`
	NodeStatus         NodeHealthStatus           `json:"nodeStatus"`
	PodStatus          PodHealthStatus            `json:"podStatus"`
	ControlPlaneStatus ControlPlaneStatus         `json:"controlPlaneStatus"`
	NetworkStatus      NetworkStatus              `json:"networkStatus"`
	ResourceUsage      ResourceUsageStatus        `json:"resourceUsage"`
	ComponentStatuses  []ComponentStatus          `json:"componentStatuses"`
	NamespaceHealth    map[string]NamespaceHealth `json:"namespaceHealth"`
	HealthScore        int                        `json:"healthScore"` // 0-100
	Issues             []HealthIssue              `json:"issues"`
}

// NodeHealthStatus contains node health information
type NodeHealthStatus struct {
	TotalNodes              int                 `json:"totalNodes"`
	ReadyNodes              int                 `json:"readyNodes"`
	MemoryPressureNodes     int                 `json:"memoryPressureNodes"`
	DiskPressureNodes       int                 `json:"diskPressureNodes"`
	PIDPressureNodes        int                 `json:"pidPressureNodes"`
	NetworkUnavailableNodes int                 `json:"networkUnavailableNodes"`
	NodeConditions          map[string][]string `json:"nodeConditions"` // Node name -> conditions
	AverageLoad             float64             `json:"averageLoad"`
}

// PodHealthStatus contains pod health information
type PodHealthStatus struct {
	TotalPods        int            `json:"totalPods"`
	RunningPods      int            `json:"runningPods"`
	PendingPods      int            `json:"pendingPods"`
	SucceededPods    int            `json:"succeededPods"`
	FailedPods       int            `json:"failedPods"`
	UnknownPods      int            `json:"unknownPods"`
	RestartingPods   int            `json:"restartingPods"`
	PodsPerNode      map[string]int `json:"podsPerNode"`
	CrashLoopingPods []string       `json:"crashLoopingPods"`
}

// ControlPlaneStatus contains control plane health information
type ControlPlaneStatus struct {
	APIServerHealthy  bool    `json:"apiServerHealthy"`
	ControllerHealthy bool    `json:"controllerHealthy"`
	SchedulerHealthy  bool    `json:"schedulerHealthy"`
	EtcdHealthy       bool    `json:"etcdHealthy"`
	CoreDNSHealthy    bool    `json:"coreDNSHealthy"`
	OverallHealthy    bool    `json:"overallHealthy"`
	APIServerLatency  float64 `json:"apiServerLatency"` // in milliseconds
}

// NetworkStatus contains network health information
type NetworkStatus struct {
	CNIHealthy              bool `json:"cniHealthy"`
	DNSResolutionOK         bool `json:"dnsResolutionOK"`
	ServiceEndpointsHealthy bool `json:"serviceEndpointsHealthy"`
	IngressHealthy          bool `json:"ingressHealthy"`
	NetworkPoliciesCount    int  `json:"networkPoliciesCount"`
}

// ResourceUsageStatus contains resource usage information
type ResourceUsageStatus struct {
	ClusterCPUUsage     float64  `json:"clusterCPUUsage"`     // percentage
	ClusterMemoryUsage  float64  `json:"clusterMemoryUsage"`  // percentage
	ClusterStorageUsage float64  `json:"clusterStorageUsage"` // percentage
	HighCPUNodes        []string `json:"highCPUNodes"`
	HighMemoryNodes     []string `json:"highMemoryNodes"`
	LowResourceNodes    []string `json:"lowResourceNodes"`
	HighUsageNamespaces []string `json:"highUsageNamespaces"`
}

// ComponentStatus represents a cluster component's health
type ComponentStatus struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Message string `json:"message,omitempty"`
	Version string `json:"version,omitempty"`
}

// NamespaceHealth contains health information for a namespace
type NamespaceHealth struct {
	PodStatus        PodHealthStatus     `json:"podStatus"`
	DeploymentStatus DeploymentStatus    `json:"deploymentStatus"`
	ServiceStatus    ServiceStatus       `json:"serviceStatus"`
	ResourceUsage    ResourceUsageStatus `json:"resourceUsage"`
	HealthScore      int                 `json:"healthScore"` // 0-100
}

// DeploymentStatus contains deployment health information
type DeploymentStatus struct {
	TotalDeployments       int `json:"totalDeployments"`
	HealthyDeployments     int `json:"healthyDeployments"`
	ProgressingDeployments int `json:"progressingDeployments"`
	FailedDeployments      int `json:"failedDeployments"`
}

// ServiceStatus contains service health information
type ServiceStatus struct {
	TotalServices            int `json:"totalServices"`
	ServicesWithEndpoints    int `json:"servicesWithEndpoints"`
	ServicesWithoutEndpoints int `json:"servicesWithoutEndpoints"`
}

// HealthIssue represents a detected health issue
type HealthIssue struct {
	Severity   string    `json:"severity"` // "critical", "warning", "info"
	Resource   string    `json:"resource"`
	Namespace  string    `json:"namespace,omitempty"`
	Name       string    `json:"name,omitempty"`
	Message    string    `json:"message"`
	Timestamp  time.Time `json:"timestamp"`
	Suggestion string    `json:"suggestion,omitempty"`
}

// GetClusterHealth performs a comprehensive health check of the Kubernetes cluster
func GetClusterHealth(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	metricsClient *metricsv.Clientset,
) (*ClusterHealth, error) {
	health := &ClusterHealth{
		Timestamp:       time.Now(),
		NamespaceHealth: make(map[string]NamespaceHealth),
		Issues:          make([]HealthIssue, 0),
	}

	// Check node health
	if err := checkNodeHealth(ctx, clientset, &health.NodeStatus); err != nil {
		return nil, fmt.Errorf("node health check failed: %w", err)
	}

	// Check pod health
	if err := checkPodHealth(ctx, clientset, &health.PodStatus); err != nil {
		return nil, fmt.Errorf("pod health check failed: %w", err)
	}

	// Check control plane health
	if err := checkControlPlaneHealth(ctx, clientset, &health.ControlPlaneStatus); err != nil {
		log.Printf("Control plane health check failed: %v", err)
		// Continue with partial data
	}

	// Check network health
	if err := checkNetworkHealth(ctx, clientset, &health.NetworkStatus); err != nil {
		log.Printf("Network health check failed: %v", err)
		// Continue with partial data
	}

	// Check resource usage
	if err := checkResourceUsage(ctx, clientset, metricsClient, &health.ResourceUsage); err != nil {
		log.Printf("Resource usage check failed: %v", err)
		// Continue with partial data
	}

	// Check component statuses
	if err := checkComponentStatuses(ctx, clientset, &health.ComponentStatuses); err != nil {
		log.Printf("Component status check failed: %v", err)
		// Continue with partial data
	}

	// Check namespace health
	if err := checkNamespaceHealth(ctx, clientset, metricsClient, health); err != nil {
		log.Printf("Namespace health check failed: %v", err)
		// Continue with partial data
	}

	// Identify health issues
	identifyHealthIssues(health)

	// Calculate overall health score
	health.HealthScore = calculateHealthScore(health)

	return health, nil
}

// checkNodeHealth checks the health status of all nodes
func checkNodeHealth(ctx context.Context, clientset *kubernetes.Clientset, status *NodeHealthStatus) error {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	status.TotalNodes = len(nodes.Items)
	status.NodeConditions = make(map[string][]string)
	totalLoad := 0.0

	for _, node := range nodes.Items {
		isReady := false
		nodeConditions := make([]string, 0)

		for _, condition := range node.Status.Conditions {
			if condition.Status == v1.ConditionTrue {
				nodeConditions = append(nodeConditions, string(condition.Type))

				switch condition.Type {
				case v1.NodeReady:
					isReady = true
					status.ReadyNodes++
				case v1.NodeMemoryPressure:
					status.MemoryPressureNodes++
				case v1.NodeDiskPressure:
					status.DiskPressureNodes++
				case v1.NodePIDPressure:
					status.PIDPressureNodes++
				case v1.NodeNetworkUnavailable:
					status.NetworkUnavailableNodes++
				}
			}
		}

		status.NodeConditions[node.Name] = nodeConditions

		// Get node load (simplified)
		for _, metric := range node.Status.Allocatable {
			totalLoad += float64(metric.MilliValue()) / 1000
		}
	}

	if status.TotalNodes > 0 {
		status.AverageLoad = totalLoad / float64(status.TotalNodes)
	}

	return nil
}

// checkPodHealth checks the health status of all pods
func checkPodHealth(ctx context.Context, clientset *kubernetes.Clientset, status *PodHealthStatus) error {
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	status.TotalPods = len(pods.Items)
	status.PodsPerNode = make(map[string]int)
	status.CrashLoopingPods = make([]string, 0)

	for _, pod := range pods.Items {
		// Update pod count per node
		nodeName := pod.Spec.NodeName
		if nodeName != "" {
			if _, exists := status.PodsPerNode[nodeName]; !exists {
				status.PodsPerNode[nodeName] = 0
			}
			status.PodsPerNode[nodeName]++
		}

		// Update pod phase counts
		switch pod.Status.Phase {
		case v1.PodRunning:
			status.RunningPods++
		case v1.PodPending:
			status.PendingPods++
		case v1.PodSucceeded:
			status.SucceededPods++
		case v1.PodFailed:
			status.FailedPods++
		default:
			status.UnknownPods++
		}

		// Check for restarting pods
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.RestartCount > 5 {
				status.RestartingPods++
			}

			// Check for crash loop back off
			if containerStatus.State.Waiting != nil &&
				containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
				podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
				status.CrashLoopingPods = append(status.CrashLoopingPods, podKey)
			}
		}
	}

	return nil
}

// checkControlPlaneHealth checks the health of control plane components
func checkControlPlaneHealth(ctx context.Context, clientset *kubernetes.Clientset, status *ControlPlaneStatus) error {
	// Check API server
	startTime := time.Now()
	_, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	apiCallDuration := time.Since(startTime)

	status.APIServerLatency = float64(apiCallDuration.Milliseconds())
	status.APIServerHealthy = err == nil && apiCallDuration < 1*time.Second

	// Check kube-system components
	pods, err := clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list kube-system pods: %w", err)
	}

	status.ControllerHealthy = true
	status.SchedulerHealthy = true
	status.EtcdHealthy = true
	status.CoreDNSHealthy = true

	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, "kube-controller-manager") && pod.Status.Phase != v1.PodRunning {
			status.ControllerHealthy = false
		}
		if strings.Contains(pod.Name, "kube-scheduler") && pod.Status.Phase != v1.PodRunning {
			status.SchedulerHealthy = false
		}
		if strings.Contains(pod.Name, "etcd") && pod.Status.Phase != v1.PodRunning {
			status.EtcdHealthy = false
		}
		if strings.Contains(pod.Name, "coredns") && pod.Status.Phase != v1.PodRunning {
			status.CoreDNSHealthy = false
		}
	}

	status.OverallHealthy = status.APIServerHealthy && status.ControllerHealthy &&
		status.SchedulerHealthy && status.EtcdHealthy && status.CoreDNSHealthy

	return nil
}

// checkNetworkHealth checks the health of network components
func checkNetworkHealth(ctx context.Context, clientset *kubernetes.Clientset, status *NetworkStatus) error {
	// Check CNI pods (assuming they're in kube-system)
	cniPods, err := clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "k8s-app in (calico-node,flannel,weave-net,cilium)",
	})

	if err != nil {
		log.Printf("Failed to check CNI pods: %v", err)
		status.CNIHealthy = false
	} else {
		status.CNIHealthy = true
		for _, pod := range cniPods.Items {
			if pod.Status.Phase != v1.PodRunning {
				status.CNIHealthy = false
				break
			}
		}
	}

	// Check DNS resolution - CoreDNS
	coredns, err := clientset.CoreV1().Pods("kube-system").List(ctx, metav1.ListOptions{
		LabelSelector: "k8s-app=kube-dns",
	})

	if err != nil {
		log.Printf("Failed to check CoreDNS pods: %v", err)
		status.DNSResolutionOK = false
	} else {
		status.DNSResolutionOK = true
		for _, pod := range coredns.Items {
			if pod.Status.Phase != v1.PodRunning {
				status.DNSResolutionOK = false
				break
			}
		}
	}

	// Check service endpoints health
	services, err := clientset.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to list services: %v", err)
		status.ServiceEndpointsHealthy = false
	} else {
		status.ServiceEndpointsHealthy = true

		for _, svc := range services.Items {
			if svc.Spec.Selector == nil || len(svc.Spec.Selector) == 0 {
				// Skip services without selectors (e.g., ExternalName)
				continue
			}

			// Check if service has endpoints
			endpoints, err := clientset.CoreV1().Endpoints(svc.Namespace).Get(ctx, svc.Name, metav1.GetOptions{})
			if err != nil || len(endpoints.Subsets) == 0 {
				status.ServiceEndpointsHealthy = false
				break
			}
		}
	}

	// Check Ingress controller
	ingressControllers, err := clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{
		LabelSelector: "app in (ingress-nginx,traefik,istio-ingressgateway)",
	})

	if err != nil {
		log.Printf("Failed to check ingress controllers: %v", err)
		status.IngressHealthy = false
	} else {
		if len(ingressControllers.Items) == 0 {
			// No ingress controller found - might be normal for some clusters
			status.IngressHealthy = true
		} else {
			status.IngressHealthy = true
			for _, ingress := range ingressControllers.Items {
				if ingress.Status.ReadyReplicas < *ingress.Spec.Replicas {
					status.IngressHealthy = false
					break
				}
			}
		}
	}

	// Count network policies
	netpols, err := clientset.NetworkingV1().NetworkPolicies("").List(ctx, metav1.ListOptions{})
	if err != nil {
		log.Printf("Failed to count network policies: %v", err)
	} else {
		status.NetworkPoliciesCount = len(netpols.Items)
	}

	return nil
}
