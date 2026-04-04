package converter

import (
	"testing"

	"adf-converter/adf_types"
)

func TestDefaultClassifier_MentionIsEditable(t *testing.T) {
	c := NewDefaultClassifier()

	if !c.IsEditable(adf_types.NodeTypeMention) {
		t.Error("mention should be editable")
	}
	if c.IsPreserved(adf_types.NodeTypeMention) {
		t.Error("mention should not be preserved")
	}
}
