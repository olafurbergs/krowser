package k8s

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDiscoverKinds(t *testing.T) {
	resources := []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{Name: "pods", Kind: "Pod", Namespaced: true},
				{Name: "pods/log", Kind: "Pod"}, // subresource: must be skipped
				{Name: "namespaces", Kind: "Namespace", Namespaced: false},
			},
		},
		{
			GroupVersion: "stable.example.com/v1",
			APIResources: []metav1.APIResource{
				{Name: "crontabs", Kind: "CronTab", Namespaced: true},
			},
		},
	}

	kinds := kindsFromDiscovery(resources)
	if len(kinds) != 3 {
		t.Fatalf("got %d kinds, want 3 (subresource should be skipped)", len(kinds))
	}

	var pods, crontabs, namespaces *Kind
	for i := range kinds {
		switch kinds[i].Resource.Plural {
		case "pods":
			pods = &kinds[i]
		case "crontabs":
			crontabs = &kinds[i]
		case "namespaces":
			namespaces = &kinds[i]
		}
	}

	if pods == nil {
		t.Fatal("pods kind not found")
	}
	if len(pods.Resource.Columns) != 5 {
		t.Errorf("pods columns = %d, want 5 (registered favorite should be preferred)", len(pods.Resource.Columns))
	}

	if crontabs == nil {
		t.Fatal("crontabs kind not found")
	}
	if !crontabs.CRD || crontabs.Group != "stable.example.com" {
		t.Errorf("crontabs should be marked as CRD in group stable.example.com, got CRD=%v group=%q", crontabs.CRD, crontabs.Group)
	}
	if crontabs.Resource.Title != "CronTab" {
		t.Errorf("crontabs title = %q, want CronTab", crontabs.Resource.Title)
	}
	if !crontabs.Resource.Namespaced {
		t.Error("crontabs should be namespaced")
	}
	if len(crontabs.Resource.Columns) != 2 {
		t.Errorf("crontabs columns = %d, want 2 generic (NAME, AGE)", len(crontabs.Resource.Columns))
	}

	if namespaces == nil {
		t.Fatal("namespaces kind not found")
	}
	if namespaces.Resource.Namespaced {
		t.Error("namespaces should be cluster scoped")
	}
}
