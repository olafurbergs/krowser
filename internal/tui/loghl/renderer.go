package loghl

import (
	"fmt"
	"strings"
)

type keywordMap map[string]*Highlight

// Renderer matches a log line against the configured syntax and produces an
// ANSI-colored rendering.
type Renderer struct {
	Config Config
	Theme  Theme

	builtinLowerKeywordMap keywordMap
	builtinKeywordMap      keywordMap
	userKeywordMap         keywordMap
}

func New(cfg Config, th Theme) (*Renderer, error) {
	renderer := &Renderer{
		Config: cfg,
		Theme:  th,
	}
	for _, hl := range cfg.Highlight {
		th.Insert(hl)
	}
	err := th.ResolveAllLinks()
	if err != nil {
		return nil, err
	}

	// precompile keyword map
	renderer.builtinLowerKeywordMap = make(keywordMap)
	for _, syntax := range renderer.Config.BuiltInSyntaxLower {
		for _, keyword := range syntax.Keywords {
			hl, ok := th.HighlightMap[syntax.Group]
			if !ok {
				return nil, fmt.Errorf("highlight group %q not found", syntax.Group)
			}
			renderer.builtinLowerKeywordMap[keyword] = hl
		}
	}

	renderer.builtinKeywordMap = make(keywordMap)
	for _, syntax := range renderer.Config.BuiltInSyntax {
		for _, keyword := range syntax.Keywords {
			hl, ok := th.HighlightMap[syntax.Group]
			if !ok {
				return nil, fmt.Errorf("highlight group %q not found", syntax.Group)
			}
			renderer.builtinKeywordMap[keyword] = hl
		}
	}

	renderer.userKeywordMap = make(keywordMap)
	for _, syntax := range renderer.Config.UserSyntax {
		for _, keyword := range syntax.Keywords {
			hl, ok := th.HighlightMap[syntax.Group]
			if !ok {
				return nil, fmt.Errorf("highlight group %q not found", syntax.Group)
			}
			renderer.userKeywordMap[keyword] = hl
		}
	}

	return renderer, nil
}

type Match struct {
	Start     int
	End       int
	AnsiStart string
	AnsiEnd   string
}

func (r Renderer) Render(text string) (string, error) {
	facts := MakeLineFacts(text)
	builtInLowerMatches, err := findMatches(
		r.Config.BuiltInSyntaxLower,
		r.Theme.HighlightMap,
		r.builtinLowerKeywordMap,
		text,
		facts,
	)
	if err != nil {
		return text, err
	}

	builtInMatches, err := findMatches(
		r.Config.BuiltInSyntax,
		r.Theme.HighlightMap,
		r.builtinKeywordMap,
		text,
		facts,
	)
	if err != nil {
		return text, err
	}

	builtinMatchesCombined := Stack(builtInMatches, builtInLowerMatches)

	userMatches, err := findMatches(
		r.Config.UserSyntax,
		r.Theme.HighlightMap,
		r.userKeywordMap,
		text,
		facts,
	)
	if err != nil {
		return text, err
	}

	matches := Stack(userMatches, builtinMatchesCombined)

	prefix := ""
	suffix := ""

	if userMatches.Len() > 0 {
		userBgHighlight, ok := r.Theme.HighlightMap["UserMatchLineBackground"]
		if !ok {
			return text, fmt.Errorf("highlight group %q not found", "UserMatchLineBackground")
		}

		// If we have user matches, the default background color for the line
		// should be the user match line background color.
		userBgAnsi := userBgHighlight.BuildAnsi()
		for i := range matches {
			if strings.Contains(matches[i].AnsiEnd, ResetBgAnsi) {
				matches[i].AnsiEnd = strings.ReplaceAll(matches[i].AnsiEnd, ResetBgAnsi, userBgAnsi)
			}
		}

		prefix = userBgAnsi
		suffix = userBgHighlight.BuildAnsiReset()
	}

	if len(matches) == 0 {
		return text, nil
	}

	matches.Sort()

	return prefix + buildHighlightedString(text, matches) + suffix, nil
}

func findMatches(
	syntaxList []Syntax,
	highlights map[string]*Highlight,
	keywordMap map[string]*Highlight,
	text string,
	facts LineFacts,
) (MatchLayer, error) {
	var matches MatchLayer
	var err error

	err = findPatternMatches(&matches, syntaxList, highlights, text, facts)
	if err != nil {
		return nil, err
	}

	err = findKeywordMatches(&matches, keywordMap, text)
	if err != nil {
		return nil, err
	}

	matches.removeOverlaps()
	matches.Sort()

	return matches, nil
}

func buildHighlightedString(text string, matches MatchLayer) string {
	var b strings.Builder
	b.Grow(len(text) * 2)

	b.WriteString(text[:matches[0].Start])
	for i, match := range matches {
		b.WriteString(match.AnsiStart)
		b.WriteString(text[match.Start:match.End])
		b.WriteString(match.AnsiEnd)
		if i == len(matches)-1 {
			b.WriteString(text[match.End:])
		} else {
			nextMatch := matches[i+1]
			b.WriteString(text[match.End:nextMatch.Start])
		}
	}

	return b.String()
}
