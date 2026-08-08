package k8s

import (
	"context"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodContainers returns the container names (including init containers) of a pod.
func (c *Client) PodContainers(ctx context.Context, ns, name string) ([]string, error) {
	pod, err := c.clientset.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(pod.Spec.Containers)+len(pod.Spec.InitContainers))
	for _, ic := range pod.Spec.InitContainers {
		names = append(names, ic.Name)
	}
	for _, ctr := range pod.Spec.Containers {
		names = append(names, ctr.Name)
	}
	return names, nil
}

// OpenLogStream returns a read stream for a pod's logs.
func (c *Client) OpenLogStream(ctx context.Context, ns, name, container string, follow bool, tail int, timestamps bool) (io.ReadCloser, error) {
	opts := &corev1.PodLogOptions{
		Container:  container,
		Follow:     follow,
		Timestamps: timestamps,
	}
	if tail > 0 {
		t := int64(tail)
		opts.TailLines = &t
	}
	req := c.clientset.CoreV1().Pods(ns).GetLogs(name, opts)
	return req.Stream(ctx)
}
