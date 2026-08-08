package k8s

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Column is a single column in a resource table.
type Column struct {
	Name string
}

// Resource describes how to list and render one Kubernetes resource type.
type Resource struct {
	Key        byte   // numeric/symbol keybinding, 0 = no binding
	Title      string // display title, e.g. "Pods"
	Plural     string // API plural, e.g. "pods"
	GVR        schema.GroupVersionResource
	Namespaced bool
	StatusCol  int // index of the column to color by status, -1 = none
	Columns    []Column
	Extract    func(*unstructured.Unstructured) []string
}

// Row is a single resource instance rendered as table columns.
type Row struct {
	Cols []string
	Obj  *unstructured.Unstructured
}

// Resources is the ordered registry of browsable resources.
var Resources = []Resource{
	{Key: '1', Title: "Namespaces", Plural: "namespaces", GVR: schema.GroupVersionResource{Version: "v1", Resource: "namespaces"}, Namespaced: false, StatusCol: 1,
		Columns: []Column{{Name: "NAME"}, {Name: "STATUS"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), nestedString(o, "status", "phase"), age(o)}
		}},
	{Key: '2', Title: "Nodes", Plural: "nodes", GVR: schema.GroupVersionResource{Version: "v1", Resource: "nodes"}, Namespaced: false, StatusCol: 1,
		Columns: []Column{{Name: "NAME"}, {Name: "STATUS"}, {Name: "ROLES"}, {Name: "VERSION"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), nodeStatus(o), nodeRoles(o), nestedString(o, "status", "nodeInfo", "kubeletVersion"), age(o)}
		}},
	{Key: '3', Title: "Pods", Plural: "pods", GVR: schema.GroupVersionResource{Version: "v1", Resource: "pods"}, Namespaced: true, StatusCol: 2,
		Columns: []Column{{Name: "NAME"}, {Name: "READY"}, {Name: "STATUS"}, {Name: "RESTARTS"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			ready, total := podReady(o)
			restarts := podRestarts(o)
			return []string{name(o), fmt.Sprintf("%d/%d", ready, total), podStatus(o), strconv.Itoa(restarts), age(o)}
		}},
	{Key: '4', Title: "Deployments", Plural: "deployments", GVR: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}, Namespaced: true,
		Columns: []Column{{Name: "NAME"}, {Name: "READY"}, {Name: "UP-TO-DATE"}, {Name: "AVAILABLE"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), readyFraction(o, "replicas"), nestedString(o, "status", "updatedReplicas"), nestedString(o, "status", "availableReplicas"), age(o)}
		}},
	{Key: '5', Title: "StatefulSets", Plural: "statefulsets", GVR: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"}, Namespaced: true,
		Columns: []Column{{Name: "NAME"}, {Name: "READY"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), readyFraction(o, "replicas"), age(o)}
		}},
	{Key: '6', Title: "DaemonSets", Plural: "daemonsets", GVR: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"}, Namespaced: true,
		Columns: []Column{{Name: "NAME"}, {Name: "DESIRED"}, {Name: "CURRENT"}, {Name: "READY"}, {Name: "UP-TO-DATE"}, {Name: "AVAILABLE"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{
				name(o),
				nestedString(o, "status", "desiredNumberScheduled"),
				nestedString(o, "status", "currentNumberScheduled"),
				nestedString(o, "status", "numberReady"),
				nestedString(o, "status", "updatedNumberScheduled"),
				nestedString(o, "status", "numberAvailable"),
				age(o),
			}
		}},
	{Key: '7', Title: "Services", Plural: "services", GVR: schema.GroupVersionResource{Version: "v1", Resource: "services"}, Namespaced: true,
		Columns: []Column{{Name: "NAME"}, {Name: "TYPE"}, {Name: "CLUSTER-IP"}, {Name: "EXTERNAL-IP"}, {Name: "PORT(S)"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), nestedString(o, "spec", "type"), nestedString(o, "spec", "clusterIP"), externalIP(o), servicePorts(o)}
		}},
	{Key: '8', Title: "ConfigMaps", Plural: "configmaps", GVR: schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}, Namespaced: true,
		Columns: []Column{{Name: "NAME"}, {Name: "DATA"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), strconv.Itoa(len(o.Object["data"].(map[string]any))), age(o)}
		}},
	{Key: '9', Title: "Secrets", Plural: "secrets", GVR: schema.GroupVersionResource{Version: "v1", Resource: "secrets"}, Namespaced: true,
		Columns: []Column{{Name: "NAME"}, {Name: "TYPE"}, {Name: "DATA"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), nestedString(o, "type"), strconv.Itoa(len(o.Object["data"].(map[string]any))), age(o)}
		}},
	{Key: '0', Title: "PVCs", Plural: "persistentvolumeclaims", GVR: schema.GroupVersionResource{Version: "v1", Resource: "persistentvolumeclaims"}, Namespaced: true, StatusCol: 1,
		Columns: []Column{{Name: "NAME"}, {Name: "STATUS"}, {Name: "VOLUME"}, {Name: "CAPACITY"}, {Name: "STORAGECLASS"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), nestedString(o, "status", "phase"), nestedString(o, "spec", "volumeName"), pvcCapacity(o), nestedString(o, "spec", "storageClassName"), age(o)}
		}},
	{Key: '-', Title: "Ingresses", Plural: "ingresses", GVR: schema.GroupVersionResource{Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"}, Namespaced: true,
		Columns: []Column{{Name: "NAME"}, {Name: "CLASS"}, {Name: "HOSTS"}, {Name: "ADDRESS"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), nestedString(o, "spec", "ingressClassName"), ingressHosts(o), nestedString(o, "status", "loadBalancer", "ingress"), age(o)}
		}},
	{Key: '=', Title: "Jobs", Plural: "jobs", GVR: schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"}, Namespaced: true,
		Columns: []Column{{Name: "NAME"}, {Name: "COMPLETIONS"}, {Name: "DURATION"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), completions(o), jobDuration(o), age(o)}
		}},
	{Key: ',', Title: "CronJobs", Plural: "cronjobs", GVR: schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "cronjobs"}, Namespaced: true,
		Columns: []Column{{Name: "NAME"}, {Name: "SCHEDULE"}, {Name: "SUSPEND"}, {Name: "ACTIVE"}, {Name: "LAST SCHEDULE"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), nestedString(o, "spec", "schedule"), strconv.FormatBool(nestedBool(o, "spec", "suspend")), nestedString(o, "status", "active"), lastSchedule(o), age(o)}
		}},
	{Key: '.', Title: "PersistentVolumes", Plural: "persistentvolumes", GVR: schema.GroupVersionResource{Version: "v1", Resource: "persistentvolumes"}, Namespaced: false, StatusCol: 2,
		Columns: []Column{{Name: "NAME"}, {Name: "CAPACITY"}, {Name: "STATUS"}, {Name: "CLAIM"}, {Name: "STORAGECLASS"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), pvCapacity(o), nestedString(o, "status", "phase"), pvClaim(o), nestedString(o, "spec", "storageClassName"), age(o)}
		}},
}

// FindResource returns the resource registered under key k, or nil.
func FindResource(k byte) *Resource {
	for i := range Resources {
		if Resources[i].Key == k {
			return &Resources[i]
		}
	}
	return nil
}

// List returns all instances of the resource in the given namespace.
// An empty namespace means all namespaces for namespaced resources, and is
// ignored for cluster-scoped resources.
func (c *Client) List(ctx context.Context, r Resource, ns string) ([]Row, error) {
	var resource dynamic.ResourceInterface
	dr := c.dyn.Resource(r.GVR)
	if r.Namespaced {
		resource = dr.Namespace(ns)
	} else {
		resource = dr.Namespace("")
	}
	obj, err := resource.List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	rows := make([]Row, 0, len(obj.Items))
	for _, item := range obj.Items {
		item := item
		rows = append(rows, Row{Cols: r.Extract(&item), Obj: &item})
	}
	return rows, nil
}

// Get returns a single resource by name.
func (c *Client) Get(ctx context.Context, r Resource, ns, name string) (*unstructured.Unstructured, error) {
	var resource dynamic.ResourceInterface
	dr := c.dyn.Resource(r.GVR)
	if r.Namespaced {
		resource = dr.Namespace(ns)
	} else {
		resource = dr.Namespace("")
	}
	return resource.Get(ctx, name, metav1.GetOptions{})
}

// --- extraction helpers ---

func name(o *unstructured.Unstructured) string {
	return nestedString(o, "metadata", "name")
}

func nestedString(o *unstructured.Unstructured, path ...string) string {
	s, ok, err := unstructured.NestedString(o.Object, path...)
	if err != nil || !ok {
		return "<none>"
	}
	return s
}

func nestedBool(o *unstructured.Unstructured, path ...string) bool {
	b, ok, err := unstructured.NestedBool(o.Object, path...)
	if err != nil || !ok {
		return false
	}
	return b
}

func nestedInt(o *unstructured.Unstructured, path ...string) int64 {
	i, ok, err := unstructured.NestedInt64(o.Object, path...)
	if err != nil || !ok {
		return 0
	}
	return i
}

func created(o *unstructured.Unstructured) time.Time {
	ts, ok, err := unstructured.NestedString(o.Object, "metadata", "creationTimestamp")
	if err != nil || !ok {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339, ts)
	return t
}

func age(o *unstructured.Unstructured) string {
	t := created(o)
	if t.IsZero() {
		return "<unknown>"
	}
	return AgeString(time.Since(t))
}

// AgeString renders a duration in a compact human form (e.g. "5m", "3d").
func AgeString(d time.Duration) string {
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

func podReady(o *unstructured.Unstructured) (ready, total int) {
	statuses, ok, _ := unstructured.NestedSlice(o.Object, "status", "containerStatuses")
	if !ok {
		return 0, 0
	}
	for _, s := range statuses {
		m, ok := s.(map[string]any)
		if !ok {
			continue
		}
		total++
		if r, ok := m["ready"].(bool); ok && r {
			ready++
		}
	}
	return ready, total
}

func podStatus(o *unstructured.Unstructured) string {
	if phase := nestedString(o, "status", "phase"); phase != "" && phase != "<none>" {
		return phase
	}
	statuses, ok, _ := unstructured.NestedSlice(o.Object, "status", "containerStatuses")
	if !ok {
		return "<none>"
	}
	for _, s := range statuses {
		m, ok := s.(map[string]any)
		if !ok {
			continue
		}
		if reason, ok := m["state"].(map[string]any)["waiting"].(map[string]any)["reason"].(string); ok && reason != "" {
			return reason
		}
	}
	return "<none>"
}

func podRestarts(o *unstructured.Unstructured) int {
	statuses, ok, _ := unstructured.NestedSlice(o.Object, "status", "containerStatuses")
	if !ok {
		return 0
	}
	total := int64(0)
	for _, s := range statuses {
		m, ok := s.(map[string]any)
		if !ok {
			continue
		}
		total += nestedIntFromMap(m, "restartCount")
	}
	return int(total)
}

func nestedIntFromMap(m map[string]any, key string) int64 {
	switch v := m[key].(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	}
	return 0
}

func readyFraction(o *unstructured.Unstructured, replicasField string) string {
	desired := nestedInt(o, "spec", "replicas")
	ready := nestedInt(o, "status", "readyReplicas")
	return fmt.Sprintf("%d/%d", ready, desired)
}

func nodeStatus(o *unstructured.Unstructured) string {
	conds, ok, _ := unstructured.NestedSlice(o.Object, "status", "conditions")
	if !ok {
		return "<none>"
	}
	for _, s := range conds {
		m, ok := s.(map[string]any)
		if !ok {
			continue
		}
		if m["type"] == "Ready" {
			if st, ok := m["status"].(string); ok {
				if st == "True" {
					return "Ready"
				}
				return "NotReady"
			}
		}
	}
	return "<none>"
}

func nodeRoles(o *unstructured.Unstructured) string {
	var roles []string
	labels, ok, _ := unstructured.NestedStringMap(o.Object, "metadata", "labels")
	if ok {
		for k := range labels {
			if strings.HasPrefix(k, "node-role.kubernetes.io/") {
				roles = append(roles, strings.TrimPrefix(k, "node-role.kubernetes.io/"))
			}
		}
	}
	if len(roles) == 0 {
		return "<none>"
	}
	return strings.Join(roles, ",")
}

func externalIP(o *unstructured.Unstructured) string {
	ing, ok, _ := unstructured.NestedSlice(o.Object, "status", "loadBalancer", "ingress")
	if !ok {
		return "<none>"
	}
	for _, i := range ing {
		m, ok := i.(map[string]any)
		if !ok {
			continue
		}
		if ip, ok := m["ip"].(string); ok && ip != "" {
			return ip
		}
		if h, ok := m["hostname"].(string); ok && h != "" {
			return h
		}
	}
	return "<pending>"
}

func servicePorts(o *unstructured.Unstructured) string {
	ports, ok, _ := unstructured.NestedSlice(o.Object, "spec", "ports")
	if !ok || len(ports) == 0 {
		return "<none>"
	}
	var parts []string
	for _, p := range ports {
		m, ok := p.(map[string]any)
		if !ok {
			continue
		}
		port := intStr(m["port"])
		target := intStr(m["targetPort"])
		if target != "" && target != port {
			parts = append(parts, port+":"+target)
		} else {
			parts = append(parts, port)
		}
	}
	return strings.Join(parts, ",")
}

func intStr(v any) string {
	switch t := v.(type) {
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case string:
		return t
	case map[string]any:
		if s, ok := t["strVal"].(string); ok {
			return s
		}
		if i, ok := t["intVal"].(int64); ok {
			return strconv.FormatInt(i, 10)
		}
		if f, ok := t["intVal"].(float64); ok {
			return strconv.FormatInt(int64(f), 10)
		}
	}
	return ""
}

func pvcCapacity(o *unstructured.Unstructured) string {
	cap, ok, _ := unstructured.NestedStringMap(o.Object, "status", "capacity")
	if ok {
		if v, ok := cap["storage"]; ok {
			return v
		}
	}
	return ""
}

func pvCapacity(o *unstructured.Unstructured) string {
	cap, ok, _ := unstructured.NestedStringMap(o.Object, "spec", "capacity")
	if ok {
		if v, ok := cap["storage"]; ok {
			return v
		}
	}
	return ""
}

func pvClaim(o *unstructured.Unstructured) string {
	claim, ok, _ := unstructured.NestedStringMap(o.Object, "spec", "claimRef")
	if ok {
		ns := claim["namespace"]
		name := claim["name"]
		if ns != "" && name != "" {
			return ns + "/" + name
		}
	}
	return ""
}

func ingressHosts(o *unstructured.Unstructured) string {
	rules, ok, _ := unstructured.NestedSlice(o.Object, "spec", "rules")
	if !ok {
		return "<none>"
	}
	var hosts []string
	for _, r := range rules {
		m, ok := r.(map[string]any)
		if !ok {
			continue
		}
		if h, ok := m["host"].(string); ok && h != "" {
			hosts = append(hosts, h)
		}
	}
	if len(hosts) == 0 {
		return "<none>"
	}
	return strings.Join(hosts, ",")
}

func completions(o *unstructured.Unstructured) string {
	desired := int64(0)
	if spec, ok := o.Object["spec"].(map[string]any); ok {
		desired = intStrToInt64(spec["completions"])
	}
	succeeded := nestedInt(o, "status", "succeeded")
	return fmt.Sprintf("%d/%d", succeeded, desired)
}

func intStrToInt64(v any) int64 {
	switch t := v.(type) {
	case int64:
		return t
	case float64:
		return int64(t)
	case string:
		if i, err := strconv.ParseInt(t, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

func jobDuration(o *unstructured.Unstructured) string {
	start := created(o)
	if start.IsZero() {
		return ""
	}
	if comp, ok, err := unstructured.NestedString(o.Object, "status", "completionTime"); err == nil && ok {
		if ct, err := time.Parse(time.RFC3339, comp); err == nil {
			return AgeString(ct.Sub(start))
		}
	}
	return ""
}

func lastSchedule(o *unstructured.Unstructured) string {
	ts, ok, err := unstructured.NestedString(o.Object, "status", "lastScheduleTime")
	if err != nil || !ok {
		return "<none>"
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return "<none>"
	}
	return AgeString(time.Since(t))
}
