package tui

import "testing"

func TestStripMarker(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"● kind-kind", "kind-kind"},
		{" default", "default"},
		{"  Dracula", "Dracula"},
		{"plain", "plain"},
		{"", ""},
	}
	for _, c := range cases {
		if got := stripMarker(c.in); got != c.want {
			t.Errorf("stripMarker(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestPickerFilter(t *testing.T) {
	items := []pickerItem{
		{title: "Pods", search: "Pods pods"},
		{title: "Deployments", search: "Deployments deployments"},
		{title: "ResourceQuota", search: "ResourceQuota resourcequotas quota"},
	}
	p := newPicker(nil, DarkTheme, pickerResources, "Kinds", items)

	p.setQuery("quota")
	if n := len(p.list.VisibleItems()); n != 1 {
		t.Fatalf("expected 1 filtered item, got %d", n)
	}
	if it := p.list.VisibleItems()[0].(pickerItem); it.title != "ResourceQuota" {
		t.Errorf("filtered item = %q, want ResourceQuota", it.title)
	}
	if p.list.Title != "Kinds · quota" {
		t.Errorf("title = %q, want %q", p.list.Title, "Kinds · quota")
	}

	p.setQuery("")
	if n := len(p.list.VisibleItems()); n != 3 {
		t.Errorf("after clearing filter, items = %d, want 3", n)
	}
}
