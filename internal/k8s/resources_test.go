package k8s

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestAgeString(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m"},
		{2 * time.Hour, "2h"},
		{3 * 24 * time.Hour, "3d"},
	}
	for _, c := range cases {
		if got := AgeString(c.in); got != c.want {
			t.Errorf("AgeString(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func podObject(t *testing.T, phase string, ready bool, restarts int64) *unstructured.Unstructured {
	t.Helper()
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]any{
			"name":              "nginx",
			"namespace":         "default",
			"creationTimestamp": "2026-01-01T00:00:00Z",
		},
		"spec": map[string]any{
			"containers": []any{map[string]any{"name": "nginx", "image": "nginx"}},
		},
		"status": map[string]any{
			"phase": phase,
			"containerStatuses": []any{map[string]any{
				"name":         "nginx",
				"ready":        ready,
				"restartCount": restarts,
			}},
		},
	}}
}

func TestPodExtract(t *testing.T) {
	r := FindResource('3')
	if r == nil {
		t.Fatal("FindResource('3') returned nil, want pods")
	}
	pod := podObject(t, "Running", true, 2)
	cols := r.Extract(pod)
	want := []string{"nginx", "1/1", "Running", "2", AgeString(time.Since(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))}
	if len(cols) != len(want) {
		t.Fatalf("got %d columns, want %d: %v", len(cols), len(want), cols)
	}
	for i := range want {
		if cols[i] != want[i] {
			t.Errorf("column %d = %q, want %q", i, cols[i], want[i])
		}
	}
}

func TestListPods(t *testing.T) {
	pod := podObject(t, "Running", true, 0)
	client := newFakeClient(t, pod)
	rows, err := client.List(t.Context(), Resources[2], "default")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}
	if rows[0].Cols[0] != "nginx" {
		t.Errorf("row name = %q, want nginx", rows[0].Cols[0])
	}
}
