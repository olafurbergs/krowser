package loghl

import (
	"fmt"
	"maps"
)

// Theme maps syntax groups to highlight styles.
type Theme struct {
	Name         string
	HighlightMap map[string]*Highlight
	linked       bool
}

func fg(raw string) *string {
	return Ptr(FgHex(raw))
}

func bg(raw string) *string {
	return Ptr(BgHex(raw))
}

var DefaultTheme = Theme{
	Name:   "default",
	linked: false,
	HighlightMap: map[string]*Highlight{
		"Constant": {
			Group: "Constant",
			Fg:    fg("#FF966C"),
		},
		"Number": {
			Group: "Number",
			Link:  Ptr("Constant"),
		},
		"Float": {
			Group: "Float",
			Link:  Ptr("Number"),
		},
		"Special": {
			Group: "Special",
			Fg:    fg("#65BCFF"),
		},
		"Comment": {
			Group:  "Comment",
			Fg:     fg("#636DA6"),
			Italic: true,
		},
		"Boolean": {
			Group: "Boolean",
			Link:  Ptr("Constant"),
		},
		"String": {
			Group: "String",
			Fg:    fg("#C3E88D"),
		},
		"Type": {
			Group: "Type",
			Fg:    fg("#65BCFF"),
		},
		"Operator": {
			Group: "Operator",
			Fg:    fg("#89DDFF"),
		},
		"Statement": {
			Group: "Statement",
			Fg:    fg("#C099FF"),
		},
		"Function": {
			Group: "Function",
			Fg:    fg("#82AAFF"),
		},
		"Underlined": {
			Group:     "Underlined",
			Underline: true,
		},
		"Label": {
			Group: "Label",
			Link:  Ptr("Statement"),
		},
		"Structure": {
			Group: "Structure",
			Link:  Ptr("Type"),
		},
		"ErrorMsg": {
			Group: "ErrorMsg",
			Fg:    fg("#C53B53"),
		},
		"WarningMsg": {
			Group: "WarningMsg",
			Fg:    fg("#FFC777"),
		},
		"Exception": {
			Group: "Exception",
			Link:  Ptr("Statement"),
		},
		"Debug": {
			Group: "Debug",
			Fg:    fg("#FF966C"),
		},
		"LogGreen": {
			Group: "LogGreen",
			Fg:    fg("#C3E88D"),
		},
		"LogBlue": {
			Group: "LogBlue",
			Fg:    fg("#65BCFF"),
		},
		"UserPattern": {
			Group: "UserPattern",
			Fg:    fg("#222436"),
			Bg:    bg("#C099FF"),
			Bold:  true,
		},
		"UserMatchLineBackground": {
			Group: "UserMatchLineBackground",
			Bg:    bg("#403355"),
		},
	},
}

func GetDefaultTheme() Theme {
	return DefaultTheme
}

func (t *Theme) ResolveOneLink(name string) error {
	return t.resolveOneLink(name, make(map[string]bool))
}

func (t *Theme) resolveOneLink(name string, visited map[string]bool) error {
	if visited[name] {
		return fmt.Errorf("highlight link cycle detected for %q", name)
	}
	visited[name] = true

	hl, ok := t.HighlightMap[name]
	if !ok {
		return fmt.Errorf("highlight %q not found", name)
	}

	if hl.Link == nil {
		delete(visited, name)
		return nil
	}

	targetName := *hl.Link
	targetHl, ok := t.HighlightMap[targetName]
	if !ok {
		return fmt.Errorf("highlight link target %q not found", targetName)
	}

	if targetHl.Link != nil {
		if err := t.resolveOneLink(targetName, visited); err != nil {
			return err
		}
	}

	hl.Fg = targetHl.Fg
	hl.Bg = targetHl.Bg
	hl.Bold = targetHl.Bold
	hl.Italic = targetHl.Italic
	hl.Underline = targetHl.Underline
	hl.Link = nil

	delete(visited, name)
	return nil
}

func (t *Theme) ResolveAllLinks() error {
	if t.linked {
		return nil
	}

	groupNames := maps.Keys(t.HighlightMap)
	for name := range groupNames {
		err := t.ResolveOneLink(name)
		if err != nil {
			return err
		}
	}

	t.linked = true
	return nil
}

func (t *Theme) Insert(hl Highlight) {
	t.HighlightMap[hl.Group] = &hl
	if hl.Link != nil {
		t.linked = false
	}
}

func (t *Theme) GetHighlight(group string) (Highlight, bool) {
	hl, ok := t.HighlightMap[group]
	return *hl, ok
}
