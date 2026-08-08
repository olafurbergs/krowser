package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/olafurb/krowser/internal/k8s"
)

func TestFormatCPU(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0m"},
		{1, "1m"},
		{250, "250m"},
		{1250, "1.25"},
		{1255, "1.25"},
		{3000, "3"},
		{1000, "1"},
		{333, "333m"},
	}
	for _, c := range cases {
		if got := formatCPU(c.in); got != c.want {
			t.Errorf("formatCPU(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{512, "512"},
		{2048, "2Ki"},
		{1024 * 1024, "1Mi"},
		{1536 * 1024 * 1024, "1.5Gi"},
		{1540 * 1024, "1.5Mi"},
	}
	for _, c := range cases {
		if got := formatBytes(c.in); got != c.want {
			t.Errorf("formatBytes(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestGauge(t *testing.T) {
	s := &topScreen{theme: DarkTheme, res: &k8s.Resource{Plural: "pods"}}
	bar := s.gauge(0.5)
	if got := lipgloss.Width(bar); got != 12 {
		t.Errorf("gauge width = %d, want 12", got)
	}
	if !strings.Contains(bar, "█") || !strings.Contains(bar, "░") {
		t.Errorf("gauge should contain filled and empty blocks: %q", bar)
	}
	over := s.gauge(1.5)
	if !strings.Contains(over, "█") || strings.Contains(over, "░") {
		t.Errorf("over-limit gauge should be fully filled: %q", over)
	}
}

func TestTopCellPod(t *testing.T) {
	s := &topScreen{theme: DarkTheme, res: &k8s.Resource{Plural: "pods"}}
	e := k8s.TopEntry{Name: "p", CPU: 500, CPULim: 1000, CPUReq: 250, Mem: 1024 * 1024, MemLim: 2 * 1024 * 1024, MemReq: 512 * 1024}
	cpu := s.cell(e, true)
	if !strings.Contains(cpu, "500m/1 rq250m") {
		t.Errorf("cpu cell = %q, want it to contain 500m/1 rq250m", cpu)
	}
	mem := s.cell(e, false)
	if !strings.Contains(mem, "1Mi/2Mi rq512Ki") {
		t.Errorf("mem cell = %q, want it to contain 1Mi/2Mi rq512Ki", mem)
	}
}

func TestTopCellNode(t *testing.T) {
	s := &topScreen{theme: DarkTheme, res: &k8s.Resource{Plural: "nodes"}}
	e := k8s.TopEntry{Name: "n", CPU: 500, CPULim: 2000}
	cpu := s.cell(e, true)
	if !strings.Contains(cpu, "500m/2 25%") {
		t.Errorf("node cpu cell = %q, want it to contain 500m/2 25%%", cpu)
	}
}

func TestGaugePalette(t *testing.T) {
	s := newTopScreen(nil, DarkTheme, &k8s.Resource{Plural: "pods"}, "", false)
	if len(s.palette) != 12 {
		t.Fatalf("palette length = %d, want 12", len(s.palette))
	}
	if s.palette[0] == s.palette[11] {
		t.Error("palette endpoints are identical, expected a green-to-red gradient")
	}
}

func TestTopApplyTableColumns(t *testing.T) {
	s := newTopScreen(nil, DarkTheme, &k8s.Resource{Plural: "pods", Namespaced: true, Title: "Pods"}, "default", false)
	s.width = 120
	s.entries = []k8s.TopEntry{
		{Name: "a-very-long-pod-name-that-must-fit", CPU: 500, CPULim: 1000, CPUReq: 250, Mem: 1 << 20, MemLim: 2 << 20, MemReq: 512 << 10},
		{Name: "short", CPU: 10, Mem: 1024},
	}
	s.applyTable()

	cols := s.table.Columns()
	if len(cols) != 3 {
		t.Fatalf("expected 3 columns, got %d", len(cols))
	}
	longest := "a-very-long-pod-name-that-must-fit"
	if cols[0].Width < len(longest) {
		t.Errorf("NAME column width %d < longest name %d", cols[0].Width, len(longest))
	}
	total := 0
	for _, c := range cols {
		total += c.Width
	}
	if total < s.width-len(cols)+1 || total > s.width {
		t.Errorf("columns total width = %d, want to fill width %d", total, s.width)
	}
}
