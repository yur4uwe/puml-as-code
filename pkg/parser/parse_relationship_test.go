package parser

import (
	"testing"

	"yur4uwe/pac/internal/testutil"
	"yur4uwe/pac/pkg/parser/ast"
	"yur4uwe/pac/pkg/parser/dialect"
	"yur4uwe/pac/pkg/stream"
	"yur4uwe/pac/pkg/tokenizer"

	"github.com/stretchr/testify/require"
)

func TestParseArrowTokens(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        *ast.Relationship
		expectErr   bool
		errContains string
	}{
		{
			name:  "simple solid right arrow",
			input: "-->",
			want: &ast.Relationship{
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
			},
		},
		{
			name:  "simple solid right arrow short",
			input: "->",
			want: &ast.Relationship{
				Body:      '-',
				RArrow:    '>',
				BodyCount: 1,
				TypeRHS:   ast.RelationAssociation,
			},
		},
		{
			name:  "simple dotted right arrow",
			input: "..>",
			want: &ast.Relationship{
				Body:      '.',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationDependency,
			},
		},
		{
			name:  "double-headed arrow solid",
			input: "<-->",
			want: &ast.Relationship{
				LArrow:    '<',
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeLHS:   ast.RelationAssociation,
				TypeRHS:   ast.RelationAssociation,
			},
		},
		{
			name:  "double-headed arrow dotted",
			input: "<..>",
			want: &ast.Relationship{
				LArrow:    '<',
				Body:      '.',
				RArrow:    '>',
				BodyCount: 2,
				TypeLHS:   ast.RelationDependency,
				TypeRHS:   ast.RelationDependency,
			},
		},
		{
			name:  "left extension solid right arrow",
			input: "<|-->",
			want: &ast.Relationship{
				LArrow:    '|',
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeLHS:   ast.RelationInheritance,
				TypeRHS:   ast.RelationAssociation,
			},
		},
		{
			name:  "solid right extension arrow",
			input: "--|>",
			want: &ast.Relationship{
				Body:      '-',
				RArrow:    '|',
				BodyCount: 2,
				TypeRHS:   ast.RelationInheritance,
			},
		},
		{
			name:  "double extension solid",
			input: "<|--|>",
			want: &ast.Relationship{
				LArrow:    '|',
				Body:      '-',
				RArrow:    '|',
				BodyCount: 2,
				TypeLHS:   ast.RelationInheritance,
				TypeRHS:   ast.RelationInheritance,
			},
		},
		{
			name:  "curly braces left/right",
			input: "}--{",
			want: &ast.Relationship{
				LArrow:    '}',
				Body:      '-',
				RArrow:    '{',
				BodyCount: 2,
			},
		},
		{
			name:  "curly braces dotted left/right",
			input: "}..{",
			want: &ast.Relationship{
				LArrow:    '}',
				Body:      '.',
				RArrow:    '{',
				BodyCount: 2,
			},
		},
		{
			name:  "lolipop right arrowhead",
			input: "--()",
			want: &ast.Relationship{
				Body:      '-',
				RArrow:    '(',
				BodyCount: 2,
			},
		},
		{
			name:        "less-than pipe greater-than diamond/extension",
			input:       "<|>",
			expectErr:   true,
			errContains: "Unexpected token as the relationship body",
			want: &ast.Relationship{
				LArrow: '|',
			},
		},
		{
			name:  "custom 'x' left/right",
			input: "x--x",
			want: &ast.Relationship{
				LArrow:    'x',
				Body:      '-',
				RArrow:    'x',
				BodyCount: 2,
			},
		},
		{
			name:  "custom 'o' left/right",
			input: "o--o",
			want: &ast.Relationship{
				LArrow:    'o',
				Body:      '-',
				RArrow:    'o',
				BodyCount: 2,
				TypeLHS:   ast.RelationAggregation,
				TypeRHS:   ast.RelationAggregation,
			},
		},
		{
			name:  "asterisk left/right",
			input: "*--*",
			want: &ast.Relationship{
				LArrow:    '*',
				Body:      '-',
				RArrow:    '*',
				BodyCount: 2,
				TypeLHS:   ast.RelationComposition,
				TypeRHS:   ast.RelationComposition,
			},
		},
		{
			name:  "plus left/right",
			input: "+--+",
			want: &ast.Relationship{
				LArrow:    '+',
				Body:      '-',
				RArrow:    '+',
				BodyCount: 2,
			},
		},
		{
			name:  "caret left/right",
			input: "^--^",
			want: &ast.Relationship{
				LArrow:    '^',
				Body:      '-',
				RArrow:    '^',
				BodyCount: 2,
			},
		},
		{
			name:  "hash left/right",
			input: "#--#",
			want: &ast.Relationship{
				LArrow:    '#',
				Body:      '-',
				RArrow:    '#',
				BodyCount: 2,
			},
		},
		{
			name:  "left direction",
			input: "-left->",
			want: &ast.Relationship{
				Body:      '-',
				Direction: ast.DirectionLeft,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "right direction",
			input: "-right->",
			want: &ast.Relationship{
				Body:      '-',
				Direction: ast.DirectionRight,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "up direction",
			input: "-up->",
			want: &ast.Relationship{
				Body:      '-',
				Direction: ast.DirectionTop,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "down direction",
			input: "-down->",
			want: &ast.Relationship{
				Body:      '-',
				Direction: ast.DirectionBottom,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "single attribute line style",
			input: "-[bold]->",
			want: &ast.Relationship{
				Body:      '-',
				Style: &ast.RelationshipStyle{
					LineStyle: ast.LineBold,
				},
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "single attribute color",
			input: "-[#red]->",
			want: &ast.Relationship{
				Body:      '-',
				Style:     simpleRelStyle("red"),
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "single attribute thickness",
			input: "-[thickness=2]->",
			want: &ast.Relationship{
				Body:      '-',
				Style: &ast.RelationshipStyle{
					Thickness: 2,
				},
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "multiple attributes",
			input: "-[#red,dashed,thickness=2]->",
			want: &ast.Relationship{
				Body:      '-',
				Style: &ast.RelationshipStyle{
					LineColor: simpleColor("red"),
					LineStyle: ast.LineDashed,
					Thickness: 2,
				},
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "attributes and direction",
			input: "-[bold]left->",
			want: &ast.Relationship{
				Body:      '-',
				Style: &ast.RelationshipStyle{
					LineStyle: ast.LineBold,
				},
				Direction: ast.DirectionLeft,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "direction and attributes",
			input: "-left[bold]->",
			want: &ast.Relationship{
				Body:      '-',
				Style: &ast.RelationshipStyle{
					LineStyle: ast.LineBold,
				},
				Direction: ast.DirectionLeft,
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "plain attribute",
			input: "-[plain]->",
			want: &ast.Relationship{
				Body:      '-',
				Style: &ast.RelationshipStyle{
					LineStyle: ast.LineNormal,
				},
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "dotted attribute",
			input: "-[dotted]->",
			want: &ast.Relationship{
				Body:      '-',
				Style: &ast.RelationshipStyle{
					LineStyle: ast.LineDotted,
				},
				RArrow:    '>',
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "no arrowheads solid",
			input: "--",
			want: &ast.Relationship{
				Body:      '-',
				TypeLHS:   ast.RelationAssociation,
				TypeRHS:   ast.RelationAssociation,
				BodyCount: 2,
			},
		},
		{
			name:  "no arrowheads dotted",
			input: "..",
			want: &ast.Relationship{
				Body:      '.',
				BodyCount: 2,
			},
		},
		{
			name:  "preserve body count",
			input: "----",
			want: &ast.Relationship{
				Body:      '-',
				BodyCount: 4,
				TypeLHS:   ast.RelationAssociation,
				TypeRHS:   ast.RelationAssociation,
			},
		},
		{
			name:        "empty input error",
			input:       "",
			expectErr:   true,
			errContains: "Unexpected token at the start of relationship definition",
		},
		{
			name:        "invalid starting token error",
			input:       "@",
			expectErr:   true,
			errContains: "Unexpected token at the start of relationship definition",
		},
		{
			name:        "missing trailing body rune error",
			input:       "-up>",
			expectErr:   true,
			errContains: "Unexpected token in body relationship definition",
		},
		{
			name:        "invalid top direction value error",
			input:       "-top->",
			expectErr:   true,
			errContains: "Unexpected direction in relationship",
		},
		{
			name:        "invalid bottom direction value error",
			input:       "-bottom->",
			expectErr:   true,
			errContains: "Unexpected direction in relationship",
		},
		{
			name:        "invalid identifier error",
			input:       "-foo->",
			expectErr:   true,
			errContains: "Unexpected identifier in relationship definition",
		},
		{
			name:        "unexpected separator inside direction error",
			input:       "-left-[foo]->",
			expectErr:   true,
			errContains: "Cannot separate direction and attributes with a body token",
		},
		{
			name:        "unclosed pipe extension error",
			input:       "--|",
			expectErr:   true,
			errContains: "Expected '|>' after relationship",
		},
		{
			name:        "lolipop right after direction error",
			input:       "-left-()",
			expectErr:   true,
			errContains: "Lolipop interface cannot contain direction or attributes",
		},
		{
			name:        "lolipop right after attributes error",
			input:       "-[bold]-()",
			expectErr:   true,
			errContains: "Lolipop interface cannot contain",
		},
		{
			name:        "mixed body types error",
			input:       "-.-",
			expectErr:   true,
			errContains: "Different body type runes in relationship",
		},
		{
			name:        "unclosed attributes at EOF error",
			input:       "-[bold",
			expectErr:   true,
			errContains: "expected an attribute terminator",
			want: &ast.Relationship{
				Body:      '-',
				BodyCount: 1,
			},
		},
		{
			name:        "unclosed attributes at newline error",
			input:       "-[bold\n",
			expectErr:   true,
			errContains: "expected an attribute terminator",
			want: &ast.Relationship{
				Body:      '-',
				BodyCount: 1,
			},
		},
		{
			name:        "unclosed attributes empty at newline error",
			input:       "-[\n",
			expectErr:   true,
			errContains: "unexpected line terminator while parsing arrow attributes",
			want: &ast.Relationship{
				Body:      '-',
				BodyCount: 1,
			},
		},
		{
			name:        "missing attribute (double comma)",
			input:       "-[bold,,dashed]->",
			expectErr:   true,
			errContains: "unexpected comma in arrow attributes",
			want: &ast.Relationship{
				Body:      '-',
				BodyCount: 1,
			},
		},
		{
			name:        "unexpected identifier in arrow attributes",
			input:       "-[invalid]->",
			expectErr:   true,
			errContains: "recieved unexpected identifier in arrow attributes: invalid",
			want: &ast.Relationship{
				Body:      '-',
				BodyCount: 1,
			},
		},
		{
			name:        "expected equals after thickness",
			input:       "-[thickness]->",
			expectErr:   true,
			errContains: "expected EQAULS(=) after 'thickness'",
			want: &ast.Relationship{
				Body:      '-',
				BodyCount: 1,
			},
		},
		{
			name:        "expected number for thickness",
			input:       "-[thickness=abc]->",
			expectErr:   true,
			errContains: "expected a number to describe thickness",
			want: &ast.Relationship{
				Body:      '-',
				BodyCount: 1,
			},
		},
		{
			name:        "expected color token after hash",
			input:       "-[#]->",
			expectErr:   true,
			errContains: "expected color token in arrow attribute",
			want: &ast.Relationship{
				Body:      '-',
				BodyCount: 1,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream: stream.NewTokenStream(tc.input),
			}
			var got ast.Relationship
			err := p.parseArrowTokens(&got)
			if tc.expectErr {
				require.Error(t, err)
				if tc.errContains != "" {
					require.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				require.NoError(t, err)
			}

			if tc.want != nil {
				testutil.SemanticEqual(t, *tc.want, got)
			}
		})
	}
}

func TestParseRelationship_Valid(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ast.Relationship
	}{
		{
			name:  "composition left-headed without right arrowhead",
			input: "Car *-- Engine",
			want: ast.Relationship{
				LHS:       identRef("Car"),
				LArrow:    '*',
				Body:      '-',
				BodyCount: 2,
				TypeLHS:   ast.RelationComposition,
				RHS:       identRef("Engine"),
			},
		},
		{
			name:  "aggregation left-headed without right arrowhead",
			input: "Car o-- Wheel",
			want: ast.Relationship{
				LHS:       identRef("Car"),
				LArrow:    'o',
				Body:      '-',
				BodyCount: 2,
				TypeLHS:   ast.RelationAggregation,
				RHS:       identRef("Wheel"),
			},
		},
		{
			name:  "inheritance left-headed without right arrowhead",
			input: "Vehicle <|-- Car",
			want: ast.Relationship{
				LHS:       identRef("Vehicle"),
				LArrow:    '|',
				Body:      '-',
				BodyCount: 2,
				TypeLHS:   ast.RelationInheritance,
				RHS:       identRef("Car"),
			},
		},
		{
			name:  "composition with multiplicity on target",
			input: "Car *-- \"1..*\" Engine",
			want: ast.Relationship{
				LHS:       identRef("Car"),
				LArrow:    '*',
				Body:      '-',
				BodyCount: 2,
				TypeLHS:   ast.RelationComposition,
				MultRHS:   ast.Cardinality{Raw: "1..*", Min: 1, Max: -1},
				RHS:       identRef("Engine"),
			},
		},
		{
			name:  "association with label",
			input: "User --> Service : uses",
			want: ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("Service"),
				Label:     "uses",
			},
		},
		{
			name:  "realization with right arrowhead",
			input: "User ..|> Greeter",
			want: ast.Relationship{
				LHS:       identRef("User"),
				Body:      '.',
				RArrow:    '|',
				BodyCount: 2,
				TypeRHS:   ast.RelationRealization,
				RHS:       identRef("Greeter"),
			},
		},
		{
			name:  "relationship with trailing colon and empty label",
			input: "A --> B : ",
			want: ast.Relationship{
				LHS:       identRef("A"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("B"),
				Label:     "",
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 3, Offset: 2},
					End:   tokenizer.Pos{Col: 5, Offset: 4},
				},
			},
		},
		{
			name:  "association with simple named color",
			input: "A --> B #green",
			want: ast.Relationship{
				LHS:       identRef("A"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("B"),
				Style:     simpleRelStyle("green"),
			},
		},
		{
			name:  "association with simple hex color",
			input: "A --> B #FF0000",
			want: ast.Relationship{
				LHS:       identRef("A"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("B"),
				Style:     simpleRelStyle("FF0000"),
			},
		},
		{
			name:  "association with style and label",
			input: "A --> B #green : connects",
			want: ast.Relationship{
				LHS:       identRef("A"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("B"),
				Style:     simpleRelStyle("green"),
				Label:     "connects",
			},
		},
		{
			name:  "association with key-val line color and text color",
			input: "A --> B #line:blue;text:red",
			want: ast.Relationship{
				LHS:       identRef("A"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("B"),
				Style: &ast.RelationshipStyle{
					LineColor: simpleColor("blue"),
					TextColor: simpleColor("red"),
				},
			},
		},
		{
			name:  "association with line style and text color",
			input: "A --> B #line.dotted;text:blue",
			want: ast.Relationship{
				LHS:       identRef("A"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("B"),
				Style: &ast.RelationshipStyle{
					LineStyle: ast.LineDotted,
					TextColor: simpleColor("blue"),
				},
			},
		},
		{
			name:  "association with bare color, line style and text color",
			input: "A --> B #green;line.dashed;text:green",
			want: ast.Relationship{
				LHS:       identRef("A"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("B"),
				Style: &ast.RelationshipStyle{
					LineColor: simpleColor("green"),
					LineStyle: ast.LineDashed,
					TextColor: simpleColor("green"),
				},
			},
		},
		{
			name:  "association with line color, line style, text color and label",
			input: "A --> B #line:red;line.bold;text:red : red bold",
			want: ast.Relationship{
				LHS:       identRef("A"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("B"),
				Style: &ast.RelationshipStyle{
					LineColor: simpleColor("red"),
					LineStyle: ast.LineBold,
					TextColor: simpleColor("red"),
				},
				Label: "red bold",
			},
		},
		{
			name:  "association with arrow attribute color",
			input: "Car -[#red]-> Engine",
			want: ast.Relationship{
				LHS:       identRef("Car"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("Engine"),
				Style:     simpleRelStyle("red"),
			},
		},
		{
			name:  "association with arrow attribute style and thickness",
			input: "Car -[bold,thickness=2]-> Engine",
			want: ast.Relationship{
				LHS:       identRef("Car"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("Engine"),
				Style: &ast.RelationshipStyle{
					LineStyle: ast.LineBold,
					Thickness: 2,
				},
			},
		},
		{
			name:  "association with arrow attribute color, style, thickness and label",
			input: "Car -[#red,dashed,thickness=2]-> Engine : drives",
			want: ast.Relationship{
				LHS:       identRef("Car"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("Engine"),
				Style: &ast.RelationshipStyle{
					LineColor: simpleColor("red"),
					LineStyle: ast.LineDashed,
					Thickness: 2,
				},
				Label: "drives",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream: stream.NewTokenStream(tc.input),
			}
			firstTok := p.stream.Emit()

			rel := p.parseRelationship(firstTok)
			testutil.SemanticEqual(t, tc.want, rel)
		})
	}
}

func TestParseRelationship_Invalid(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      *ast.Relationship
		wantDiags []expectedDiag
	}{
		{
			name:  "trailing single dash without RHS (A -)",
			input: "A -",
			want: &ast.Relationship{
				LHS:       identRef("A"),
				Body:      '-',
				BodyCount: 1,
				TypeLHS:   ast.RelationAssociation,
				TypeRHS:   ast.RelationAssociation,
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 3, Offset: 2},
					End:   tokenizer.Pos{Col: 3, Offset: 2},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected identifier or string after relationship",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Col: 4, Offset: 3},
						End:   tokenizer.Pos{Col: 4, Offset: 3},
					},
				},
			},
		},
		{
			name:  "unclosed color attribute in arrow (A -[#red)",
			input: "A -[#red",
			want: &ast.Relationship{
				LHS:       identRef("A"),
				Body:      '-',
				BodyCount: 1,
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 3, Offset: 2},
					End:   tokenizer.Pos{Col: 3, Offset: 2},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "expected an attribute terminator",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Col: 9, Offset: 8},
						End:   tokenizer.Pos{Col: 9, Offset: 8},
					},
				},
			},
		},
		{
			name:  "unknown direction between dashes (A -unknown- B)",
			input: "A -unknown- B",
			want: &ast.Relationship{
				LHS:       identRef("A"),
				Body:      '-',
				BodyCount: 1,
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 3, Offset: 2},
					End:   tokenizer.Pos{Col: 3, Offset: 2},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Unexpected identifier in relationship definition, did you mean to write a direction?",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Col: 4, Offset: 3},
						End:   tokenizer.Pos{Col: 11, Offset: 10},
					},
				},
			},
		},
		{
			name:  "missing RHS after multiplicities (A \"1\" --> \"*\")",
			input: `A "1" --> "*"`,
			want: &ast.Relationship{
				LHS:       identRef("A"),
				MultLHS:   ast.Cardinality{Raw: "1", Min: 1, Max: 1},
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				MultRHS:   ast.Cardinality{Raw: "*", Min: 0, Max: -1},
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 7, Offset: 6},
					End:   tokenizer.Pos{Col: 9, Offset: 8},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected identifier or string after relationship",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Col: 14, Offset: 13},
						End:   tokenizer.Pos{Col: 14, Offset: 13},
					},
				},
			},
		},
		{
			name:  "no relationship RHS",
			input: "User -->",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 7, Offset: 7},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected identifier or string after relationship",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Col: 9, Offset: 8},
						End:   tokenizer.Pos{Col: 9, Offset: 8},
					},
				},
			},
		},
		{
			name:  "different body runes in arrow",
			input: "User --..> Order",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				BodyCount: 2,
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 7, Offset: 7},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Different body type runes in relationship",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Col: 8, Offset: 7},
						End:   tokenizer.Pos{Col: 9, Offset: 8},
					},
				},
			},
		},
		{
			name:  "misspelled direction in arrow",
			input: "User -leftt-> Order",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				BodyCount: 1,
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 6, Offset: 6},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Unexpected identifier in relationship definition, did you mean to write a direction?",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Col: 7, Offset: 6},
						End:   tokenizer.Pos{Col: 12, Offset: 11},
					},
				},
			},
		},
		{
			name:  "unclosed bracket in arrow attributes",
			input: "User -[bold",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				BodyCount: 1,
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 6, Offset: 6},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "expected an attribute terminator",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Col: 12, Offset: 11},
						End:   tokenizer.Pos{Col: 12, Offset: 11},
					},
				},
			},
		},
		{
			name:  "invalid multiplicity syntax emits info diagnostic and continues",
			input: `User "invalid_mult" --> "1" Order`,
			want: &ast.Relationship{
				LHS:       identRef("User"),
				MultLHS:   ast.UnknownCardinality,
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				MultRHS:   ast.Cardinality{Raw: "1", Min: 1, Max: 1},
				RHS:       identRef("Order"),
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 20, Offset: 20},
					End:   tokenizer.Pos{Col: 22, Offset: 22},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityInfo,
					Msg:      "incorrect multiplicity syntax",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Col: 6, Offset: 5},
						End:   tokenizer.Pos{Col: 20, Offset: 19},
					},
				},
			},
		},
		{
			name:  "trailing junk after relationship",
			input: "User --> Order extra_junk",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("Order"),
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 7, Offset: 7},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "Expected newline or colon after relationship",
					Span: tokenizer.SourceSpan{
						Start: tokenizer.Pos{Col: 16, Offset: 15},
						End:   tokenizer.Pos{Col: 26, Offset: 25},
					},
				},
			},
		},
		{
			name:  "invalid style token after hash",
			input: "User --> Order #:",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("Order"),
				Style:     &ast.RelationshipStyle{},
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 7, Offset: 7},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "unexpected token in style declaration",
				},
			},
		},
		{
			name:  "unexpected value for color target specifier in relationship style",
			input: "User --> Order #red;unknown:green",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("Order"),
				Style: &ast.RelationshipStyle{
					LineColor: simpleColor("red"),
				},
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 7, Offset: 7},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "unexpected value for color target specifier: unknown",
				},
			},
		},
		{
			name:  "missing colon after specified color target in relationship style",
			input: "User --> Order #text",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("Order"),
				Style:     &ast.RelationshipStyle{},
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 7, Offset: 7},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "missing colon after specified color target",
				},
			},
		},
		{
			name:  "unexpected token describing line style in relationship style",
			input: "User --> Order #line.",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				RArrow:    '>',
				BodyCount: 2,
				TypeRHS:   ast.RelationAssociation,
				RHS:       identRef("Order"),
				Style:     &ast.RelationshipStyle{},
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 7, Offset: 7},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "unexpected token describing line style",
				},
			},
		},
		{
			name:  "unexpected identifier in arrow attributes",
			input: "User -[invalid]-> Order",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				BodyCount: 1,
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 6, Offset: 6},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "recieved unexpected identifier in arrow attributes: invalid",
				},
			},
		},
		{
			name:  "missing equals after thickness in arrow attributes",
			input: "User -[thickness]-> Order",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				BodyCount: 1,
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 6, Offset: 6},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "expected EQAULS(=) after 'thickness'",
				},
			},
		},
		{
			name:  "unexpected comma in arrow attributes",
			input: "User -[bold,,dashed]-> Order",
			want: &ast.Relationship{
				LHS:       identRef("User"),
				Body:      '-',
				BodyCount: 1,
				ArrowSpan: tokenizer.SourceSpan{
					Start: tokenizer.Pos{Col: 5, Offset: 5},
					End:   tokenizer.Pos{Col: 6, Offset: 6},
				},
			},
			wantDiags: []expectedDiag{
				{
					Severity: ast.SeverityError,
					Msg:      "unexpected comma in arrow attributes",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &Parser{
				Dialect: dialect.NewGoDialect(),
				stream:  stream.NewTokenStream(tc.input),
				ast:     &ast.Diagram{},
			}
			firstTok := p.stream.Emit()
			rel := p.parseRelationship(firstTok)
			testutil.SemanticEqual(t, tc.want, rel)
			assertDiagsEqual(t, tc.wantDiags, p.ast.Diagnostics, rel)
		})
	}
}

func TestParseTargetRef(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  ast.TargetRef
	}{
		{
			name:  "simple entity",
			input: "Client",
			want:  identRef("Client"),
		},
		{
			name:  "package path and entity",
			input: "net.http.Client",
			want:  identRef("Client", "net", "http"),
		},
		{
			name:  "entity and simple member",
			input: "Client::Do",
			want:  identRefMember("Client", "Do"),
		},
		{
			name:  "entity and method parens",
			input: "Client::Do()",
			want: identRefMethod("Client", "Do", []tokenizer.Token{
				{Type: tokenizer.LPAREN, Literal: "("},
				{Type: tokenizer.RPAREN, Literal: ")"},
			}),
		},
		{
			name:  "entity and method with parameters",
			input: `Client::"Do(req Request)"`,
			want:  identRefStringMember("Client", "Do(req Request)"),
		},
		{
			name:  "package path entity and method in quotes",
			input: `net.http.Client::"Do(Context)"`,
			want:  identRefStringMember("Client", "Do(Context)", "net", "http"),
		},
		{
			name:  "package separator follows the entity but not a part of it",
			input: "Foo ..|> Bar",
			want:  identRef("Foo"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  stream.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			firstTok := p.stream.Emit()
			got, err := p.parseTargetRef(firstTok)
			require.NoError(t, err)
			testutil.SemanticEqual(t, tc.want, got)
		})
	}
}

func TestRelationshipStyle_Span(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantSpan string
	}{
		{
			name:     "simple line color",
			input:    "Foo --> Bar #red\n",
			wantSpan: "#red",
		},
		{
			name:     "simple hex line color",
			input:    "Foo --> Bar #FF0000\n",
			wantSpan: "#FF0000",
		},
		{
			name:     "key-val target styles",
			input:    "Foo --> Bar #line:green;text:yellow\n",
			wantSpan: "#line:green;text:yellow",
		},
		{
			name:     "line style dotted",
			input:    "Foo --> Bar #line.dotted;text:blue\n",
			wantSpan: "#line.dotted;text:blue",
		},
		{
			name:     "bare color and line style",
			input:    "Foo --> Bar #green;line.dashed;text:green\n",
			wantSpan: "#green;line.dashed;text:green",
		},
		{
			name:     "style with label",
			input:    "Foo --> Bar #line:red;line.bold;text:red : red bold\n",
			wantSpan: "#line:red;line.bold;text:red",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := Parser{
				stream:  stream.NewTokenStream(tc.input),
				Dialect: dialect.NewGoDialect(),
				ast:     &ast.Diagram{},
			}
			tok := p.stream.Emit()
			got := p.parseRelationship(tok)
			require.NotNil(t, got)
			require.NotNil(t, got.Style)
			gotSpan := tc.input[got.Style.Span.Start.Offset:got.Style.Span.End.Offset]
			require.Equal(t, tc.wantSpan, gotSpan)
		})
	}
}
