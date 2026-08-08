package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TopEntry is a single pod or node usage sample. For pods the request and
// limit figures are the summed values from the pod spec; for nodes they
// represent the allocatable capacity.
type TopEntry struct {
	Name      string
	Namespace string

	CPU float64 // millicores used
	Mem float64 // bytes used

	CPUReq float64 // millicores requested, 0 = unset
	CPULim float64 // millicores limited, 0 = unset
	MemReq float64 // bytes requested, 0 = unset
	MemLim float64 // bytes limited, 0 = unset
}

// Top reports current CPU/memory usage for pods or nodes, similar to
// `kubectl top`. An empty namespace means all namespaces.
func (c *Client) Top(ctx context.Context, r Resource, ns string) ([]TopEntry, error) {
	switch r.Plural {
	case "pods":
		return c.topPods(ctx, ns)
	case "nodes":
		return c.topNodes(ctx)
	default:
		return nil, fmt.Errorf("top is only supported for pods and nodes")
	}
}

func (c *Client) topPods(ctx context.Context, ns string) ([]TopEntry, error) {
	metrics, err := c.metricsCli.MetricsV1beta1().PodMetricses(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	pods, err := c.clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	specs := make(map[string]*corev1.Pod, len(pods.Items))
	for i := range pods.Items {
		p := &pods.Items[i]
		specs[p.Namespace+"/"+p.Name] = p
	}

	entries := make([]TopEntry, 0, len(metrics.Items))
	for _, pm := range metrics.Items {
		e := TopEntry{Name: pm.Name, Namespace: pm.Namespace}
		for _, ct := range pm.Containers {
			e.CPU += milliValue(ct.Usage[corev1.ResourceCPU])
			e.Mem += byteValue(ct.Usage[corev1.ResourceMemory])
		}
		if pod, ok := specs[pm.Namespace+"/"+pm.Name]; ok {
			req, lim := podResources(pod)
			e.CPUReq, e.MemReq = req.cpu, req.mem
			e.CPULim, e.MemLim = lim.cpu, lim.mem
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (c *Client) topNodes(ctx context.Context) ([]TopEntry, error) {
	metrics, err := c.metricsCli.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	allocs := make(map[string]*corev1.Node, len(nodes.Items))
	for i := range nodes.Items {
		n := &nodes.Items[i]
		allocs[n.Name] = n
	}

	entries := make([]TopEntry, 0, len(metrics.Items))
	for _, nm := range metrics.Items {
		e := TopEntry{Name: nm.Name}
		e.CPU = milliValue(nm.Usage[corev1.ResourceCPU])
		e.Mem = byteValue(nm.Usage[corev1.ResourceMemory])
		if n, ok := allocs[nm.Name]; ok {
			e.CPULim = milliValue(n.Status.Allocatable[corev1.ResourceCPU])
			e.MemLim = byteValue(n.Status.Allocatable[corev1.ResourceMemory])
		}
		entries = append(entries, e)
	}
	return entries, nil
}

type resPair struct{ cpu, mem float64 }

// podResources sums the request and limit resource values across a pod's
// regular and init containers.
func podResources(p *corev1.Pod) (req, lim resPair) {
	containers := make([]corev1.Container, 0, len(p.Spec.Containers)+len(p.Spec.InitContainers))
	containers = append(containers, p.Spec.Containers...)
	containers = append(containers, p.Spec.InitContainers...)
	for _, ct := range containers {
		req.cpu += milliValue(ct.Resources.Requests[corev1.ResourceCPU])
		req.mem += byteValue(ct.Resources.Requests[corev1.ResourceMemory])
		lim.cpu += milliValue(ct.Resources.Limits[corev1.ResourceCPU])
		lim.mem += byteValue(ct.Resources.Limits[corev1.ResourceMemory])
	}
	return req, lim
}

func milliValue(q resource.Quantity) float64 {
	return float64(q.MilliValue())
}

func byteValue(q resource.Quantity) float64 {
	return float64(q.Value())
}
