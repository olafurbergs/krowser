package k8s

import (
	"os"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/fake"
)

func newFakeClient(t *testing.T, objs ...runtime.Object) *Client {
	t.Helper()
	var dyn dynamic.Interface = fake.NewSimpleDynamicClient(runtime.NewScheme(), objs...)
	return &Client{dyn: dyn, context: "test"}
}

func writeTempKubeconfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	content := `apiVersion: v1
kind: Config
current-context: test
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: test
contexts:
- context:
    cluster: test
    user: test
  name: test
users:
- name: test
  user: {}
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestNew(t *testing.T) {
	c, err := New(writeTempKubeconfig(t), "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.Context(); got != "test" {
		t.Errorf("Context() = %q, want test", got)
	}
	if got := c.Namespace(); got != "" {
		t.Errorf("Namespace() = %q, want empty when context has no namespace", got)
	}
}

func TestNamespaceFromContext(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")
	content := `apiVersion: v1
kind: Config
current-context: prod
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: prod
contexts:
- context:
    cluster: prod
    user: prod
    namespace: payments
  name: prod
users:
- name: prod
  user: {}
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := New(path, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := c.Namespace(); got != "payments" {
		t.Errorf("Namespace() = %q, want payments", got)
	}
}

func TestContexts(t *testing.T) {
	infos, err := Contexts(writeTempKubeconfig(t))
	if err != nil {
		t.Fatalf("Contexts: %v", err)
	}
	if len(infos) != 1 {
		t.Fatalf("got %d contexts, want 1", len(infos))
	}
	if !infos[0].Current {
		t.Error("context should be marked current")
	}
}
