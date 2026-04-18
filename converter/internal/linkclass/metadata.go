package linkclass

import (
	"fmt"
	"sort"

	"github.com/seflue/adf-converter/adf_types"
)

// LinkMetadata encapsulates link URL and all associated metadata
type LinkMetadata struct {
	Href       string
	Attributes map[string]any
}

func NewLinkMetadata(href string) LinkMetadata {
	return LinkMetadata{
		Href:       href,
		Attributes: make(map[string]any),
	}
}

func NewLinkMetadataWithAttributes(href string, attributes map[string]any) LinkMetadata {
	metadata := LinkMetadata{
		Href:       href,
		Attributes: make(map[string]any),
	}
	for key, value := range attributes {
		metadata.Attributes[key] = value
	}
	metadata.Attributes["href"] = href
	return metadata
}

func (lm *LinkMetadata) AddAttribute(key string, value any) {
	if lm.Attributes == nil {
		lm.Attributes = make(map[string]any)
	}
	lm.Attributes[key] = value
	if key == "href" {
		if hrefStr, ok := value.(string); ok {
			lm.Href = hrefStr
		}
	}
}

func (lm *LinkMetadata) GetAttribute(key string) (any, bool) {
	if lm.Attributes == nil {
		return nil, false
	}
	value, exists := lm.Attributes[key]
	return value, exists
}

func (lm *LinkMetadata) GetAttributeAsString(key string) (string, bool) {
	if value, exists := lm.GetAttribute(key); exists {
		if strValue, ok := value.(string); ok {
			return strValue, true
		}
		return fmt.Sprintf("%v", value), true
	}
	return "", false
}

func (lm *LinkMetadata) HasAttribute(key string) bool {
	_, exists := lm.GetAttribute(key)
	return exists
}

func (lm *LinkMetadata) RemoveAttribute(key string) {
	if lm.Attributes != nil {
		delete(lm.Attributes, key)
	}
}

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

func (lm *LinkMetadata) AttributeCount() int {
	if lm.Attributes == nil {
		return 0
	}
	return len(lm.Attributes)
}

func (lm *LinkMetadata) IsEmpty() bool {
	return lm.Href == "" && len(lm.Attributes) == 0
}

func (lm *LinkMetadata) HasSignificantMetadata() bool {
	if lm.Attributes == nil {
		return false
	}
	for key := range lm.Attributes {
		if key != "href" {
			return true
		}
	}
	return false
}

func (lm *LinkMetadata) Copy() LinkMetadata {
	out := LinkMetadata{
		Href:       lm.Href,
		Attributes: make(map[string]any),
	}
	if lm.Attributes != nil {
		for key, value := range lm.Attributes {
			out.Attributes[key] = value
		}
	}
	return out
}

func ExtractLinkMetadata(mark adf_types.ADFMark) LinkMetadata {
	metadata := LinkMetadata{Attributes: make(map[string]any)}
	if mark.Attrs != nil {
		if href, ok := mark.Attrs["href"].(string); ok {
			metadata.Href = href
		}
		for key, value := range mark.Attrs {
			metadata.Attributes[key] = value
		}
	}
	return metadata
}

func BuildADFMark(metadata LinkMetadata) adf_types.ADFMark {
	attrs := make(map[string]any)
	if metadata.Attributes != nil {
		for key, value := range metadata.Attributes {
			attrs[key] = value
		}
	}
	if metadata.Href != "" {
		attrs["href"] = metadata.Href
	}
	return adf_types.ADFMark{
		Type:  adf_types.MarkTypeLink,
		Attrs: attrs,
	}
}

func IsInternalHref(href string) bool {
	if href == "" {
		return false
	}
	if len(href) > 7 && (href[:8] == "https://" || href[:7] == "http://") {
		return false
	}
	if len(href) > 6 && href[:7] == "mailto:" {
		return false
	}
	if len(href) > 5 && href[:6] == "ftp://" {
		return false
	}
	if len(href) >= 2 && href[:2] == "//" {
		return false
	}
	if href[0] == '/' {
		return true
	}
	return false
}

func ClassifyLinkFromMetadata(metadata LinkMetadata) LinkType {
	isInternal := IsInternalHref(metadata.Href)
	if !isInternal {
		return WebLink
	}
	if metadata.HasSignificantMetadata() {
		return ComplexInternalLink
	}
	return SimpleInternalLink
}

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
		sort.Strings(metadataKeys)
	}

	return LinkClassification{
		Type:         linkType,
		HasMetadata:  hasMetadata,
		IsInternal:   isInternal,
		MetadataKeys: metadataKeys,
	}
}
