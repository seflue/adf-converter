package converter

// LinkType represents the category of a link for conversion purposes
type LinkType int

const (
	WebLink             LinkType = iota // External web links (https://, http://)
	SimpleInternalLink                  // Atlassian internal links without metadata
	ComplexInternalLink                 // Atlassian internal links with metadata
)

// String returns a string representation of the LinkType
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

// NewLinkClassification creates a new LinkClassification with the given parameters
func NewLinkClassification(linkType LinkType, hasMetadata bool, isInternal bool, metadataKeys []string) LinkClassification {
	return LinkClassification{
		Type:         linkType,
		HasMetadata:  hasMetadata,
		IsInternal:   isInternal,
		MetadataKeys: metadataKeys,
	}
}

// AddMetadataKey adds a metadata key to the classification if not already present
func (lc *LinkClassification) AddMetadataKey(key string) {
	for _, existing := range lc.MetadataKeys {
		if existing == key {
			return // Already present
		}
	}
	lc.MetadataKeys = append(lc.MetadataKeys, key)
}

// HasMetadataKey checks if a specific metadata key is present
func (lc *LinkClassification) HasMetadataKey(key string) bool {
	for _, metadataKey := range lc.MetadataKeys {
		if metadataKey == key {
			return true
		}
	}
	return false
}

// MetadataKeyCount returns the number of metadata keys
func (lc *LinkClassification) MetadataKeyCount() int {
	return len(lc.MetadataKeys)
}

// IsValid performs basic validation on the classification
func (lc *LinkClassification) IsValid() bool {
	// ComplexInternalLink must have metadata
	if lc.Type == ComplexInternalLink && !lc.HasMetadata {
		return false
	}

	// Internal link types must have IsInternal = true
	if (lc.Type == SimpleInternalLink || lc.Type == ComplexInternalLink) && !lc.IsInternal {
		return false
	}

	// Web links should not be internal
	if lc.Type == WebLink && lc.IsInternal {
		return false
	}

	// HasMetadata should be consistent with MetadataKeys
	if lc.HasMetadata && len(lc.MetadataKeys) == 0 {
		return false
	}

	return true
}
