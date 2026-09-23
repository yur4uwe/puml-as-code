package parser

import (
	"slices"
	"strings"

	"yur4uwe/pac/pkg/tokenizer"
)

type UnhandledKeyword struct {
	Keyword  string
	Closer   string
	Nestable bool

	BlockModifiers []string
}

func (uh UnhandledKeyword) IsBlock(remainderTokens []string) bool {
	if uh.Closer == "" {
		return false
	}
	if uh.BlockModifiers == nil {
		return true
	}
	for _, tok := range remainderTokens {
		if !slices.Contains(uh.BlockModifiers, tok) {
			return false // real content -> inline form
		}
	}
	return true // empty remainder, or modifiers only -> block form
}

var unhandledKeywords = []UnhandledKeyword{
	{
		Keyword: "caption",
	},
	{
		Keyword: "sprite",
	},

	{
		Keyword:        "footer",
		Closer:         "endfooter",
		BlockModifiers: []string{"left", "right", "center", "top", "bottom"},
	},
	{
		Keyword:        "header",
		Closer:         "endheader",
		BlockModifiers: []string{"left", "right", "center", "top", "bottom"},
	},
	{
		Keyword:        "legend",
		Closer:         "endlegend",
		BlockModifiers: []string{"left", "right", "center", "top", "bottom"},
	},
	{
		Keyword:        "title",
		Closer:         "endtitle",
		BlockModifiers: []string{"left", "right", "center", "top", "bottom"},
	},

	// directives
	{
		Keyword:  "!if",
		Closer:   "!endif",
		Nestable: true,
	},
	{
		Keyword: "!define",
	},
	{
		Keyword:  "!definelong",
		Closer:   "!enddefinelong",
		Nestable: true,
	},
	{
		Keyword: "!global",
	},
	{
		// single-line functions and procedures
		// aren't really representable in this structure
		Keyword: "!procedure",
		Closer:  "!endprocedure",
	},
	{
		Keyword: "!function",
		Closer:  "!endfunction",
	},
	{
		Keyword:  "!while",
		Closer:   "!endwhile",
		Nestable: true,
	},
	{
		Keyword:  "!foreach",
		Closer:   "!endfor",
		Nestable: true,
	},
	{
		// unquoted is a special case
		// it is used for procedure and function
		// and can be terminated by any of them
		Keyword: "!unquoted",
	},
	{
		Keyword: "!startsub",
	},
	{
		Keyword: "!endsub",
	},
}

func FindUnhandledKeyword(kw string) *UnhandledKeyword {
	for i := range unhandledKeywords {
		if strings.EqualFold(unhandledKeywords[i].Keyword, kw) {
			return &unhandledKeywords[i]
		}
	}
	return nil
}

func isCloser(lineToks []tokenizer.Token, closer string) bool {
	if len(lineToks) == 0 || closer == "" {
		return false
	}
	if strings.EqualFold(lineToks[0].Literal, closer) {
		return true
	}
	if len(lineToks) >= 2 && strings.EqualFold(lineToks[0].Literal, "end") {
		combined := "end" + lineToks[1].Literal
		if strings.EqualFold(combined, closer) {
			return true
		}
	}
	var sb strings.Builder
	for _, t := range lineToks {
		sb.WriteString(t.Literal)
	}
	return strings.EqualFold(sb.String(), closer)
}
