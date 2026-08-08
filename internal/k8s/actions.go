package k8s

import (
	"context"
	"errors"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// Delete removes the named resource.
func (c *Client) Delete(ctx context.Context, r Resource, ns, name string) error {
	var resource dynamic.ResourceInterface
	dr := c.dyn.Resource(r.GVR)
	if r.Namespaced {
		resource = dr.Namespace(ns)
	} else {
		resource = dr.Namespace("")
	}
	return resource.Delete(ctx, name, metav1.DeleteOptions{})
}

// Scale sets the replica count on workloads that support it.
func (c *Client) Scale(ctx context.Context, r Resource, ns, name string, replicas int32) error {
	switch r.Plural {
	case "deployments":
		d, err := c.clientset.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		d.Spec.Replicas = &replicas
		_, err = c.clientset.AppsV1().Deployments(ns).Update(ctx, d, metav1.UpdateOptions{})
		return err
	case "statefulsets":
		s, err := c.clientset.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		s.Spec.Replicas = &replicas
		_, err = c.clientset.AppsV1().StatefulSets(ns).Update(ctx, s, metav1.UpdateOptions{})
		return err
	default:
		return fmt.Errorf("scaling %q is not supported", r.Title)
	}
}

// Restart performs a rollout restart on workloads that support it.
func (c *Client) Restart(ctx context.Context, r Resource, ns, name string) error {
	restartedAt := time.Now().Format(time.RFC3339)
	patch := []byte(fmt.Sprintf(
		`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":%q}}}}}`,
		restartedAt,
	))
	switch r.Plural {
	case "deployments":
		_, err := c.clientset.AppsV1().Deployments(ns).Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
		return err
	case "statefulsets":
		_, err := c.clientset.AppsV1().StatefulSets(ns).Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
		return err
	case "daemonsets":
		_, err := c.clientset.AppsV1().DaemonSets(ns).Patch(ctx, name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
		return err
	default:
		return fmt.Errorf("restarting %q is not supported", r.Title)
	}
}

// SupportsScale reports whether the resource supports scaling.
func (r Resource) SupportsScale() bool {
	return r.Plural == "deployments" || r.Plural == "statefulsets"
}

// SupportsRestart reports whether the resource supports rollout restart.
func (r Resource) SupportsRestart() bool {
	return r.Plural == "deployments" || r.Plural == "statefulsets" || r.Plural == "daemonsets"
}

// SupportsLogs reports whether the resource exposes logs.
func (r Resource) SupportsLogs() bool {
	return r.Plural == "pods"
}

// SupportsTop reports whether the resource supports `kubectl top` usage.
func (r Resource) SupportsTop() bool {
	return r.Plural == "pods" || r.Plural == "nodes"
}

var errEditNotSupported = errors.New("edit is not supported for this resource")
