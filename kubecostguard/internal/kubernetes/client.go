// 6. Kubernetes API Integration

// File: internal/kubernetes/client.go
package kubernetes

import (
	"context"
	"fmt"

	"github.com/ochestra-tech/kubecostguard/internal/config"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
	metricsv1beta1 "k8s.io/metrics/pkg/client/clientset/versioned/typed/metrics/v1beta1"
)

// Client provides access to the Kubernetes API
type Client struct {
	clientset     *kubernetes.Clientset
	metricsClient *metricsv.Clientset
	metrics       *metricsv1beta1.MetricsV1beta1Client
	config        config.KubernetesConfig
}

func (c *Client) CoreV1() {
	panic("unimplemented")
}

func (c *Client) AppsV1() {
	panic("unimplemented")
}

// NewClient creates a new Kubernetes client
func NewClient(config config.KubernetesConfig) (*Client, error) {
	var restConfig *rest.Config
	var err error

	if config.InCluster {
		// Use in-cluster config when running in a pod
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	} else {
		// Use kubeconfig file
		restConfig, err = clientcmd.BuildConfigFromFlags("", config.Kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
		}
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Create metrics client
	metricsClient, err := metricsv.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %w", err)
	}

	return &Client{
		clientset:     clientset,
		metricsClient: metricsClient,
		config:        config,
	}, nil
}

// GetNodes returns all nodes in the cluster
func (c *Client) GetNodes(ctx context.Context) (*corev1.NodeList, error) {
	return c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
}

// GetPods returns all pods in the cluster or in a specific namespace
func (c *Client) GetPods(ctx context.Context, namespace string) (*corev1.PodList, error) {
	if namespace != "" {
		return c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	}
	return c.clientset.CoreV1().Pods(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
}

// GetClusterResources retrieves all resource information for cost analysis
func (c *Client) GetClusterResources() (map[string]interface{}, error) {
	ctx := context.Background()

	// Get nodes
	nodes, err := c.GetNodes(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes: %w", err)
	}

	// Get pods
	pods, err := c.GetPods(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	// Get node metrics
	nodeMetrics, err := c.metricsClient.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node metrics: %w", err)
	}

	// Get pod metrics
	podMetrics, err := c.metricsClient.MetricsV1beta1().PodMetricses(corev1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	// Build resource map
	resources := map[string]interface{}{
		"nodes":       nodes.Items,
		"pods":        pods.Items,
		"nodeMetrics": nodeMetrics.Items,
		"podMetrics":  podMetrics.Items,
	}

	return resources, nil
}
