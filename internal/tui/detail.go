package tui

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/olafurb/krowser/internal/k8s"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// detailScreen shows the YAML manifest or describe output of a resource.
type detailScreen struct {
	client *k8s.Client
	theme  Theme

	res  *k8s.Resource
	obj  *unstructured.Unstructured
	kind string // "yaml" or "describe"

	data    string
	loading bool
	err     error

	width  int
	height int
	vp     viewport.Model

	// secretKeys and secretVals hold a Secret's data items with their decoded
	// values; secretMode replaces the raw YAML with a selectable list.
	secretKeys   []string
	secretVals   []string
	secretCursor int
	secretMode   bool
}

func newDetailScreen(client *k8s.Client, theme Theme, res *k8s.Resource, obj *unstructured.Unstructured, kind string) *detailScreen {
	d := &detailScreen{
		client: client,
		theme:  theme,
		res:    res,
		obj:    obj,
		kind:   kind,
		vp: viewport.New(
			viewport.WithWidth(80),
			viewport.WithHeight(20),
		),
	}
	return d
}

func (d *detailScreen) loadCmd() tea.Cmd {
	client := d.client
	res := d.res
	ns := d.obj.GetNamespace()
	name := d.obj.GetName()
	kind := d.kind
	return func() tea.Msg {
		var (
			data string
			err  error
		)
		switch kind {
		case "describe":
			data, err = client.Describe(tuiCtx, *res, ns, name)
		default:
			data, err = client.YAML(tuiCtx, *res, ns, name)
		}
		if err != nil {
			return loadErrorMsg{err: err}
		}
		return loadedDetailMsg{kind: kind, data: data}
	}
}

func (d *detailScreen) Init() tea.Cmd {
	d.loading = true
	return d.loadCmd()
}

func (d *detailScreen) Update(msg tea.Msg) (screenView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "left":
			return d, cmdMsg(backMsg{})
		case "y":
			if d.kind != "yaml" {
				d.kind = "yaml"
				d.secretMode = false
				d.loading = true
				return d, d.loadCmd()
			}
		case "d":
			if d.kind != "describe" {
				d.kind = "describe"
				d.secretMode = false
				d.loading = true
				return d, d.loadCmd()
			}
		case "up":
			if d.secretMode {
				d.moveSecret(-1)
				return d, nil
			}
		case "down":
			if d.secretMode {
				d.moveSecret(1)
				return d, nil
			}
		}
	case tea.MouseWheelMsg:
		if d.secretMode {
			switch msg.Mouse().Button {
			case tea.MouseWheelUp:
				d.moveSecret(-1)
			case tea.MouseWheelDown:
				d.moveSecret(1)
			}
			return d, nil
		}
	case loadedDetailMsg:
		d.loading = false
		d.err = nil
		d.data = msg.data
		d.renderContent()
		return d, nil
	case loadErrorMsg:
		d.loading = false
		d.err = msg.err
		return d, nil
	}
	vp, cmd := d.vp.Update(msg)
	d.vp = vp
	return d, cmd
}

func (d *detailScreen) moveSecret(delta int) {
	d.secretCursor += delta
	if d.secretCursor < 0 {
		d.secretCursor = 0
	}
	if d.secretCursor >= len(d.secretKeys) {
		d.secretCursor = len(d.secretKeys) - 1
	}
}

func (d *detailScreen) renderContent() {
	d.parseSecretData()
	content := d.data
	if d.kind == "yaml" {
		content = d.highlightYAML(d.data)
	} else {
		content = d.highlightDescribe(d.data)
	}
	d.vp.SetContent(content)
	d.vp.GotoTop()
}

// parseSecretData enables the decoded data-item list when showing a Secret's
// YAML, using the base64 data embedded in the fetched object.
func (d *detailScreen) parseSecretData() {
	d.secretMode = false
	if d.res == nil || d.obj == nil || d.res.Plural != "secrets" || d.kind != "yaml" {
		return
	}
	m, ok := d.obj.Object["data"].(map[string]any)
	if !ok || len(m) == 0 {
		return
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	d.secretKeys = keys
	d.secretVals = make([]string, len(keys))
	for i, k := range keys {
		enc, _ := m[k].(string)
		d.secretVals[i] = decodeSecretValue(enc)
	}
	if d.secretCursor >= len(keys) {
		d.secretCursor = 0
	}
	d.secretMode = true
}

// decodeSecretValue base64-decodes a Secret data value, marking binary payloads.
func decodeSecretValue(enc string) string {
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "<invalid base64>"
	}
	if len(raw) == 0 {
		return "(empty)"
	}
	if !isText(raw) {
		return fmt.Sprintf("binary data · %d bytes", len(raw))
	}
	return string(raw)
}

func isText(b []byte) bool {
	if !utf8.Valid(b) {
		return false
	}
	for _, c := range b {
		if c < 0x20 && c != '\t' && c != '\n' && c != '\r' {
			return false
		}
	}
	return true
}

// highlightYAML applies light syntax coloring to YAML output.
func (d *detailScreen) highlightYAML(s string) string {
	keyStyle := lipgloss.NewStyle().Foreground(d.theme.Accent)
	valueStyle := lipgloss.NewStyle().Foreground(d.theme.Good)
	commentStyle := lipgloss.NewStyle().Foreground(d.theme.Muted)
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		trimmed := line
		if idx := strings.Index(trimmed, " #"); idx >= 0 {
			b.WriteString(commentStyle.Render(trimmed[idx:]))
			trimmed = trimmed[:idx]
		}
		if idx := strings.Index(trimmed, ": "); idx >= 0 {
			b.WriteString(keyStyle.Render(trimmed[:idx+1]))
			b.WriteString(valueStyle.Render(trimmed[idx+1:]))
		} else {
			b.WriteString(trimmed)
		}
		b.WriteString("\n")
	}
	return b.String()
}

// highlightDescribe colors the leading field names in describe output.
func (d *detailScreen) highlightDescribe(s string) string {
	keyStyle := lipgloss.NewStyle().Foreground(d.theme.Accent).Bold(true)
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		if idx := strings.Index(line, ": "); idx >= 0 {
			b.WriteString(keyStyle.Render(line[:idx+1]))
			b.WriteString(line[idx+1:])
		} else {
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (d *detailScreen) Resize(width, height int) {
	d.width = width
	d.height = height
	d.vp.SetWidth(width)
	d.vp.SetHeight(height)
}

func (d *detailScreen) View() string {
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(d.theme.Primary).
		Render(d.res.Title + " · " + d.obj.GetName())
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	if d.loading {
		b.WriteString(d.theme.MutedText("loading…"))
		return b.String()
	}
	if d.err != nil {
		b.WriteString(d.theme.BadStyle().Render(d.err.Error()))
		return b.String()
	}
	if d.secretMode {
		b.WriteString(d.secretDataView())
		return b.String()
	}
	b.WriteString(d.vp.View())
	return b.String()
}

// secretDataView renders the selectable Secret data list with the decoded
// value of the selected item below the separator.
func (d *detailScreen) secretDataView() string {
	width := d.width
	if width <= 0 {
		width = 80
	}
	height := d.height
	if height <= 0 {
		height = 20
	}

	var b strings.Builder
	b.WriteString(d.theme.MutedText(fmt.Sprintf("data (%d) · ↑/↓ select", len(d.secretKeys))))
	b.WriteString("\n")

	maxList := height - 6
	if maxList < 1 {
		maxList = 1
	}
	start := 0
	if d.secretCursor >= maxList {
		start = d.secretCursor - maxList + 1
	}
	end := min(len(d.secretKeys), start+maxList)

	for i := start; i < end; i++ {
		if i == d.secretCursor {
			b.WriteString(d.theme.AccentText("▸ "))
		} else {
			b.WriteString("  ")
		}
		b.WriteString(d.secretKeys[i])
		b.WriteString("\n")
	}

	b.WriteString(d.theme.MutedText(strings.Repeat("─", width)))
	b.WriteString("\n")

	if d.secretCursor >= 0 && d.secretCursor < len(d.secretVals) {
		b.WriteString(d.theme.AccentText(d.secretKeys[d.secretCursor]))
		b.WriteString("\n")
		panelH := height - 4 - (end - start)
		if panelH < 0 {
			panelH = 0
		}
		lines := strings.Split(lipgloss.NewStyle().Width(width).Render(d.secretVals[d.secretCursor]), "\n")
		if panelH > 0 && len(lines) > 0 {
			if len(lines) > panelH {
				lines = lines[:panelH]
			}
			b.WriteString(strings.Join(lines, "\n"))
			b.WriteString("\n")
		}
	}
	return b.String()
}
