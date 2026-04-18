package converter

import (
	"errors"

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/placeholder"
)

// ConversionContext provides context information during conversion
type ConversionContext struct {
	Strategy       ConversionStrategy
	PreserveAttrs  bool
	NestedLevel    int
	ParentNodeType ADFNodeType
	RoundTripMode  bool
	ErrorRecovery  bool
	ListDepth      int // Current nesting depth for lists (1 = top level)

	// Classifier and Manager for handling preserved nodes in element converters
	Classifier         ContentClassifier
	PlaceholderManager placeholder.Manager
	PlaceholderSession *placeholder.EditSession
}

// ConversionContextManager manages conversion context during processing
type ConversionContextManager struct {
	contexts []ConversionContext
	current  int
}

// NewConversionContextManager creates a new context manager
func NewConversionContextManager() *ConversionContextManager {
	return &ConversionContextManager{
		contexts: make([]ConversionContext, 0),
		current:  -1,
	}
}

// PushContext adds a new context level for nested processing
func (ccm *ConversionContextManager) PushContext(context ConversionContext) {
	ccm.contexts = append(ccm.contexts, context)
	ccm.current = len(ccm.contexts) - 1
}

// PopContext removes the current context level
func (ccm *ConversionContextManager) PopContext() ConversionContext {
	if ccm.current < 0 {
		return ConversionContext{} // Return empty context if none available
	}

	context := ccm.contexts[ccm.current]
	ccm.contexts = ccm.contexts[:ccm.current]
	ccm.current--

	return context
}

// GetCurrentContext returns the current context
func (ccm *ConversionContextManager) GetCurrentContext() ConversionContext {
	if ccm.current < 0 {
		return ConversionContext{
			Strategy:       StandardMarkdown,
			PreserveAttrs:  true,
			NestedLevel:    0,
			ParentNodeType: "",
			RoundTripMode:  true,
			ErrorRecovery:  true,
		}
	}

	return ccm.contexts[ccm.current]
}

// GetNestedLevel returns the current nesting level
func (ccm *ConversionContextManager) GetNestedLevel() int {
	return len(ccm.contexts)
}

// CreateChildContext creates a new context for a child node
func (ccm *ConversionContextManager) CreateChildContext(node adf_types.ADFNode, strategy ConversionStrategy) ConversionContext {
	parent := ccm.GetCurrentContext()

	return ConversionContext{
		Strategy:       strategy,
		PreserveAttrs:  parent.PreserveAttrs,
		NestedLevel:    parent.NestedLevel + 1,
		ParentNodeType: ADFNodeType(node.Type),
		RoundTripMode:  parent.RoundTripMode,
		ErrorRecovery:  parent.ErrorRecovery,
	}
}

// ShouldPreserveAttributes returns true if attributes should be preserved in current context
func (ccm *ConversionContextManager) ShouldPreserveAttributes() bool {
	context := ccm.GetCurrentContext()
	return context.PreserveAttrs
}

// IsInRoundTripMode returns true if conversion is in round-trip mode
func (ccm *ConversionContextManager) IsInRoundTripMode() bool {
	context := ccm.GetCurrentContext()
	return context.RoundTripMode
}

// ShouldEnableErrorRecovery returns true if error recovery should be enabled
func (ccm *ConversionContextManager) ShouldEnableErrorRecovery() bool {
	context := ccm.GetCurrentContext()
	return context.ErrorRecovery
}

// GetParentStrategy returns the strategy of the parent context
func (ccm *ConversionContextManager) GetParentStrategy() ConversionStrategy {
	if ccm.current < 1 {
		return StandardMarkdown // No parent or only one level
	}

	return ccm.contexts[ccm.current-1].Strategy
}

// IsNestingTooDeep returns true if nesting level exceeds safe limits
func (ccm *ConversionContextManager) IsNestingTooDeep() bool {
	const MaxNestingLevel = 10 // Reasonable limit for nested content
	return ccm.GetNestedLevel() > MaxNestingLevel
}

// CreateRootContext creates the initial root context for document conversion
func CreateRootContext(strategy ConversionStrategy, roundTripMode bool) ConversionContext {
	return ConversionContext{
		Strategy:       strategy,
		PreserveAttrs:  true,
		NestedLevel:    0,
		ParentNodeType: NodeDoc,
		RoundTripMode:  roundTripMode,
		ErrorRecovery:  true,
	}
}

// CreateTableContext creates a specialized context for table processing
func CreateTableContext(parent ConversionContext) ConversionContext {
	return ConversionContext{
		Strategy:       MarkdownTable,
		PreserveAttrs:  parent.PreserveAttrs,
		NestedLevel:    parent.NestedLevel + 1,
		ParentNodeType: NodeTable,
		RoundTripMode:  parent.RoundTripMode,
		ErrorRecovery:  parent.ErrorRecovery,
	}
}

// CreateTaskListContext creates a specialized context for task list processing
func CreateTaskListContext(parent ConversionContext) ConversionContext {
	return ConversionContext{
		Strategy:       MarkdownTaskList,
		PreserveAttrs:  parent.PreserveAttrs,
		NestedLevel:    parent.NestedLevel + 1,
		ParentNodeType: NodeTaskList,
		RoundTripMode:  parent.RoundTripMode,
		ErrorRecovery:  parent.ErrorRecovery,
	}
}

// CreateXMLContext creates a specialized context for XML-preserved elements
func CreateXMLContext(parent ConversionContext, nodeType ADFNodeType) ConversionContext {
	return ConversionContext{
		Strategy:       XMLPreserved,
		PreserveAttrs:  true, // Always preserve attributes for XML elements
		NestedLevel:    parent.NestedLevel + 1,
		ParentNodeType: nodeType,
		RoundTripMode:  parent.RoundTripMode,
		ErrorRecovery:  parent.ErrorRecovery,
	}
}

// CreateErrorRecoveryContext creates a context for error recovery scenarios
func CreateErrorRecoveryContext(parent ConversionContext) ConversionContext {
	return ConversionContext{
		Strategy:       Placeholder, // Use placeholder for safety
		PreserveAttrs:  true,
		NestedLevel:    parent.NestedLevel,
		ParentNodeType: parent.ParentNodeType,
		RoundTripMode:  parent.RoundTripMode,
		ErrorRecovery:  true,
	}
}

// ValidateContext checks if a conversion context is valid
func ValidateContext(context ConversionContext) error {
	// Check for reasonable nesting depth
	if context.NestedLevel > 20 {
		return errors.New("nesting level too deep")
	}

	if context.NestedLevel > 0 {
		switch context.ParentNodeType {
		case NodeTable:
			if context.Strategy != MarkdownTable && context.Strategy != StandardMarkdown {
				return errors.New("invalid strategy for table context")
			}
		case NodeTaskList:
			if context.Strategy != MarkdownTaskList && context.Strategy != StandardMarkdown {
				return errors.New("invalid strategy for task list context")
			}
		}
	}

	return nil
}
