package k8s

import (
	"context"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Kind describes a cluster-discovered resource type, including CRDs.
type Kind struct {
	Resource Resource
	Kind     string
	Group    string
	Version  string
	CRD      bool
}

// DiscoverKinds lists every listable API resource the cluster exposes,
// including custom resources provided by CRDs. A registered favorite is
// preferred when a discovered kind matches, preserving its richer columns
// and extractors.
func (c *Client) DiscoverKinds(ctx context.Context) ([]Kind, error) {
	lists, err := c.discovery.ServerPreferredResources()
	if err != nil && lists == nil {
		return nil, err
	}
	return kindsFromDiscovery(lists), nil
}

// kindsFromDiscovery builds the ordered Kind list from API resource lists,
// preferring registered favorites when a kind matches.
func kindsFromDiscovery(lists []*metav1.APIResourceList) []Kind {
	byGVR := make(map[schema.GroupVersionResource]Resource, len(Resources))
	for _, r := range Resources {
		byGVR[r.GVR] = r
	}

	seen := make(map[string]bool)
	var kinds []Kind
	for _, list := range lists {
		if list == nil {
			continue
		}
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			continue
		}
		for _, api := range list.APIResources {
			if strings.Contains(api.Name, "/") || api.Name == "" || api.Kind == "" {
				continue // subresources and unhelpful entries
			}
			gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: api.Name}
			if seen[gvr.String()] {
				continue
			}
			seen[gvr.String()] = true
			r, ok := byGVR[gvr]
			if !ok {
				r = genericResource(gvr, api)
			}
			kinds = append(kinds, Kind{
				Resource: r,
				Kind:     api.Kind,
				Group:    gv.Group,
				Version:  gv.Version,
				CRD:      gv.Group != "",
			})
		}
	}

	sort.Slice(kinds, func(i, j int) bool {
		return strings.ToLower(kinds[i].Resource.Title) < strings.ToLower(kinds[j].Resource.Title)
	})
	return kinds
}

// genericResource builds a minimal Resource for an arbitrary discovered kind.
func genericResource(gvr schema.GroupVersionResource, api metav1.APIResource) Resource {
	return Resource{
		Title:      kindTitle(api.Kind, gvr.Resource),
		Plural:     gvr.Resource,
		GVR:        gvr,
		Namespaced: api.Namespaced,
		StatusCol:  -1,
		Columns:    []Column{{Name: "NAME"}, {Name: "AGE"}},
		Extract: func(o *unstructured.Unstructured) []string {
			return []string{name(o), age(o)}
		},
	}
}

func kindTitle(kind, plural string) string {
	if kind != "" {
		return kind
	}
	if plural == "" {
		return plural
	}
	return strings.ToUpper(plural[:1]) + plural[1:]
}
