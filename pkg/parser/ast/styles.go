package ast

import (
	"fmt"
	"slices"
	"strings"
)

var knownSkinparamTargets = []string{
	"class",
	"interface",
	"component",
	"package",
	"usecase",
	"node",
	"actor",
	"abstract",
	"annotation",
	"circle",
	"dataclass",
	"diamond",
	"entity",
	"enum",
	"exception",
	"metaclass",
	"protocol",
	"record",
	"stereotype",
}

var knownSkinparamParams = []string{
	"backgroundColor",
	"borderColor",
	"color",
	"defaultTextColor",
	"footerBackgroundColor",
	"footerBorderColor",
	"footerFontColor",
	"footerTextColor",
	"headerBackgroundColor",
	"headerBorderColor",
	"headerFontColor",
	"headerTextColor",
	"shadowing",
	"titleBackgroundColor",
	"titleBorderColor",
	"titleFontColor",
	"titleTextColor",
}

func init() {
	slices.SortFunc(knownSkinparamTargets, func(a, b string) int {
		return len(b) - len(a)
	})
}

type Skinparam map[string]string

func (s Skinparam) Set(target, param, stereotype, value string) error {
	if slices.Contains(knownSkinparamTargets, target) {
		return fmt.Errorf("unknown skinparam target: %s", target)
	}
	if slices.Contains(knownSkinparamParams, param) {
		return fmt.Errorf("unknown skinparam param: %s", param)
	}
	s[fmt.Sprintf("%s.%s.%s", strings.ToLower(target), strings.ToLower(param), strings.ToLower(stereotype))] = value
	return nil
}

func (s Skinparam) Get(target, param, stereotype string) string {
	target = strings.ToLower(target)
	param = strings.ToLower(param)
	stereotype = strings.ToLower(stereotype)

	// 1. Most specific: target + param + stereotype
	if val, ok := s[fmt.Sprintf("%s.%s.%s", target, param, stereotype)]; ok {
		return val
	}
	// 2. Target default: target + param + no stereotype
	if val, ok := s[fmt.Sprintf("%s.%s.", target, param)]; ok {
		return val
	}
	// 3. Global default: no target + param + no stereotype
	if val, ok := s[fmt.Sprintf(".%s.", param)]; ok {
		return val
	}
	// Fallback to exactly what was requested if it was a raw parameter
	return s[param]
}

func (s Skinparam) SetAndDecode(param, stereotype, value string) error {
	knownTarget := ""
	paramSlice := []rune(strings.ToLower(param))
	for _, t := range knownSkinparamTargets {
		if len(t) > len(paramSlice) {
			continue
		}
		if string(paramSlice[:len(t)]) != t {
			continue
		}
		knownTarget = t
		break
	}
	if knownTarget == "" {
		return s.Set("", param, stereotype, value)
	} else {
		return s.Set(knownTarget, param[len(knownTarget):], strings.ToLower(stereotype), value)
	}
}
