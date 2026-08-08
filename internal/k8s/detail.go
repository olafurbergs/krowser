package k8s

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

// YAML returns the object serialized as YAML.
func (c *Client) YAML(ctx context.Context, r Resource, ns, name string) (string, error) {
	obj, err := c.Get(ctx, r, ns, name)
	if err != nil {
		return "", err
	}
	// Strip status-only noise? Keep full manifest for fidelity.
	out, err := yaml.Marshal(obj.Object)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// Describe returns a human-friendly description of a resource. Pods get a rich
// summary; all other resources fall back to their YAML manifest.
func (c *Client) Describe(ctx context.Context, r Resource, ns, name string) (string, error) {
	if r.Plural == "pods" {
		return c.describePod(ctx, ns, name)
	}
	return c.YAML(ctx, r, ns, name)
}

func (c *Client) describePod(ctx context.Context, ns, name string) (string, error) {
	pod, err := c.clientset.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Name:         %s\n", pod.Name)
	fmt.Fprintf(&b, "Namespace:    %s\n", pod.Namespace)
	if pod.Spec.NodeName != "" {
		fmt.Fprintf(&b, "Node:         %s\n", pod.Spec.NodeName)
	}
	fmt.Fprintf(&b, "Start Time:   %s\n", pod.CreationTimestamp.Time.Format("Mon, 02 Jan 2006 15:04:05 -0700"))
	fmt.Fprintf(&b, "Labels:       %s\n", formatMap(pod.Labels))
	fmt.Fprintf(&b, "Status:       %s\n", string(pod.Status.Phase))
	if pod.Status.PodIP != "" {
		fmt.Fprintf(&b, "IP:           %s\n", pod.Status.PodIP)
	}
	if pod.Spec.PriorityClassName != "" {
		fmt.Fprintf(&b, "Priority:     %s\n", pod.Spec.PriorityClassName)
	}

	if len(pod.Status.ContainerStatuses) > 0 {
		b.WriteString("\nContainers:\n")
		for _, cs := range pod.Status.ContainerStatuses {
			fmt.Fprintf(&b, "  %s:\n", cs.Name)
			fmt.Fprintf(&b, "    Ready:       %v\n", cs.Ready)
			fmt.Fprintf(&b, "    Restart Count: %d\n", cs.RestartCount)
			fmt.Fprintf(&b, "    Image:       %s\n", cs.Image)
			state := describeContainerState(cs.State)
			fmt.Fprintf(&b, "    State:       %s\n", state)
			if cs.LastTerminationState.Terminated != nil {
				fmt.Fprintf(&b, "    Last State:  %s\n", cs.LastTerminationState.Terminated.Reason)
			}
		}
	}

	if len(pod.Status.Conditions) > 0 {
		b.WriteString("\nConditions:\n")
		for _, cond := range pod.Status.Conditions {
			fmt.Fprintf(&b, "  %-20s %-12s %s\n", cond.Type, cond.Status, cond.Message)
		}
	}

	if events := c.podEvents(ctx, ns, name); len(events) > 0 {
		b.WriteString("\nEvents:\n")
		for _, e := range events {
			fmt.Fprintf(&b, "  %-16s %-6s %s\n", e.LastTimestamp.Time.Format("15:04:05"), e.Type, e.Message)
		}
	}

	return b.String(), nil
}

func describeContainerState(state corev1.ContainerState) string {
	switch {
	case state.Running != nil:
		return fmt.Sprintf("Running (started %s)", state.Running.StartedAt.Time.Format("15:04:05"))
	case state.Waiting != nil:
		return fmt.Sprintf("Waiting (%s)", state.Waiting.Reason)
	case state.Terminated != nil:
		return fmt.Sprintf("Terminated (%s)", state.Terminated.Reason)
	default:
		return "Unknown"
	}
}

func (c *Client) podEvents(ctx context.Context, ns, name string) []corev1.Event {
	list, err := c.clientset.CoreV1().Events(ns).List(ctx, metav1.ListOptions{
		FieldSelector: "involvedObject.name=" + name,
	})
	if err != nil {
		return nil
	}
	return list.Items
}

func formatMap(m map[string]string) string {
	if len(m) == 0 {
		return "<none>"
	}
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, ",")
}
