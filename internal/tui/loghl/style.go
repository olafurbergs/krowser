package loghl

import (
	"fmt"
	"strings"
)

const ESCAPE = "\033["

const ResetAllAnsi = ESCAPE + "0m"

const ResetFgAnsi = ESCAPE + "39m"

const ResetBgAnsi = ESCAPE + "49m"

func Fg(r int, g int, b int) string {
	return fmt.Sprintf("%s38;2;%d;%d;%dm", ESCAPE, r, g, b)
}

func FgHex(hex string) string {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return Fg(r, g, b)
}

func Bg(r int, g int, b int) string {
	return fmt.Sprintf("%s48;2;%d;%d;%dm", ESCAPE, r, g, b)
}

func BgHex(hex string) string {
	var r, g, b int
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	return Bg(r, g, b)
}

const ItalicAnsi = "\033[3m"

const ResetItalicAnsi = "\033[23m"

const BoldAnsi = "\033[1m"

const ResetBoldAnsi = "\033[22m"

const UnderlineAnsi = "\033[4m"

const ResetUnderlineAnsi = "\033[24m"

// Highlight is an ANSI style applied to a syntax group.
type Highlight struct {
	Group     string
	Link      *string
	Fg        *string
	Bg        *string
	Italic    bool
	Bold      bool
	Underline bool
	ansi      *string
	ansiReset *string
}

func (h Highlight) BuildAnsi() string {
	if h.ansi != nil {
		return *h.ansi
	}
	var b strings.Builder
	if h.Fg != nil {
		b.WriteString(*h.Fg)
	}
	if h.Bg != nil {
		b.WriteString(*h.Bg)
	}
	if h.Bold {
		b.WriteString(BoldAnsi)
	}
	if h.Italic {
		b.WriteString(ItalicAnsi)
	}
	if h.Underline {
		b.WriteString(UnderlineAnsi)
	}
	h.ansi = Ptr(b.String())
	return *h.ansi
}

func (h Highlight) BuildAnsiReset() string {
	if h.ansiReset != nil {
		return *h.ansiReset
	}
	var b strings.Builder
	if h.Fg != nil {
		b.WriteString(ResetFgAnsi)
	}
	if h.Bg != nil {
		b.WriteString(ResetBgAnsi)
	}
	if h.Bold {
		b.WriteString(ResetBoldAnsi)
	}
	if h.Italic {
		b.WriteString(ResetItalicAnsi)
	}
	if h.Underline {
		b.WriteString(ResetUnderlineAnsi)
	}
	h.ansiReset = Ptr(b.String())
	return *h.ansiReset
}
