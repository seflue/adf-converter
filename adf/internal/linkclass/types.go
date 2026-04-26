// Package linkclass holds the internal link-classification and strategy-classifier
// machinery formerly exported from the converter package.
package linkclass

// LinkType represents the category of a link for conversion purposes
type LinkType int

const (
	WebLink             LinkType = iota // External web links (https://, http://)
	SimpleInternalLink                  // Atlassian internal links without metadata
	ComplexInternalLink                 // Atlassian internal links with metadata
)

func (lt LinkType) String() string {
	switch lt {
	case WebLink:
		return "WebLink"
	case SimpleInternalLink:
		return "SimpleInternalLink"
	case ComplexInternalLink:
		return "ComplexInternalLink"
	default:
		return "Unknown"
	}
}

// LinkClassification contains the analysis result for a link
type LinkClassification struct {
	Type         LinkType
	HasMetadata  bool
	IsInternal   bool
	MetadataKeys []string
}

func NewLinkClassification(linkType LinkType, hasMetadata bool, isInternal bool, metadataKeys []string) LinkClassification {
	return LinkClassification{
		Type:         linkType,
		HasMetadata:  hasMetadata,
		IsInternal:   isInternal,
		MetadataKeys: metadataKeys,
	}
}

func (lc *LinkClassification) AddMetadataKey(key string) {
	for _, existing := range lc.MetadataKeys {
		if existing == key {
			return
		}
	}
	lc.MetadataKeys = append(lc.MetadataKeys, key)
}

func (lc *LinkClassification) HasMetadataKey(key string) bool {
	for _, metadataKey := range lc.MetadataKeys {
		if metadataKey == key {
			return true
		}
	}
	return false
}

func (lc *LinkClassification) MetadataKeyCount() int {
	return len(lc.MetadataKeys)
}

func (lc *LinkClassification) IsValid() bool {
	if lc.Type == ComplexInternalLink && !lc.HasMetadata {
		return false
	}
	if (lc.Type == SimpleInternalLink || lc.Type == ComplexInternalLink) && !lc.IsInternal {
		return false
	}
	if lc.Type == WebLink && lc.IsInternal {
		return false
	}
	if lc.HasMetadata && len(lc.MetadataKeys) == 0 {
		return false
	}
	return true
}
