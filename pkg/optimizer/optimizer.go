package optimizer

import (
	"context"
	"fmt"
	"log"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type OptimizationReport struct {
	PotentialSavings float64
	Recommendations  []Recommendation
}

type Recommendation struct {
	Type            string
	Description     string
	PotentialSaving float64
}

type ResourceOptimizer struct {
	clientset     *kubernetes.Clientset
	metricsClient *versioned.Clientset
}

func (o *ResourceOptimizer) GenerateOptimizationReport(ctx context.Context) (*OptimizationReport, error) {
	return &OptimizationReport{
		PotentialSavings: 0.0,
		Recommendations:  []Recommendation{},
	}, nil
}

func NewResourceOptimizer(clientset *kubernetes.Clientset, metricsClient *versioned.Clientset) *ResourceOptimizer {
	return &ResourceOptimizer{
		clientset:     clientset,
		metricsClient: metricsClient,
	}
}

func initKubernetesClients() (*kubernetes.Clientset, *versioned.Clientset) {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	metricsClient, err := versioned.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return clientset, metricsClient
}

type CleanupRecommendation struct {
	ResourceType string
	Namespace    string
	Name         string
	Reason       string
	Age          time.Duration
}

func CleanupUnusedResources(ctx context.Context, clientset *kubernetes.Clientset, dryRun bool) ([]CleanupRecommendation, error) {
	recommendations := make([]CleanupRecommendation, 0)

	// Find unused ConfigMaps
	configMaps, err := clientset.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list configmaps: %w", err)
	}

	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	// Create a map of configmaps in use
	configMapsInUse := make(map[string]bool)
	for _, pod := range pods.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.ConfigMap != nil {
				key := fmt.Sprintf("%s/%s", pod.Namespace, volume.ConfigMap.Name)
				configMapsInUse[key] = true
			}
		}

		for _, container := range pod.Spec.Containers {
			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil {
					key := fmt.Sprintf("%s/%s", pod.Namespace, env.ValueFrom.ConfigMapKeyRef.Name)
					configMapsInUse[key] = true
				}
			}
		}
	}

	// Find unused configmaps
	for _, cm := range configMaps.Items {
		key := fmt.Sprintf("%s/%s", cm.Namespace, cm.Name)
		if !configMapsInUse[key] {
			rec := CleanupRecommendation{
				ResourceType: "ConfigMap",
				Namespace:    cm.Namespace,
				Name:         cm.Name,
				Reason:       "Not referenced by any pod",
				Age:          time.Since(cm.CreationTimestamp.Time),
			}
			recommendations = append(recommendations, rec)

			if !dryRun {
				// Delete unused configmap
				err := clientset.CoreV1().ConfigMaps(cm.Namespace).Delete(ctx, cm.Name, metav1.DeleteOptions{})
				if err != nil {
					log.Printf("Failed to delete configmap %s/%s: %v", cm.Namespace, cm.Name, err)
				} else {
					log.Printf("Deleted unused configmap %s/%s", cm.Namespace, cm.Name)
				}
			}
		}
	}

	// Find failed pods older than 7 days
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded {
			age := time.Since(pod.CreationTimestamp.Time)
			if age > 7*24*time.Hour {
				rec := CleanupRecommendation{
					ResourceType: "Pod",
					Namespace:    pod.Namespace,
					Name:         pod.Name,
					Reason:       fmt.Sprintf("Failed/Completed pod older than 7 days (status: %s)", pod.Status.Phase),
					Age:          age,
				}
				recommendations = append(recommendations, rec)

				if !dryRun {
					// Delete old failed/succeeded pod
					err := clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
					if err != nil {
						log.Printf("Failed to delete pod %s/%s: %v", pod.Namespace, pod.Name, err)
					} else {
						log.Printf("Deleted old pod %s/%s", pod.Namespace, pod.Name)
					}
				}
			}
		}
	}
	return []CleanupRecommendation{}, nil
}

func main() {
	clientset, metricsClient := initKubernetesClients()

	// Run resource optimization analysis
	optimizer := NewResourceOptimizer(clientset, metricsClient)
	report, err := optimizer.GenerateOptimizationReport(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Potential Monthly Savings: $%.2f\n", report.PotentialSavings)
	fmt.Printf("Optimization Recommendations: %d\n", len(report.Recommendations))

	for _, rec := range report.Recommendations {
		fmt.Printf("- %s: %s (Save $%.2f/month)\n",
			rec.Type, rec.Description, rec.PotentialSaving)
	}

	// Run cleanup with dry-run
	cleanupRecs, err := CleanupUnusedResources(context.Background(), clientset, true)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nCleanup Recommendations: %d\n", len(cleanupRecs))
	for _, rec := range cleanupRecs {
		fmt.Printf("- Delete %s %s/%s: %s\n",
			rec.ResourceType, rec.Namespace, rec.Name, rec.Reason)
	}
}
