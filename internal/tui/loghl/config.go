package loghl

import "strings"

var strPtr = Ptr[string]

// Config holds the syntax rules and highlight links used by the renderer.
type Config struct {
	BuiltInSyntaxLower []Syntax
	BuiltInSyntax      []Syntax
	UserSyntax         []Syntax
	Highlight          []Highlight
}

func cap(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 'a' + 'A'
	}
	return c
}

func low(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c - 'A' + 'a'
	}
	return c
}

func lowerCapSCREAM(word string) []string {
	n := len(word)
	upper := make([]byte, n)
	lower := make([]byte, n)
	capitalize := make([]byte, n)

	for i := range n {
		upper[i] = cap(word[i])
		lower[i] = low(word[i])
	}

	copy(capitalize, lower)
	if n > 0 {
		capitalize[0] = cap(word[0])
	}

	return []string{
		string(lower),
		string(capitalize),
		string(upper),
	}
}

var DefaultConfig = Config{
	BuiltInSyntaxLower: []Syntax{
		// symbols
		{
			Group: "LogSymbol",
			Pattern: MustCompileWithGate(
				`[!@#$%^&*;:?=]`,
				func(f LineFacts) bool { return f.HasSymbol },
			),
		},

		// string
		{
			Group: "LogString",
			Pattern: MustCompileWithGate(
				`'([^'\\]|\\.)*'`,
				func(f LineFacts) bool { return f.HasSingleQuote },
			),
		},
		{
			Group: "LogString",
			Pattern: MustCompileWithGate(
				`"([^"\\]|\\.)*"`,
				func(f LineFacts) bool { return f.HasDoubleQuote },
			),
		},
	},

	BuiltInSyntax: []Syntax{
		// raw \n, \t, \r
		{
			Group: "LogSymbol",
			Pattern: MustCompileWithGate(
				`\\[ntr]`,
				func(f LineFacts) bool { return f.HasBackslash },
			),
		},

		// separators
		{
			Group: "LogSeparatorLine",
			Pattern: MustCompileWithGate(
				`(-{3,}|={3,}|#{3,}|\*{3,}|<{3,}|>{3,})`,
				func(f LineFacts) bool { return f.MaxSeparatorRun >= 3 }),
		},

		// numbers
		{
			Group: "LogNumber",
			Pattern: MustCompileWithGate(
				`\b\d+\b`,
				func(f LineFacts) bool { return f.HasDigit },
			),
		},
		{
			Group: "LogNumberFloat",
			Pattern: MustCompileWithGate(
				`\b\d+\.\d+([eE][+-]?\d+)?\b`,
				func(f LineFacts) bool { return f.HasDigit && f.HasDot },
			),
		},
		{
			Group: "LogNumberBin",
			Pattern: MustCompileWithGate(
				`\b0[bB][01]+\b`,
				func(f LineFacts) bool { return f.HasDigit },
			),
		},
		{
			Group: "LogNumberOctal",
			Pattern: MustCompileWithGate(
				`\b0[oO]?[0-7]+\b`,
				func(f LineFacts) bool { return f.HasDigit },
			),
		},
		{
			Group: "LogNumberHex",
			Pattern: MustCompileWithGate(
				`\b0[xX][0-9a-fA-F]+\b`,
				func(f LineFacts) bool { return f.HasDigit },
			),
		},
		{
			Group: "LogNumberHex",
			Pattern: MustCompileWithGate(
				`\b[0-9a-fA-F]{4,}\b`,
				func(f LineFacts) bool { return f.MaxHexRun >= 4 },
			),
		},

		// constants
		{
			Group:    "LogBool",
			Keywords: JoinSlices(lowerCapSCREAM("true"), lowerCapSCREAM("false")),
		},
		{
			Group:    "LogNull",
			Keywords: lowerCapSCREAM("null"),
		},

		// date and time
		// MM-DD, DD-MM, MM/DD, DD/MM
		{
			Group: "LogDate",
			Pattern: MustCompileWithGate(
				`\b\d{2}[-/]\d{2}\b`,
				func(f LineFacts) bool { return f.MaxDigitRun >= 2 && (f.HasHyphen || f.HasSlash) },
			),
		},
		// YYYY-MM-DD, YYYY/MM/DD, DD-MM-YYYY, DD/MM/YYYY
		{
			Group: "LogDate",
			Pattern: MustCompileWithGate(
				`\b\d{4}-\d{2}-\d{2}\b`,
				func(f LineFacts) bool { return f.MaxDigitRun >= 4 && (f.HasHyphen || f.HasSlash) },
			),
		},
		{
			Group: "LogDate",
			Pattern: MustCompileWithGate(
				`\b\d{4}/\d{2}/\d{2}\b`,
				func(f LineFacts) bool { return f.MaxDigitRun >= 4 && (f.HasHyphen || f.HasSlash) },
			),
		},
		{
			Group: "LogDate",
			Pattern: MustCompileWithGate(
				`\b\d{2}-\d{2}-\d{4}\b`,
				func(f LineFacts) bool { return f.MaxDigitRun >= 4 && (f.HasHyphen || f.HasSlash) },
			),
		},
		{
			Group: "LogDate",
			Pattern: MustCompileWithGate(
				`\b\d{2}/\d{2}/\d{4}\b`,
				func(f LineFacts) bool { return f.MaxDigitRun >= 4 && (f.HasHyphen || f.HasSlash) },
			),
		},
		// RFC3339
		{
			Group: "LogDate",
			Pattern: MustCompileWithGate(
				`(?:(\d{4}-\d{2}-\d{2})T(\d{2}:\d{2}:\d{2}(?:\.\d+)?))(Z|[\+-]\d{2}:\d{2})?`,
				func(f LineFacts) bool { return f.MaxDigitRun >= 4 && strings.Contains(f.Text, "T") },
			),
		},
		// 'Dec 31', 'Dec 31, 2023', 'Dec 31 2023'
		{
			Group: "LogDate",
			Pattern: MustCompileWithGate(
				`\b(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d{1,2}(,? [0-9]{4})?\b`,
				func(f LineFacts) bool { return f.HasDigit && f.HasUpper },
			),
		},
		// '31-Dec-2023', '31 Dec 2023'
		{
			Group: "LogDate",
			Pattern: MustCompileWithGate(
				`\b\d{1,2}[- ](Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)[- ]\d{4}\b`,
				func(f LineFacts) bool { return f.MaxDigitRun >= 4 && f.HasUpper },
			),
		},
		// weekday string
		{
			Group:    "LogWeekdayStr",
			Keywords: []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"},
		},
		// 12:34:56, 12:34:56.700000
		{
			Group: "LogTime",
			Pattern: MustCompileWithGate(
				`\b\d{2}:\d{2}:\d{2}(,\d{1,6}|\.\d{1,6})?\b`,
				func(f LineFacts) bool { return f.MaxDigitRun >= 2 && f.HasColon },
			),
		},
		// AM / PM
		{
			Group:    "LogTimeAMPM",
			Keywords: []string{"AM", "am", "PM", "pm"},
		},

		// Duration e.g. 10d20h30m40s, 123.456s, 123ms, 456us, 789ns
		{
			Group: "LogDuration",
			Pattern: MustCompileWithGate(
				`\b((\d+d)?(\d+h)?(\d+m)?\d+(\.\d+)?[µmun]?s)\b`,
				func(f LineFacts) bool { return f.HasDigit },
			),
		},

		// Objects
		{
			Group: "LogUrl",
			Pattern: MustCompileWithGate(
				`\bhttps?://\S+`,
				func(f LineFacts) bool {
					return strings.Contains(f.Text, "http://") || strings.Contains(f.Text, "https://")
				},
			),
		},
		{
			Group: "LogMacAddr",
			Pattern: MustCompileWithGate(
				`\b[0-9a-fA-F]{2}([:-][0-9a-fA-F]{2}){5}\b`,
				func(f LineFacts) bool { return f.MaxHexRun >= 2 && (f.HasColon || f.HasHyphen) },
			),
		},
		{
			Group: "LogIPv4",
			Pattern: MustCompileWithGate(
				`\b\d{1,3}(\.\d{1,3}){3}(\/\d+)?\b`,
				func(f LineFacts) bool { return f.HasDigit && f.DotCount >= 3 },
			),
		},
		{
			Group: "LogIPv6",
			Pattern: MustCompileWithGate(
				`\b[0-9a-fA-F]{1,4}(:[0-9a-fA-F]{1,4}){7}(\/\d+)?\b`,
				func(f LineFacts) bool { return f.HasColon && f.ColonCount >= 7 },
			),
		},
		{
			Group: "LogUUID",
			Pattern: MustCompileWithGate(
				`\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`,
				func(f LineFacts) bool { return f.MaxHexRun >= 12 && f.HyphenCount >= 4 },
			),
		},
		{
			Group: "LogMD5",
			Pattern: MustCompileWithGate(
				`\b[0-9a-fA-F]{32}\b`,
				func(f LineFacts) bool { return f.MaxHexRun >= 32 },
			),
		},
		{
			Group: "LogSHA",
			Pattern: MustCompileWithGate(
				`\b([0-9a-fA-F]{40}|[0-9a-fA-F]{56}|[0-9a-fA-F]{64}|[0-9a-fA-F]{96}|[0-9a-fA-F]{128})\b`,
				func(f LineFacts) bool { return f.MaxHexRun >= 40 },
			),
		},

		// log levels
		{
			Group:    "LogLvFatal",
			Keywords: lowerCapSCREAM("fatal"),
		},
		{
			Group:    "LogLvEmergency",
			Keywords: JoinSlices(lowerCapSCREAM("emerg"), lowerCapSCREAM("emergency")),
		},
		{
			Group:    "LogLvAlert",
			Keywords: lowerCapSCREAM("alert"),
		},
		{
			Group:    "LogLvCritical",
			Keywords: JoinSlices(lowerCapSCREAM("crit"), lowerCapSCREAM("critical")),
		},
		{
			Group: "LogLvError",
			Keywords: JoinSlices(
				[]string{"E"},
				lowerCapSCREAM("err"),
				lowerCapSCREAM("error"),
				lowerCapSCREAM("errors"),
			),
		},
		{
			Group: "LogLvFail",
			Keywords: JoinSlices(
				[]string{"F"},
				lowerCapSCREAM("fail"),
				lowerCapSCREAM("failed"),
				lowerCapSCREAM("failure"),
			),
		},
		{
			Group:    "LogLvFault",
			Keywords: lowerCapSCREAM("fault"),
		},
		{
			Group:    "LogLvNack",
			Keywords: JoinSlices(lowerCapSCREAM("nack"), lowerCapSCREAM("nak")),
		},
		{
			Group:    "LogLvWarning",
			Keywords: JoinSlices([]string{"W"}, lowerCapSCREAM("warn"), lowerCapSCREAM("warning")),
		},
		{
			Group:    "LogLvBad",
			Keywords: lowerCapSCREAM("bad"),
		},
		{
			Group:    "LogLvNotice",
			Keywords: lowerCapSCREAM("notice"),
		},
		{
			Group:    "LogLvInfo",
			Keywords: JoinSlices([]string{"I"}, lowerCapSCREAM("info")),
		},
		{
			Group:    "LogLvDebug",
			Keywords: JoinSlices([]string{"D"}, lowerCapSCREAM("dbg"), lowerCapSCREAM("debug")),
		},
		{
			Group:    "LogLvTrace",
			Keywords: lowerCapSCREAM("trace"),
		},
		{
			Group:    "LogLvVerbose",
			Keywords: JoinSlices([]string{"V"}, lowerCapSCREAM("verbose")),
		},
		{
			Group:    "LogLvPass",
			Keywords: JoinSlices(lowerCapSCREAM("pass"), lowerCapSCREAM("passed")),
		},
		{
			Group:    "LogLvSuccess",
			Keywords: JoinSlices(lowerCapSCREAM("succeed"), lowerCapSCREAM("succeeded"), lowerCapSCREAM("success")),
		},

		// Composite log levels e.g. *_INFO
		{
			Group: "LogLvFatal",
			Pattern: MustCompileWithGate(
				`[A-Z_]+_FATAL\b`,
				func(f LineFacts) bool {
					return f.HasUpper && f.HasUnderscore && strings.Contains(f.Text, "FATAL")
				}),
		},
		{
			Group: "LogLvEmergency",
			Pattern: MustCompileWithGate(
				`[A-Z_]+_EMERG(ENCY)?\b`,
				func(f LineFacts) bool {
					return f.HasUpper && f.HasUnderscore && strings.Contains(f.Text, "EMERG")
				},
			),
		},
		{
			Group: "LogLvAlert",
			Pattern: MustCompileWithGate(
				`[A-Z_]+_ALERT\b`,
				func(f LineFacts) bool {
					return f.HasUpper && f.HasUnderscore && strings.Contains(f.Text, "ALERT")
				},
			),
		},
		{
			Group: "LogLvCritical",
			Pattern: MustCompileWithGate(
				`[A-Z_]+_CRIT(ICAL)?\b`,
				func(f LineFacts) bool {
					return f.HasUpper && f.HasUnderscore && strings.Contains(f.Text, "CRIT")
				},
			),
		},
		{
			Group: "LogLvError",
			Pattern: MustCompileWithGate(
				`[A-Z_]+_ERR(OR)?\b`,
				func(f LineFacts) bool {
					return f.HasUpper && f.HasUnderscore && strings.Contains(f.Text, "ERR")
				},
			),
		},
		{
			Group: "LogLvFail",
			Pattern: MustCompileWithGate(
				`[A-Z_]+_FAIL(URE)?\b`,
				func(f LineFacts) bool {
					return f.HasUpper && f.HasUnderscore && strings.Contains(f.Text, "FAIL")
				},
			),
		},
		{
			Group: "LogLvWarning",
			Pattern: MustCompileWithGate(
				`[A-Z_]+_WARN(ING)?\b`,
				func(f LineFacts) bool {
					return f.HasUpper && f.HasUnderscore && strings.Contains(f.Text, "WARN")
				},
			),
		},
		{
			Group: "LogLvNotice",
			Pattern: MustCompileWithGate(
				`[A-Z_]+_NOTICE\b`,
				func(f LineFacts) bool {
					return f.HasUpper && f.HasUnderscore && strings.Contains(f.Text, "NOTICE")
				},
			),
		},
		{
			Group: "LogLvInfo",
			Pattern: MustCompileWithGate(
				`[A-Z_]+_INFO\b`,
				func(f LineFacts) bool {
					return f.HasUpper && f.HasUnderscore && strings.Contains(f.Text, "INFO")
				},
			),
		},
		{
			Group: "LogLvDebug",
			Pattern: MustCompileWithGate(
				`[A-Z_]+_DEBUG\b`,
				func(f LineFacts) bool {
					return f.HasUpper && f.HasUnderscore && strings.Contains(f.Text, "DEBUG")
				},
			),
		},
		{
			Group: "LogLvTrace",
			Pattern: MustCompileWithGate(
				`[A-Z_]+_TRACE\b`,
				func(f LineFacts) bool {
					return f.HasUpper && f.HasUnderscore && strings.Contains(f.Text, "TRACE")
				},
			),
		},
	},

	Highlight: []Highlight{
		{Group: "LogNumber", Link: strPtr("Number")},
		{Group: "LogNumberFloat", Link: strPtr("Float")},
		{Group: "LogNumberBin", Link: strPtr("Number")},
		{Group: "LogNumberOctal", Link: strPtr("Number")},
		{Group: "LogNumberHex", Link: strPtr("Number")},
		{Group: "LogSymbol", Link: strPtr("Special")},
		{Group: "LogSeparatorLine", Link: strPtr("Comment")},
		{Group: "LogBool", Link: strPtr("Boolean")},
		{Group: "LogNull", Link: strPtr("Constant")},
		{Group: "LogString", Link: strPtr("String")},
		{Group: "LogDate", Link: strPtr("Type")},
		{Group: "LogWeekdayStr", Link: strPtr("Type")},
		{Group: "LogTime", Link: strPtr("Operator")},
		{Group: "LogTimeAMPM", Link: strPtr("Operator")},
		{Group: "LogTimeZone", Link: strPtr("Operator")},
		{Group: "LogDuration", Link: strPtr("Operator")},
		{Group: "LogSysColumns", Link: strPtr("Statement")},
		{Group: "LogSysProcess", Link: strPtr("Function")},
		{Group: "LogUrl", Link: strPtr("Underlined")},
		{Group: "LogMacAddr", Link: strPtr("Underlined")},
		{Group: "LogIPv4", Link: strPtr("Underlined")},
		{Group: "LogIPv6", Link: strPtr("Underlined")},
		{Group: "LogUUID", Link: strPtr("Label")},
		{Group: "LogMD5", Link: strPtr("Label")},
		{Group: "LogSHA", Link: strPtr("Label")},
		{Group: "LogPath", Link: strPtr("Function")},
		{Group: "LogLvFatal", Link: strPtr("ErrorMsg")},
		{Group: "LogLvEmergency", Link: strPtr("ErrorMsg")},
		{Group: "LogLvAlert", Link: strPtr("ErrorMsg")},
		{Group: "LogLvCritical", Link: strPtr("ErrorMsg")},
		{Group: "LogLvError", Link: strPtr("ErrorMsg")},
		{Group: "LogLvFail", Link: strPtr("ErrorMsg")},
		{Group: "LogLvFault", Link: strPtr("ErrorMsg")},
		{Group: "LogLvNack", Link: strPtr("ErrorMsg")},
		{Group: "LogLvWarning", Link: strPtr("WarningMsg")},
		{Group: "LogLvBad", Link: strPtr("WarningMsg")},
		{Group: "LogLvNotice", Link: strPtr("Exception")},
		{Group: "LogLvInfo", Link: strPtr("LogBlue")},
		{Group: "LogLvDebug", Link: strPtr("Debug")},
		{Group: "LogLvTrace", Link: strPtr("Special")},
		{Group: "LogLvVerbose", Link: strPtr("Special")},
		{Group: "LogLvPass", Link: strPtr("LogGreen")},
		{Group: "LogLvSuccess", Link: strPtr("LogGreen")},
	},
}

func GetDefaultConfig() Config {
	return DefaultConfig
}
