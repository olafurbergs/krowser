package loghl

import "fmt"

func findPatternMatches(
	matches *MatchLayer,
	syntaxList []Syntax,
	highlights map[string]*Highlight,
	text string,
	facts LineFacts,
) error {
	for _, syn := range syntaxList {
		p := syn.Pattern
		if !p.HasValue() {
			continue
		}
		if !p.CanRun(facts) {
			continue
		}
		hl, ok := highlights[syn.Group]
		if !ok {
			return fmt.Errorf("highlight group %s not found", syn.Group)
		}
		for _, idx := range p.FindAllStringIndex(text, -1) {
			*matches = append(*matches, Match{
				Start:     idx[0],
				End:       idx[1],
				AnsiStart: hl.BuildAnsi(),
				AnsiEnd:   hl.BuildAnsiReset(),
			})
		}
	}

	return nil
}
