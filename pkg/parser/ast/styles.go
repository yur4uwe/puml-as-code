package ast

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type SkinparamValueType int

const (
	SkinparamString SkinparamValueType = iota
	SkinparamColor
	SkinparamInt
	SkinparamBool
	SkinparamEnum
)

func (vt SkinparamValueType) Validate(value string) error {
	switch vt {
	case SkinparamColor:
		if !isValidColor(value) {
			return fmt.Errorf("invalid color: %s", value)
		}
	case SkinparamInt:
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("invalid integer: %s", value)
		}
	case SkinparamBool:
		v := strings.ToLower(value)
		if v != "true" && v != "false" && v != "on" && v != "off" && v != "yes" && v != "no" {
			return fmt.Errorf("invalid boolean: %s", value)
		}
	}
	return nil
}

func isValidColor(color string) bool {
	color = strings.TrimSpace(strings.ToLower(color))
	if color == "" || color == "transparent" {
		return true
	}

	if strings.HasPrefix(color, "#") {
		hex := color[1:]

		if len(hex) != 3 && len(hex) != 6 && len(hex) != 8 {
			return false
		}
		for _, r := range hex {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
				return false
			}
		}
		return true
	}

	if strings.HasPrefix(color, "rgb") {
		start := strings.Index(color, "(")
		end := strings.LastIndex(color, ")")
		if start == -1 || end == -1 || end < start {
			return false
		}

		strs := strings.Split(color[start+1:end], ",")
		isRGBA := strings.HasPrefix(color, "rgba")

		if (isRGBA && len(strs) != 4) || (!isRGBA && len(strs) != 3) {
			return false
		}

		for i, s := range strs {
			val, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			if err != nil {
				return false
			}
			if isRGBA && i == 3 {
				if val < 0 || val > 1 {
					return false
				}
				continue
			}
			if val < 0 || val > 255 {
				return false
			}
		}
		return true
	}

	return isValidNamedColor(color)
}

func isValidNamedColor(color string) bool {
	// Standard PlantUML colors are quite extensive.
	// For now we accept basic ones or common patterns.
	basicColors := []string{"red", "green", "blue", "white", "black", "grey", "gray", "yellow", "cyan", "magenta", "orange", "pink", "brown"}
	return slices.Contains(basicColors, strings.ToLower(color))
}

type SkinparamKey struct {
	MainTarget string
	SubTarget  string
	Param      string
	Stereotype string
}

func (k SkinparamKey) String() string {
	return fmt.Sprintf("%s.%s.%s.%s", k.MainTarget, k.SubTarget, k.Param, k.Stereotype)
}

func (k SkinparamKey) Lower() SkinparamKey {
	return SkinparamKey{
		MainTarget: strings.ToLower(k.MainTarget),
		SubTarget:  strings.ToLower(k.SubTarget),
		Param:      strings.ToLower(k.Param),
		Stereotype: strings.ToLower(k.Stereotype),
	}
}

type Skinparam map[string]string

var (
	// Base Property Groups
	fontProps = []string{"fontname", "fontsize", "fontstyle", "fontcolor"}
	bgBorder  = []string{"backgroundcolor", "bordercolor"}

	// Composite Groups
	stdNode       = append(fontProps, bgBorder...)
	stdStereoNode = append(stdNode, "stereotype")
	boxNode       = append(stdNode, "borderthickness")
	stdBoxNode    = append(boxNode, "stereotype")

	paramRegistry = map[string]SkinparamValueType{
		"backgroundcolor":    SkinparamColor,
		"bordercolor":        SkinparamColor,
		"borderthickness":    SkinparamInt,
		"fontcolor":          SkinparamColor,
		"fontname":           SkinparamString,
		"fontsize":           SkinparamInt,
		"fontstyle":          SkinparamString,
		"color":              SkinparamColor,
		"roundcorner":        SkinparamInt,
		"shadowing":          SkinparamBool,
		"iconsize":           SkinparamInt,
		"padding":            SkinparamInt,
		"style":              SkinparamString,
		"textalignment":      SkinparamString,
		"sep":                SkinparamInt,
		"position":           SkinparamString,
		"size":               SkinparamInt,
		"radius":             SkinparamInt,
		"thickness":          SkinparamInt,
		"margin":             SkinparamInt,
		"wraptitlewidth":     SkinparamInt,
		"width":              SkinparamInt,
		"underline":          SkinparamBool,
		"barcolor":           SkinparamColor,
		"externalcolor":      SkinparamColor,
		"alignment":          SkinparamString,
		"titlealignment":     SkinparamString,
		"messagealignment":   SkinparamString,
		"monospacedfontname": SkinparamString,
	}

	subTargetRegistry = map[string][]string{
		"stereotype":       fontProps,
		"attribute":        append(fontProps, "iconsize"),
		"header":           {"backgroundcolor"},
		"footer":           fontProps,
		"title":            fontProps,
		"diamond":          append(fontProps, bgBorder...),
		"end":              {"color"},
		"start":            {"color"},
		"arrow":            {"thickness"},
		"lollipop":         {"color"},
		"iemandatory":      {"color"},
		"private":          append(bgBorder, "color"),
		"public":           append(bgBorder, "color"),
		"protected":        append(bgBorder, "color"),
		"package":          append(bgBorder, "color"),
		"box":              stdNode,
		"delay":            fontProps,
		"divider":          boxNode,
		"group":            boxNode,
		"groupheader":      fontProps,
		"lifeline":         append(bgBorder, "borderthickness"),
		"message":          {"allignment", "textalignment"},
		"newpageseparator": {"color"},
		"reference":        append(boxNode, "alignment"),
		"referenceheader":  {"backgroundcolor"},
		"participant":      {"borderthickness"},
	}

	mainTargetRegistry = map[string][]string{
		"activity":         append(stdNode, "barcolor", "borderthickness", "diamond", "end", "start"),
		"actor":            stdStereoNode,
		"agent":            stdBoxNode,
		"arrow":            append(fontProps, "color", "lollipop", "messagealignment", "thickness"),
		"artifact":         stdStereoNode,
		"biddable":         bgBorder,
		"boundary":         stdStereoNode,
		"box":              {"padding"},
		"caption":          fontProps,
		"card":             stdBoxNode,
		"circledcharacter": append(fontProps, "radius"),
		"class":            append(stdBoxNode, "attribute", "header"),
		"cloud":            stdStereoNode,
		"collections":      bgBorder,
		"component":        append(stdBoxNode, "style"),
		"control":          stdStereoNode,
		"database":         stdStereoNode,
		"default":          append(fontProps, "monospacedfontname", "textalignment"),
		"designeddomain":   append(fontProps, "borderthickness", "stereotype"),
		"diagram":          {"bordercolor", "borderthickness"},
		"domain":           stdBoxNode,
		"entity":           stdStereoNode,
		"file":             stdStereoNode,
		"folder":           stdStereoNode,
		"footer":           fontProps,
		"frame":            stdStereoNode,
		"header":           fontProps,
		"hyperlink":        {"color", "underline"},
		"icon":             {"iemandatory", "package", "private", "protected", "public"},
		"interface":        stdStereoNode,
		"legend":           boxNode,
		"lexical":          bgBorder,
		"machine":          stdBoxNode,
		"node":             append(stdStereoNode, "sep"),
		"note":             append(bgBorder, "borderthickness", "textalignment", "shadowing"),
		"object":           append(stdBoxNode, "attribute"),
		"package":          append(stdBoxNode, "style", "titlealignment"),
		"page":             append(fontProps, "bordercolor", "margin", "externalcolor"),
		"participant":      append(stdStereoNode, "padding"),
		"partition":        boxNode,
		"queue":            stdStereoNode,
		"rank":             {"sep"},
		"rectangle":        stdBoxNode,
		"requirement":      stdBoxNode,
		"sequence":         {"actor", "arrow", "box", "divider", "group", "groupheader", "lifeline", "participant"},
		"stack":            stdStereoNode,
		"state":            append(stdNode, "attribute", "end", "start"),
		"stereotype":       {"a", "c", "e", "i", "n", "position"},
		"storage":          stdStereoNode,
		"swimlane":         append(boxNode, "stereotype", "title", "width", "wraptitlewidth"),
		"tab":              {"size"},
		"title":            append(boxNode, "borderroundcorner"),
		"usecase":          stdBoxNode,
		"wrap":             {"width"},
	}

	globalRegistry = map[string]SkinparamValueType{
		"colorarrowseparationspace": SkinparamInt,
		"dpi":                       SkinparamInt,
		"genericdisplay":            SkinparamString,
		"guillement":                SkinparamBool,
		"handwritten":               SkinparamBool,
		"linetype":                  SkinparamString,
		"maxasciimessagelength":     SkinparamInt,
		"maxmessagesize":            SkinparamInt,
		"minclasswidth":             SkinparamInt,
		"monochrome":                SkinparamBool,
		"pathhovercolor":            SkinparamColor,
		"responsemessagebelowarrow": SkinparamBool,
		"sameclasswidth":            SkinparamBool,
		"svglinktarget":             SkinparamString,
	}
)

func (s Skinparam) Set(key SkinparamKey, value string) error {
	lk := key.Lower()
	if lk.Param == "" {
		return fmt.Errorf("param cannot be empty")
	}

	var allowed []string
	var ok bool
	if lk.MainTarget != "" {
		allowed, ok = mainTargetRegistry[lk.MainTarget]
		if !ok {
			return fmt.Errorf("unknown main target: %s", key.MainTarget)
		}
	}

	if lk.MainTarget != "" && lk.SubTarget != "" {
		if !slices.Contains(allowed, lk.SubTarget) {
			return fmt.Errorf("sub-target %s not allowed for main target %s", key.SubTarget, key.MainTarget)
		}

		allowed, ok = subTargetRegistry[lk.SubTarget]
		if !ok {
			panic(fmt.Sprintf("unknown sub-target: %s", lk.SubTarget))
		}
	}

	if len(allowed) != 0 && !slices.Contains(allowed, lk.Param) {
		return fmt.Errorf("param %s not allowed for target \"%s.%s\"", lk.Param, lk.MainTarget, lk.SubTarget)
	}

	valueType, ok := paramRegistry[lk.Param]
	if !ok {
		valueType, ok = globalRegistry[lk.Param]
		if !ok {
			return fmt.Errorf("unknown param: %s", key.Param)
		}
	}

	if err := valueType.Validate(value); err != nil {
		return err
	}

	s[lk.String()] = value
	return nil
}

func (s Skinparam) Get(key SkinparamKey) string {
	lk := key.Lower()

	// 1. Most specific: Main + Sub + Param + Stereo
	if val, ok := s[lk.String()]; ok {
		return val
	}
	// 2. Main + Sub + Param
	if val, ok := s[fmt.Sprintf("%s.%s.%s.", lk.MainTarget, lk.SubTarget, lk.Param)]; ok {
		return val
	}
	// 3. Main + Param + Stereo
	if val, ok := s[fmt.Sprintf("%s..%s.%s", lk.MainTarget, lk.Param, lk.Stereotype)]; ok {
		return val
	}
	// 4. Main + Param
	if val, ok := s[fmt.Sprintf("%s..%s.", lk.MainTarget, lk.Param)]; ok {
		return val
	}
	// 5. Sub + Param
	if val, ok := s[fmt.Sprintf(".%s.%s.", lk.SubTarget, lk.Param)]; ok {
		return val
	}
	// 6. Global Param
	if val, ok := s[fmt.Sprintf("..%s.", lk.Param)]; ok {
		return val
	}

	return ""
}

func (s Skinparam) SetAndDecode(combinedName, stereotype, value string) error {
	return s.SetAndDecodeWithContext(SkinparamKey{Stereotype: stereotype}, combinedName, value)
}

func (s Skinparam) SetAndDecodeWithContext(context SkinparamKey, combinedName, value string) error {
	name := strings.ToLower(combinedName)
	key := context

	// If context doesn't have a MainTarget, try to find it in combinedName
	if key.MainTarget == "" {
		for t := range mainTargetRegistry {
			if strings.HasPrefix(name, t) {
				if len(t) > len(key.MainTarget) {
					key.MainTarget = t
				}
			}
		}
		if key.MainTarget != "" {
			name = name[len(key.MainTarget):]
		}
	}

	// If context doesn't have a SubTarget, try to find it in name
	if key.SubTarget == "" {
		for st := range subTargetRegistry {
			if strings.HasPrefix(name, st) {
				key.SubTarget = st
				name = name[len(st):]
				break
			}
		}
	}

	key.Param = name
	if key.Param == "" && key.SubTarget == "" && key.MainTarget == "" {
		return fmt.Errorf("could not decode skinparam name: %s", combinedName)
	}

	return s.Set(key, value)
}
