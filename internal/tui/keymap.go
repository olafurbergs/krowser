package tui

import "charm.land/bubbles/v2/key"

// KeyMap defines the global key bindings for the application.
type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Open         key.Binding
	Back         key.Binding
	Quit         key.Binding
	Filter       key.Binding
	ClearFilter  key.Binding
	YAML         key.Binding
	Describe     key.Binding
	Logs         key.Binding
	Namespace    key.Binding
	AllNamespace key.Binding
	Context      key.Binding
	Kinds        key.Binding
	Refresh      key.Binding
	Delete       key.Binding
	Scale        key.Binding
	Restart      key.Binding
	Edit         key.Binding
	Forward      key.Binding
	Top          key.Binding
	Themes       key.Binding
	Help         key.Binding
	Confirm      key.Binding
	Cancel       key.Binding
}

// DefaultKeyMap returns the application-wide key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		Open: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open"),
		),
		Back: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		ClearFilter: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear filter"),
		),
		YAML: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "yaml"),
		),
		Describe: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "describe"),
		),
		Logs: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "logs"),
		),
		Namespace: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "namespace"),
		),
		AllNamespace: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("g", "all namespaces"),
		),
		Context: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "context"),
		),
		Kinds: key.NewBinding(
			key.WithKeys("k"),
			key.WithHelp("k", "kinds"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Delete: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete"),
		),
		Scale: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "scale"),
		),
		Restart: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "restart"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit"),
		),
		Forward: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "port-forward"),
		),
		Top: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "top"),
		),
		Themes: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "themes"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("n", "esc"),
			key.WithHelp("n", "cancel"),
		),
	}
}

// ShortHelp implements the help.KeyMap interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		k.Open, k.Filter, k.Logs, k.YAML, k.Describe,
		k.Namespace, k.Context, k.Kinds, k.Refresh, k.Help, k.Quit,
	}
}

// FullHelp implements the help.KeyMap interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Open, k.YAML, k.Describe, k.Logs, k.Filter, k.ClearFilter},
		{k.Namespace, k.AllNamespace, k.Context, k.Kinds, k.Refresh, k.Back},
		{k.Delete, k.Scale, k.Restart, k.Edit, k.Forward, k.Top},
		{k.Themes, k.Help, k.Quit},
	}
}
