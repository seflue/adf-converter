package elements

import (
	"fmt"

	"github.com/seflue/adf-converter/adf_types"
)

// XMLPreservedConverter handles ADF-specific elements via XML encoding
type XMLPreservedConverter struct {
	marshaler XMLMarshaler
}

func NewXMLPreservedConverter() *XMLPreservedConverter {
	return &XMLPreservedConverter{
		marshaler: NewDefaultXMLMarshaler(),
	}
}

func (xpc *XMLPreservedConverter) ToMarkdown(node adf_types.ADFNode, context ConversionContext) (EnhancedConversionResult, error) {
	xmlData, err := xpc.marshaler.MarshalToXML(node)
	if err != nil {
		return EnhancedConversionResult{}, fmt.Errorf("failed to marshal node to XML: %w", err)
	}

	return EnhancedConversionResult{
		Content:           string(xmlData),
		Strategy:          XMLPreserved,
		ElementsConverted: 1,
		PreservedAttrs:    node.Attrs,
	}, nil
}

func (xpc *XMLPreservedConverter) FromMarkdown(markdown string, context ConversionContext) (adf_types.ADFNode, error) {
	nodeType := string(context.ParentNodeType)
	return xpc.marshaler.UnmarshalFromXML([]byte(markdown), nodeType)
}

func (xpc *XMLPreservedConverter) CanHandle(nodeType ADFNodeType) bool {
	supportedElements := xpc.marshaler.GetSupportedElements()
	nodeTypeStr := string(nodeType)

	for _, element := range supportedElements {
		if element == nodeTypeStr {
			return true
		}
	}

	return false
}

func (xpc *XMLPreservedConverter) GetStrategy() ConversionStrategy {
	return XMLPreserved
}

func (xpc *XMLPreservedConverter) ValidateInput(input interface{}) error {
	switch v := input.(type) {
	case adf_types.ADFNode:
		if !xpc.CanHandle(ADFNodeType(v.Type)) {
			return fmt.Errorf("XML preserved converter cannot handle node type: %s", v.Type)
		}
		return nil
	case string:
		return xpc.marshaler.ValidateXML([]byte(v))
	default:
		return fmt.Errorf("XML preserved converter cannot validate input of type: %T", input)
	}
}
