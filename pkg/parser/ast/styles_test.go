package ast

import (
	"slices"
	"testing"
)

func TestMainTargetRegistryClass(t *testing.T) {
	classAllowed, ok := mainTargetRegistry["class"]
	if !ok {
		t.Fatal("class not found in mainTargetRegistry")
	}

	expected := []string{
		"fontname", "fontsize", "fontstyle", "fontcolor",
		"backgroundcolor", "bordercolor", "borderthickness",
		"stereotype", "attribute", "header",
	}

	if len(classAllowed) != len(expected) {
		t.Errorf("expected %d elements, got %d: %v", len(expected), len(classAllowed), classAllowed)
	}

	for _, e := range expected {
		if !slices.Contains(classAllowed, e) {
			t.Errorf("expected element %s not found in %v", e, classAllowed)
		}
	}

	if slices.Contains(classAllowed, "alignment") {
		t.Errorf("unexpected element 'alignment' found in class registry: %v", classAllowed)
	}
}

func TestSubTargetRegistryReference(t *testing.T) {
	referenceAllowed, ok := subTargetRegistry["reference"]
	if !ok {
		t.Fatal("reference not found in subTargetRegistry")
	}

	expected := []string{
		"fontname", "fontsize", "fontstyle", "fontcolor",
		"backgroundcolor", "bordercolor", "borderthickness",
		"alignment",
	}

	if len(referenceAllowed) != len(expected) {
		t.Errorf("expected %d elements, got %d: %v", len(expected), len(referenceAllowed), referenceAllowed)
	}

	for _, e := range expected {
		if !slices.Contains(referenceAllowed, e) {
			t.Errorf("expected element %s not found in %v", e, referenceAllowed)
		}
	}
}
