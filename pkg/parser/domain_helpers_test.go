package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
	"yur4uwe/pac/pkg/tokenizer"
)

func newParserForTest(input string) *Parser {
	return &Parser{
		stream: tokenizer.NewTokenStream(input),
	}
}

func TestTryReadModifier(t *testing.T) {
	p := newParserForTest("{abstract} {static} class")

	mod, err := p.tryReadModifier()
	require.NoError(t, err)
	require.Equal(t, "abstract", mod)

	mod, err = p.tryReadModifier()
	require.NoError(t, err)
	require.Equal(t, "static", mod)

	mod, err = p.tryReadModifier()
	require.Error(t, err)
	require.Equal(t, "", mod)
}

func TestTryReadStereotype(t *testing.T) {
	p := newParserForTest("<<stereotype>> <<foo bar>>")

	stereo, err := p.tryReadStereotype()
	require.NoError(t, err)
	require.Equal(t, "stereotype", stereo)

	stereo, err = p.tryReadStereotype()
	require.NoError(t, err)
	require.Equal(t, "foo bar", stereo)
}

func TestTryReadGeneric(t *testing.T) {
	p := newParserForTest("<T> <T, U>")

	gen, err := p.tryReadGeneric()
	require.NoError(t, err)
	require.Equal(t, "T", gen)

	gen, err = p.tryReadGeneric()
	require.NoError(t, err)
	require.Equal(t, "T, U", gen)
}

func TestTryReadClassSeparator(t *testing.T) {
	p := newParserForTest(".. separator ..\n== sep ==")

	sep, err := p.tryReadClassSeparator()
	require.NoError(t, err)
	require.Equal(t, "separator", sep.Label)
	require.Equal(t, '.', sep.Type)

	p.stream.TryConsumeType(tokenizer.NEWLINE)

	sep, err = p.tryReadClassSeparator()
	require.NoError(t, err)
	require.Equal(t, "sep", sep.Label)
	require.Equal(t, '=', sep.Type)
}

func TestReadDiagramBounds(t *testing.T) {
	tt := []struct {
		name             string
		input            string
		expectedKvps     map[string]string
		expectError      bool
		expectedFilename string
		expectedType     string
		expectedID       string
	}{
		{
			name:         "Default case",
			input:        "@startuml\n@enduml",
			expectedType: "uml",
		},
		{
			name:             "With filename",
			input:            "@startuml filename.puml\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
		},
		{
			name:         "With tag",
			input:        "@startuml(id=tag)\n@enduml",
			expectedType: "uml",
			expectedID:   "tag",
		},
		{
			name:             "With filename and tag",
			input:            "@startuml(id=tag) filename.puml\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
			expectedID:       "tag",
		},
		{
			name:             "filename using tool options",
			input:            "@startuml{filename.puml}\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
		},
		{
			name:             "filename and caption using tool options",
			input:            "@startuml{filename.puml, foo bar}\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
			expectedKvps: map[string]string{
				"caption": "foo bar",
			},
		},
		{
			name:             "kvp parsing",
			input:            "@startuml{filename.puml, foo bar, key=value}\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
			expectedKvps: map[string]string{
				"caption": "foo bar",
				"key":     "value",
			},
		},
		{
			name:             "parsing kvp's right after filename and no caption",
			input:            "@startuml{filename.puml, key=value}\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
			expectedKvps: map[string]string{
				"key": "value",
			},
		},
		{
			name:             "Tools and id parsing simulatenously",
			input:            "@startuml(id=tag){filename.puml, foo bar, key=value}\n@enduml",
			expectedFilename: "filename.puml",
			expectedType:     "uml",
			expectedID:       "tag",
			expectedKvps: map[string]string{
				"caption": "foo bar",
				"key":     "value",
			},
		},
		{
			name:        "Incorrect order of tools and id",
			input:       "@startuml{filename.puml, foo bar, key=value}(id=tag)\n@enduml",
			expectError: true,
		},
		{
			name:        "legacy filename syntax after tools",
			input:       "@startuml{filename.puml, foo bar, key=value} filename.puml\n@enduml",
			expectError: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			p := newParserForTest(tc.input)
			b, err := p.readDiagramBounds()
			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, b.IsStart)
			require.Equal(t, tc.expectedType, b.Type)

			if tc.expectedKvps != nil {
				require.Equal(t, tc.expectedKvps, b.Opts)
			}
			if tc.expectedFilename != "" {
				require.Equal(t, tc.expectedFilename, b.Name)
			}
			if tc.expectedID != "" {
				require.Equal(t, tc.expectedID, b.ID)
			}

			p.stream.TryConsumeType(tokenizer.NEWLINE)
			b, err = p.readDiagramBounds()
			require.NoError(t, err)
			require.False(t, b.IsStart)
			require.Equal(t, tc.expectedType, b.Type)
		})
	}
}
