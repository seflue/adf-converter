package linkclass

import (
	"github.com/seflue/adf-converter/adf"
)

// DefaultLinkClassifier is the default implementation of link classification.
type DefaultLinkClassifier struct{}

func NewDefaultLinkClassifier() *DefaultLinkClassifier {
	return &DefaultLinkClassifier{}
}

func (dlc *DefaultLinkClassifier) ClassifyLink(mark adf.Mark) LinkClassification {
	metadata := ExtractLinkMetadata(mark)
	return CreateClassificationFromMetadata(metadata)
}

func (dlc *DefaultLinkClassifier) DetermineStrategy(classification LinkClassification) adf.ConversionStrategy {
	mapping := CreateMappingFromClassification(classification)
	return mapping.DetermineStrategy()
}

func (dlc *DefaultLinkClassifier) IsInternalHref(href string) bool {
	return IsInternalHref(href)
}

func (dlc *DefaultLinkClassifier) HasSignificantMetadata(attrs map[string]any) bool {
	if attrs == nil {
		return false
	}
	for key := range attrs {
		if key != "href" {
			return true
		}
	}
	return false
}

func (dlc *DefaultLinkClassifier) ClassifyLinkFromADFMark(mark adf.Mark) (LinkClassification, adf.ConversionStrategy) {
	classification := dlc.ClassifyLink(mark)
	strategy := dlc.DetermineStrategy(classification)
	return classification, strategy
}

func (dlc *DefaultLinkClassifier) ValidateClassification(classification LinkClassification, attrs map[string]any) bool {
	href, ok := attrs["href"].(string)
	if !ok {
		return false
	}
	expectedIsInternal := dlc.IsInternalHref(href)
	if classification.IsInternal != expectedIsInternal {
		return false
	}
	expectedHasMetadata := dlc.HasSignificantMetadata(attrs)
	if classification.HasMetadata != expectedHasMetadata {
		return false
	}
	expectedType := WebLink
	if expectedIsInternal {
		if expectedHasMetadata {
			expectedType = ComplexInternalLink
		} else {
			expectedType = SimpleInternalLink
		}
	}
	return classification.Type == expectedType
}
