package parser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"yur4uwe/pac/internal/helpers"
	"yur4uwe/pac/pkg/parser/ast"
)

func ParseClassDiagram(puml string) (*ast.ClassDiagram, error) {
	classDiagram := &ast.ClassDiagram{
		Classes:       make(map[string]*ast.Class),
		Relationships: make(ast.RelationshipsRepo),
	}

	lines := strings.Split(puml, "\n")
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if after, ok := strings.CutPrefix(line, "class "); ok {
			className := strings.TrimSpace(after)
			idx := strings.IndexAny(className, " {")
			if idx == -1 {
				class := &ast.Class{Name: className}
				classDiagram.Classes[className] = class
				continue
			}

			class := &ast.Class{Name: strings.TrimSpace(className[:idx])}
			classDiagram.Classes[class.Name] = class

			nextI, err := RecurseClass(class, &lines, i)
			if err != nil {
				return nil, err
			}
			i = nextI

			continue
		}

		if strings.Contains(line, "--") || strings.Contains(line, "..") {
			rel, err := ParseRelationship(line, classDiagram.Classes)
			if err != nil {
				return nil, err
			}
			classDiagram.Relationships.AddRelationship(rel)
		}

		if strings.HasPrefix(trimmed, "@enduml") {
			break
		}
	}

	return classDiagram, nil
}

func GetFuncArguments(funcLine string, argStart, argEnd int) ([]ast.Attribute, error) {
	// MAGIC NUMBER ALERT: +1 to skip opening bracket
	argStr, err := helpers.SubstrRunes(funcLine, argStart+1, argEnd)
	if err != nil {
		return nil, err
	}
	argsSlice := strings.Split(argStr, ",")
	args := []ast.Attribute{}
	if len(argsSlice) == 1 && strings.TrimSpace(argsSlice[0]) == "" {
		return args, nil
	}

	for _, arg := range argsSlice {
		argParts := strings.SplitN(strings.TrimSpace(arg), ":", 2)
		if len(argParts) != 2 {
			return nil, errors.New("invalid method argument: " + arg + " " + strconv.Itoa(len(arg)))
		}
		argName := strings.TrimSpace(argParts[0])
		argType := ast.ParseTypeRef(strings.TrimSpace(argParts[1]))
		args = append(args, ast.Attribute{Name: argName, Type: argType})
	}

	return args, nil
}

func RecurseClass(class *ast.Class, lines *[]string, startIdx int) (continueAt int, err error) {
	for i := startIdx; i < len(*lines); i++ {
		line := strings.TrimSpace((*lines)[i])
		if line == "}" {
			return i, nil
		}

		openBracketIdx := strings.IndexRune(line, '(')
		closeBracketIdx := strings.IndexRune(line, ')')

		if openBracketIdx != -1 && closeBracketIdx != -1 && closeBracketIdx > openBracketIdx {
			// Method
			args, err := GetFuncArguments(line, openBracketIdx, closeBracketIdx)
			if err != nil {
				return -1, err
			}

			nameAndVis, err := helpers.SubstrRunes(line, 0, openBracketIdx)
			if err != nil {
				return -1, err
			}
			vis := ast.GetVisibility(nameAndVis)
			methodName := strings.TrimSpace(nameAndVis[1:])

			// MAGIC NUMBER ALERT: +2 to skip ":" after closing bracket it
			// includes space as it may not be there and trim spaces later
			returnTypeStr, err := helpers.SubstrRunes(line, closeBracketIdx+2, len(line))
			if err != nil {
				return -1, err
			}
			returnType := ast.ParseTypeRef(returnTypeStr)

			method := ast.Method{
				Name:       methodName,
				ReturnType: returnType,
				Visibility: vis,
				Parameters: args,
			}
			class.Methods = append(class.Methods, method)
		} else {
			// Attribute
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				attrName := strings.TrimSpace(parts[0])
				attrType := strings.TrimSpace(parts[1])
				visibility := ast.GetVisibility(attrName)

				attrName = attrName[1:]

				attribute := ast.Attribute{
					Name:       attrName,
					Type:       ast.ParseTypeRef(attrType),
					Visibility: visibility,
				}
				class.Attributes = append(class.Attributes, attribute)
			}
		}
	}

	return -1, errors.New("missing closing bracket for class")
}

func ParseRelationship(pumlLine string, classes map[string]*ast.Class) (*ast.Relationship, error) {
	rel := &ast.Relationship{}

	parts := strings.SplitN(pumlLine, ":", 2)
	if len(parts) == 2 {
		pumlLine = strings.TrimSpace(parts[0])
		rel.Comment = strings.TrimSpace(parts[1])
	}

	matches := regexp.MustCompile(`([A-Za-z_]+)\s(\"[0-9\.*]+\")?\s*([-\.<>|o*]{2,4})\s*(\"[0-9\.*]+\")?\s([A-Za-z_]+)`).FindAllStringSubmatch(pumlLine, -1)
	if len(matches) != 1 || len(matches[0]) < 4 {
		return nil, errors.New("invalid relationship line: " + pumlLine)
	}

	groups := matches[0]

	firstClassName := groups[1]
	secondClassName := groups[len(groups)-1]

	firstClass, ok := classes[firstClassName]
	if !ok {
		firstClass = &ast.Class{Name: firstClassName}
		classes[firstClassName] = firstClass
	}
	secondClass, ok := classes[secondClassName]
	if !ok {
		secondClass = &ast.Class{Name: secondClassName}
		classes[secondClassName] = secondClass
	}

	rel.From = firstClass
	rel.To = secondClass
	MultBeforeRelType := ast.UnknownMultiplicity
	MultAfterRelType := ast.UnknownMultiplicity
	var relType = ""
	var relDirection int
	for _, group := range groups[2 : len(groups)-1] {
		if group == "" {
			continue
		}

		if strings.HasPrefix(group, "\"") && strings.HasSuffix(group, "\"") {
			multiplicity, err := ast.ParseMultiplicity(group)
			if err != nil {
				return nil, err
			}
			if relType == "" {
				MultBeforeRelType = multiplicity
			} else {
				MultAfterRelType = multiplicity
			}
			continue
		}

		relType = strings.TrimSpace(group)
		rel.Type, relDirection = ast.ToRelationType(relType)
	}

	switch relDirection {
	case 1:
		// arrow is to the right
		rel.From = firstClass
		rel.To = secondClass
		rel.MultiplicityFrom = MultBeforeRelType
		rel.MultiplicityTo = MultAfterRelType
	case -1:
		// arrow is to the left
		rel.From = secondClass
		rel.To = firstClass
		rel.MultiplicityFrom = MultAfterRelType
		rel.MultiplicityTo = MultBeforeRelType
	case 0:
		// bidirectional or association with no direction
		rel.From = firstClass
		rel.To = secondClass
		rel.MultiplicityFrom = MultBeforeRelType
		rel.MultiplicityTo = MultAfterRelType
	}

	return rel, nil
}
