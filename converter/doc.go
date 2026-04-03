// Package converter provides bidirectional ADF ↔ Markdown conversion with enhanced link support.
//
// The converter maintains 100% round-trip fidelity for Atlassian Document Format (ADF) documents,
// enabling safe collaborative editing through markdown while preserving all ADF metadata.
//
// # Enhanced Link Support
//
// The converter includes sophisticated link handling that preserves Atlassian metadata
// while providing readable markdown output. Links are classified into three types:
//
//   - Web Links: External https:// and http:// URLs → Standard markdown: [text](url)
//   - Simple Internal Links: Atlassian paths without metadata → Standard markdown: [text](/path)
//   - Complex Internal Links: Atlassian paths with metadata → HTML wrapper: <a meta="value">[text](/path)</a>
//
// # Link Classification System
//
// The LinkClassifier analyzes ADF link marks to determine the appropriate conversion strategy:
//
//	classifier := &DefaultLinkClassifier{}
//	classification := classifier.ClassifyLink(adfMark)
//	strategy := classifier.DetermineStrategy(classification)
//
// Classification is based on:
//   - URL pattern detection (https://, http://, or internal paths starting with /)
//   - Metadata presence analysis (attributes beyond href)
//   - Conversion strategy mapping (StandardMarkdown vs HTMLWrapped)
//
// # HTML Link Parsing
//
// For complex internal links, the converter uses HTML wrapper format to preserve metadata
// while maintaining markdown readability inside the tags:
//
//	parser := &DefaultHTMLLinkParser{}
//	result := parser.ParseHTMLLink(`<a title="Page" id="123">[text](/path)</a>`)
//
// The HTML parser handles:
//   - Attribute extraction from HTML tags
//   - Markdown link parsing within HTML content
//   - Malformed HTML detection and graceful fallback
//   - Round-trip metadata preservation
//
// # Conversion Strategies
//
// The converter supports three conversion strategies:
//
//   - StandardMarkdown: Direct [text](url) format for simple links
//   - HTMLWrapped: <a attr="value">[text](url)</a> format for complex links
//   - Placeholder: For elements requiring special handling (future use)
//
// # Usage Examples
//
// ## Basic Round-Trip Conversion
//
//	converter := NewDefaultConverter()
//	markdown, restored, err := converter.ConvertRoundTrip(adfDocument)
//	if err != nil {
//		return err
//	}
//	// markdown contains editable text
//	// restored contains the round-trip ADF document
//
// ## Custom Classification
//
//	classifier := NewDefaultClassifier()
//	manager := placeholder.NewManager()
//	converter := NewConverter(classifier, manager)
//
//	markdown, session, err := converter.ToMarkdown(adfDocument)
//	if err != nil {
//		return err
//	}
//
//	// User edits markdown...
//
//	result, err := converter.FromMarkdown(editedMarkdown, session)
//	if err != nil {
//		return err
//	}
//
//	restoredDocument := result.Document
//	deletions := result.Deletions
//
// ## Link Classification Analysis
//
//	classifier := &DefaultLinkClassifier{}
//
//	// Analyze a web link
//	webLink := adf_types.ADFMark{
//		Type: adf_types.MarkTypeLink,
//		Attrs: map[string]interface{}{"href": "https://example.com"},
//	}
//	classification := classifier.ClassifyLink(webLink)
//	// classification.Type == WebLink
//	// strategy == StandardMarkdown
//
//	// Analyze a complex internal link
//	complexLink := adf_types.ADFMark{
//		Type: adf_types.MarkTypeLink,
//		Attrs: map[string]interface{}{
//			"href": "/internal/page",
//			"title": "Page Title",
//			"id": "123",
//			"space": "PROJ",
//		},
//	}
//	classification = classifier.ClassifyLink(complexLink)
//	// classification.Type == ComplexInternalLink
//	// strategy == HTMLWrapped
//
// # Architecture Patterns
//
// Enhanced link support establishes foundational patterns for comprehensive ADF element support:
//
//   - Classification Pattern: Element type enumeration for conversion strategy determination
//   - Conversion Strategy: Template for all ADF elements (Standard/Wrapped/Placeholder)
//   - Parser Architecture: HTML/XML parsing foundation for complex elements
//   - Round-trip Testing: Validation pattern for all element types
//
// These patterns enable systematic expansion to support all ADF elements while maintaining
// 100% round-trip fidelity and collaborative safety.
//
// # Future ADF Element Support
//
// The enhanced link implementation serves as a reference for future ADF elements:
//
//   - Panel Elements: <adf:panel type="info">[markdown content]</adf:panel>
//   - Status Elements: <adf:status color="green">text</adf:status>
//   - Media Elements: <adf:media id="123">[caption](url)</adf:media>
//   - Table Enhancements: Complex table metadata preservation
//
// # Performance Characteristics
//
// The enhanced link system is designed for performance:
//
//   - Link Classification: O(1) operations based on href patterns
//   - HTML Parsing: Minimal overhead, only for complex links
//   - Memory Usage: No significant increase for simple links
//   - Concurrent Safety: All classifiers and parsers are thread-safe
//
// # Error Handling
//
// The system handles errors gracefully:
//
//   - Invalid HTML: Falls back to plain text preservation
//   - Missing href: Treats as plain text, logs warning
//   - Classification failure: Defaults to StandardMarkdown strategy
//   - Parse errors: Graceful degradation, preserves original text
//
// # Testing Strategy
//
// Comprehensive testing ensures reliability:
//
//   - Contract Tests: Validate interface requirements
//   - Round-trip Tests: Verify conversion fidelity for all link types
//   - Integration Tests: Test mixed documents and complex scenarios
//   - Edge Case Tests: Handle malformed input and unusual conditions
//   - Performance Tests: Benchmark conversion speed and memory usage
//
// # Collaborative Safety
//
// The enhanced link system maintains collaborative safety by:
//
//   - Preserving all ADF metadata through conversion cycles
//   - Using standard markdown for simple links (maximum tool compatibility)
//   - HTML wrapper only when necessary (metadata preservation)
//   - Graceful fallback for unsupported or malformed content
//   - Complete round-trip validation for all link types
//
// markdown editing tools or knowledge of ADF internals.
package converter
