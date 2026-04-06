package converter

import (
	"fmt"
	"sort"

	"adf-converter/adf_types"
)

// LinkMetadata encapsulates link URL and all associated metadata
type LinkMetadata struct {
	Href       string
	Attributes map[string]interface{}
}

// NewLinkMetadata creates a new LinkMetadata with the given href
func NewLinkMetadata(href string) LinkMetadata {
	return LinkMetadata{
		Href:       href,
		Attributes: make(map[string]interface{}),
	}
}

// NewLinkMetadataWithAttributes creates a new LinkMetadata with href and attributes
func NewLinkMetadataWithAttributes(href string, attributes map[string]interface{}) LinkMetadata {
	metadata := LinkMetadata{
		Href:       href,
		Attributes: make(map[string]interface{}),
	}

	// Copy all attributes
	for key, value := range attributes {
		metadata.Attributes[key] = value
	}

	// Ensure href is in attributes
	metadata.Attributes["href"] = href

	return metadata
}

// AddAttribute adds or updates an attribute
func (lm *LinkMetadata) AddAttribute(key string, value interface{}) {
	if lm.Attributes == nil {
		lm.Attributes = make(map[string]interface{})
	}
	lm.Attributes[key] = value

	// Keep href in sync
	if key == "href" {
		if hrefStr, ok := value.(string); ok {
			lm.Href = hrefStr
		}
	}
}

// GetAttribute retrieves an attribute value
func (lm *LinkMetadata) GetAttribute(key string) (interface{}, bool) {
	if lm.Attributes == nil {
		return nil, false
	}
	value, exists := lm.Attributes[key]
	return value, exists
}

// GetAttributeAsString retrieves an attribute as a string
func (lm *LinkMetadata) GetAttributeAsString(key string) (string, bool) {
	if value, exists := lm.GetAttribute(key); exists {
		if strValue, ok := value.(string); ok {
			return strValue, true
		}
		// Convert other types to string
		return fmt.Sprintf("%v", value), true
	}
	return "", false
}

// HasAttribute checks if an attribute exists
func (lm *LinkMetadata) HasAttribute(key string) bool {
	_, exists := lm.GetAttribute(key)
	return exists
}

// RemoveAttribute removes an attribute
func (lm *LinkMetadata) RemoveAttribute(key string) {
	if lm.Attributes != nil {
		delete(lm.Attributes, key)
	}
}

// AttributeKeys returns all attribute keys
func (lm *LinkMetadata) AttributeKeys() []string {
	if lm.Attributes == nil {
		return []string{}
	}

	keys := make([]string, 0, len(lm.Attributes))
	for key := range lm.Attributes {
		keys = append(keys, key)
	}
	return keys
}

// AttributeCount returns the number of attributes
func (lm *LinkMetadata) AttributeCount() int {
	if lm.Attributes == nil {
		return 0
	}
	return len(lm.Attributes)
}

// IsEmpty returns true if the metadata has no href or attributes
func (lm *LinkMetadata) IsEmpty() bool {
	return lm.Href == "" && len(lm.Attributes) == 0
}

// HasSignificantMetadata returns true if there are attributes beyond href
func (lm *LinkMetadata) HasSignificantMetadata() bool {
	if lm.Attributes == nil {
		return false
	}

	// Check if there are any attributes other than href
	for key := range lm.Attributes {
		if key != "href" {
			return true
		}
	}

	return false
}

// Copy creates a deep copy of the LinkMetadata
func (lm *LinkMetadata) Copy() LinkMetadata {
	copy := LinkMetadata{
		Href:       lm.Href,
		Attributes: make(map[string]interface{}),
	}

	if lm.Attributes != nil {
		for key, value := range lm.Attributes {
			copy.Attributes[key] = value
		}
	}

	return copy
}

// ExtractLinkMetadata retrieves all metadata from an ADF link mark
func ExtractLinkMetadata(mark adf_types.ADFMark) LinkMetadata {
	metadata := LinkMetadata{
		Attributes: make(map[string]interface{}),
	}

	if mark.Attrs != nil {
		if href, ok := mark.Attrs["href"].(string); ok {
			metadata.Href = href
		}

		// Copy all attributes including href
		for key, value := range mark.Attrs {
			metadata.Attributes[key] = value
		}
	}

	return metadata
}

// BuildADFMark constructs an ADF link mark from metadata
func BuildADFMark(metadata LinkMetadata) adf_types.ADFMark {
	attrs := make(map[string]interface{})

	// Copy all attributes
	if metadata.Attributes != nil {
		for key, value := range metadata.Attributes {
			attrs[key] = value
		}
	}

	// Ensure href is present
	if metadata.Href != "" {
		attrs["href"] = metadata.Href
	}

	return adf_types.ADFMark{
		Type:  adf_types.MarkTypeLink,
		Attrs: attrs,
	}
}

// IsInternalHref determines if an href points to Atlassian internal resources
func IsInternalHref(href string) bool {
	if href == "" {
		return false
	}

	// External protocols
	if len(href) > 7 && (href[:8] == "https://" || href[:7] == "http://") {
		return false
	}

	if len(href) > 6 && href[:7] == "mailto:" {
		return false
	}

	if len(href) > 5 && href[:6] == "ftp://" {
		return false
	}

	// Protocol-relative URLs (//example.com) are external
	if len(href) >= 2 && href[:2] == "//" {
		return false
	}

	// Relative paths and internal paths are considered internal
	if href[0] == '/' {
		return true
	}

	// Other schemes are external
	return false
}

// ClassifyLinkFromMetadata determines the link type based on metadata
func ClassifyLinkFromMetadata(metadata LinkMetadata) LinkType {
	isInternal := IsInternalHref(metadata.Href)

	if !isInternal {
		return WebLink
	}

	// Internal links
	if metadata.HasSignificantMetadata() {
		return ComplexInternalLink
	}

	return SimpleInternalLink
}

// CreateClassificationFromMetadata creates a LinkClassification from metadata
func CreateClassificationFromMetadata(metadata LinkMetadata) LinkClassification {
	linkType := ClassifyLinkFromMetadata(metadata)
	isInternal := IsInternalHref(metadata.Href)
	hasMetadata := metadata.HasSignificantMetadata()

	var metadataKeys []string
	if hasMetadata {
		for key := range metadata.Attributes {
			if key != "href" {
				metadataKeys = append(metadataKeys, key)
			}
		}
		// Sort keys for consistent ordering
		sort.Strings(metadataKeys)
	}

	return LinkClassification{
		Type:         linkType,
		HasMetadata:  hasMetadata,
		IsInternal:   isInternal,
		MetadataKeys: metadataKeys,
	}
}
