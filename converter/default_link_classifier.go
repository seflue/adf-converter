package converter

import (
	"adf-converter/adf_types"
)

// DefaultLinkClassifier is the default implementation of LinkClassifier
type DefaultLinkClassifier struct{}

// NewDefaultLinkClassifier creates a new DefaultLinkClassifier
func NewDefaultLinkClassifier() *DefaultLinkClassifier {
	return &DefaultLinkClassifier{}
}

// ClassifyLink analyzes an ADF link mark and returns its classification
func (dlc *DefaultLinkClassifier) ClassifyLink(mark adf_types.ADFMark) LinkClassification {
	metadata := ExtractLinkMetadata(mark)
	classification := CreateClassificationFromMetadata(metadata)
	return classification
}

// DetermineStrategy returns the appropriate conversion strategy for a classification
func (dlc *DefaultLinkClassifier) DetermineStrategy(classification LinkClassification) ConversionStrategy {
	mapping := CreateMappingFromClassification(classification)
	return mapping.DetermineStrategy()
}

// IsInternalHref determines if an href points to Atlassian internal resources
func (dlc *DefaultLinkClassifier) IsInternalHref(href string) bool {
	return IsInternalHref(href)
}

// HasSignificantMetadata checks if attrs contain meaningful metadata beyond href
func (dlc *DefaultLinkClassifier) HasSignificantMetadata(attrs map[string]interface{}) bool {
	if attrs == nil {
		return false
	}

	// Check if there are any attributes other than href
	for key := range attrs {
		if key != "href" {
			return true
		}
	}

	return false
}

// ClassifyLinkFromADFMark is a convenience method that combines classification steps
func (dlc *DefaultLinkClassifier) ClassifyLinkFromADFMark(mark adf_types.ADFMark) (LinkClassification, ConversionStrategy) {
	classification := dlc.ClassifyLink(mark)
	strategy := dlc.DetermineStrategy(classification)
	return classification, strategy
}

// ValidateClassification checks if a classification is consistent with its attributes
func (dlc *DefaultLinkClassifier) ValidateClassification(classification LinkClassification, attrs map[string]interface{}) bool {
	href, ok := attrs["href"].(string)
	if !ok {
		return false
	}

	// Check internal consistency
	expectedIsInternal := dlc.IsInternalHref(href)
	if classification.IsInternal != expectedIsInternal {
		return false
	}

	// Check metadata consistency
	expectedHasMetadata := dlc.HasSignificantMetadata(attrs)
	if classification.HasMetadata != expectedHasMetadata {
		return false
	}

	// Check type consistency
	expectedType := WebLink
	if expectedIsInternal {
		if expectedHasMetadata {
			expectedType = ComplexInternalLink
		} else {
			expectedType = SimpleInternalLink
		}
	}

	if classification.Type != expectedType {
		return false
	}

	return true
}
