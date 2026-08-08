package tui

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/olafurb/krowser/internal/k8s"
	"github.com/sahilm/fuzzy"
)

type pickerMode int

// Picker modes.
const (
	pickerContexts pickerMode = iota
	pickerNamespaces
	pickerResources
	pickerContainers
	pickerThemes
)

// pickerItem adapts a simple value into a list item.
type pickerItem struct {
	title  string
	desc   string
	search string
	res    *k8s.Resource
}

func (i pickerItem) Title() string       { return i.title }
func (i pickerItem) Description() string { return i.desc }
func (i pickerItem) FilterValue() string { return i.search }

// pickerScreen is a generic selection list for contexts, namespaces,
// resource types, and pod containers.
type pickerScreen struct {
	theme  Theme
	client *k8s.Client
	mode   pickerMode
	title  string
	list   list.Model
	items  []pickerItem
	query  string

	loading bool
	err     error
}

func newPicker(client *k8s.Client, theme Theme, mode pickerMode, title string, items []pickerItem) *pickerScreen {
	d := list.NewDefaultDelegate()
	d.ShowDescription = mode == pickerContexts
	d.SetHeight(1)
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(theme.Accent).
		Foreground(theme.Accent).
		Padding(0, 0, 0, 1)
	d.Styles.SelectedDesc = d.Styles.SelectedTitle.Foreground(theme.Muted)
	d.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(theme.Text).
		Padding(0, 0, 0, 2)
	d.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(theme.Muted).
		Padding(0, 0, 0, 2)

	p := &pickerScreen{
		theme:  theme,
		client: client,
		mode:   mode,
		title:  title,
		list:   list.New(nil, d, 40, 20),
		items:  items,
	}
	p.list.KeyMap.CursorUp = key.NewBinding(key.WithKeys("up"), key.WithHelp("↑", "up"))
	p.list.KeyMap.CursorDown = key.NewBinding(key.WithKeys("down"), key.WithHelp("↓", "down"))
	p.list.SetShowStatusBar(true)
	p.list.SetShowTitle(true)
	p.list.Title = title
	p.list.Styles.Title = lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true)
	itemList := make([]list.Item, len(items))
	for i, it := range items {
		itemList[i] = it
	}
	if len(itemList) > 0 {
		p.list.SetItems(itemList)
	}
	return p
}

func newContextPicker(theme Theme, contexts []k8s.ContextInfo) *pickerScreen {
	items := make([]pickerItem, 0, len(contexts))
	for _, c := range contexts {
		marker := " "
		if c.Current {
			marker = "●"
		}
		items = append(items, pickerItem{
			title:  marker + " " + c.Name,
			desc:   c.Cluster + " · " + c.User,
			search: c.Name,
		})
	}
	return newPicker(nil, theme, pickerContexts, "Contexts", items)
}

func newResourcePicker(client *k8s.Client, theme Theme) *pickerScreen {
	return newPicker(client, theme, pickerResources, "Kinds", nil)
}

func newContainerPicker(theme Theme, containers []string) *pickerScreen {
	items := make([]pickerItem, 0, len(containers))
	for _, c := range containers {
		items = append(items, pickerItem{title: c, desc: "container", search: c})
	}
	return newPicker(nil, theme, pickerContainers, "Containers", items)
}

func newNamespacePicker(client *k8s.Client, theme Theme, namespaces []string) *pickerScreen {
	items := make([]pickerItem, 0, len(namespaces))
	for _, n := range namespaces {
		items = append(items, pickerItem{title: n, desc: "namespace", search: n})
	}
	return newPicker(client, theme, pickerNamespaces, "Namespaces", items)
}

func newThemePicker(theme Theme) *pickerScreen {
	items := make([]pickerItem, 0, len(Themes))
	for _, t := range Themes {
		marker := " "
		if t.Name == theme.Name {
			marker = "●"
		}
		kind := "dark"
		if !t.Dark {
			kind = "light"
		}
		items = append(items, pickerItem{
			title:  marker + " " + t.Name,
			desc:   kind + " theme",
			search: t.Name + " " + kind,
		})
	}
	return newPicker(nil, theme, pickerThemes, "Themes", items)
}

func (p *pickerScreen) Init() tea.Cmd {
	if len(p.items) > 0 {
		return nil
	}
	switch p.mode {
	case pickerNamespaces:
		client := p.client
		return func() tea.Msg {
			names, err := client.Namespaces(tuiCtx)
			if err != nil {
				return loadErrorMsg{err: err}
			}
			return loadedNamespacesMsg{names: names}
		}
	case pickerResources:
		client := p.client
		return func() tea.Msg {
			kinds, err := client.DiscoverKinds(tuiCtx)
			if err != nil {
				return loadErrorMsg{err: err}
			}
			return loadedKindsMsg{kinds: kinds}
		}
	}
	return nil
}

func (p *pickerScreen) Update(msg tea.Msg) (screenView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "left":
			return p, cmdMsg(backMsg{})
		case "enter", "right":
			item, ok := p.list.SelectedItem().(pickerItem)
			if !ok {
				return p, nil
			}
			return p, p.emit(item)
		case "ctrl+c":
			return p, quitCmd()
		case "backspace":
			if p.query != "" {
				p.setQuery(string([]rune(p.query)[:len([]rune(p.query))-1]))
			}
			return p, nil
		default:
			if msg.Text != "" {
				p.setQuery(p.query + msg.Text)
				return p, nil
			}
		}
	case loadedNamespacesMsg:
		p.loading = false
		p.err = nil
		p.setQuery("")
		p.setItems(msg.names)
		return p, nil
	case loadedKindsMsg:
		p.loading = false
		p.err = nil
		p.setQuery("")
		p.setKinds(msg.kinds)
		return p, nil
	case loadErrorMsg:
		p.loading = false
		p.err = msg.err
		return p, nil
	}
	l, cmd := p.list.Update(msg)
	p.list = l
	return p, cmd
}

// setQuery filters the picker items by q and updates the list. The active
// query is shown in the list title so the user can see what was typed.
func (p *pickerScreen) setQuery(q string) {
	p.query = q
	query := strings.TrimSpace(q)

	var filtered []pickerItem
	if query == "" {
		filtered = p.items
	} else {
		sources := make([]string, len(p.items))
		for i, it := range p.items {
			sources[i] = it.FilterValue()
		}
		matches := fuzzy.Find(query, sources)
		filtered = make([]pickerItem, 0, len(matches))
		for _, m := range matches {
			filtered = append(filtered, p.items[m.Index])
		}
	}

	listItems := make([]list.Item, len(filtered))
	for i, it := range filtered {
		listItems[i] = it
	}
	p.list.SetItems(listItems)
	p.list.ResetSelected()

	p.list.Title = p.title
	if p.query != "" {
		p.list.Title += " · " + p.query
	}
}

func (p *pickerScreen) setItems(names []string) {
	items := make([]pickerItem, 0, len(names))
	for _, n := range names {
		items = append(items, pickerItem{title: n, desc: "namespace", search: n})
	}
	listItems := make([]list.Item, len(items))
	for i, it := range items {
		listItems[i] = it
	}
	p.items = items
	p.list.SetItems(listItems)
}

func (p *pickerScreen) setKinds(kinds []k8s.Kind) {
	items := make([]pickerItem, 0, len(kinds))
	listItems := make([]list.Item, 0, len(kinds))
	for i := range kinds {
		k := kinds[i]
		scoped := "namespaced"
		if !k.Resource.Namespaced {
			scoped = "cluster"
		}
		gv := k.Version
		if k.Group != "" {
			gv = k.Group + "/" + k.Version
		}
		it := pickerItem{
			title:  k.Resource.Title,
			desc:   k.Resource.Plural + " · " + gv + " · " + scoped,
			search: k.Resource.Title + " " + k.Resource.Plural + " " + k.Kind + " " + k.Group,
			res:    &k.Resource,
		}
		items = append(items, it)
		listItems = append(listItems, it)
	}
	p.items = items
	p.list.SetItems(listItems)
}

func (p *pickerScreen) emit(item pickerItem) tea.Cmd {
	switch p.mode {
	case pickerContexts:
		return cmdMsg(selectedContextMsg{name: stripMarker(item.title)})
	case pickerNamespaces:
		return cmdMsg(selectedNamespaceMsg{name: item.title})
	case pickerResources:
		if item.res != nil {
			return cmdMsg(selectedResourceMsg{res: item.res})
		}
	case pickerContainers:
		return cmdMsg(selectedContainerMsg{name: item.title})
	case pickerThemes:
		return cmdMsg(selectedThemeMsg{name: stripMarker(item.title)})
	}
	return nil
}

func (p *pickerScreen) Resize(width, height int) {
	p.list.SetSize(width, height)
}

func (p *pickerScreen) View() string {
	if p.err != nil {
		return p.theme.BadStyle().Render(p.err.Error())
	}
	if p.loading {
		return p.theme.MutedText("loading…")
	}
	return p.list.View()
}

func stripMarker(s string) string {
	return strings.TrimLeft(s, "● ")
}
