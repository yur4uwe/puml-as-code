package parser

import (
	"strings"
)

type UnhandledKeyword struct {
	Keyword  string
	Closer   string
	Nestable bool
}

var unhandledKeywords = []UnhandledKeyword{
	{
		Keyword: "caption",
	},
	{
		Keyword: "sprite",
	},

	// Blocked directives
	{
		Keyword:  "!function",
		Closer:   "!endfunction",
		Nestable: false,
	},
	{
		Keyword:  "!procedure",
		Closer:   "!endprocedure",
		Nestable: false,
	},
	{
		Keyword:  "!definelong",
		Closer:   "!enddefinelong",
		Nestable: false,
	},
	{
		Keyword:  "!if",
		Closer:   "!endif",
		Nestable: true,
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
}

func FindUnhandledKeyword(kw string) *UnhandledKeyword {
	for i := range unhandledKeywords {
		if strings.EqualFold(unhandledKeywords[i].Keyword, kw) {
			return &unhandledKeywords[i]
		}
	}
	return nil
}
