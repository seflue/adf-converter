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
// Classification and HTML-wrapper parsing are internal implementation details; consumers
// drive link handling through the ContentClassifier interface and the default classifier
// returned by [NewDefaultClassifier].
//
// # Conversion Strategies
//
// The converter supports three conversion strategies:
//
//   - StandardMarkdown: Direct [text](url) format for simple links
//   - HTMLWrapped: <a attr="value">[text](url)</a> format for complex links
//   - Placeholder: For elements requiring special handling
//
// # Usage Examples
//
// ## Basic Round-Trip Conversion
//
//	converter := NewDefaultConverter()
//	markdown, restored, err := ConvertRoundTrip(converter, adfDocument)
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
//	restoredDocument, deletions, err := converter.FromMarkdown(editedMarkdown, session)
//	if err != nil {
//		return err
//	}
//
// # Collaborative Safety
//
// The converter preserves all ADF metadata through conversion cycles, uses standard
// markdown for simple links (maximum tool compatibility), and falls back to HTML
// wrappers only when required to preserve metadata.
package adf
