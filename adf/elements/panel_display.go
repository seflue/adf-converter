package elements

import (
	"fmt"
	"strings"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/internal/convresult"
)

// panelDisplayMeta describes how each ADF panel type renders in display mode:
// the icon prefix, the uppercase header label, and the hex color used by the
// display/ Glamour bridge to tint the header.
type panelDisplayMeta struct {
	icon  string
	label string
	color string
}

// panelDisplayTable maps ADF panelType to its display-mode metadata.
// "tip" is rendered verbatim even though edit-mode collapses it to "note":
// the display path stays lossless for the read path.
var panelDisplayTable = map[string]panelDisplayMeta{
	"info":    {icon: "ℹ️", label: "INFO", color: "#0052CC"},
	"warning": {icon: "⚠️", label: "WARNING", color: "#FF991F"},
	"error":   {icon: "❌", label: "ERROR", color: "#DE350B"},
	"success": {icon: "✅", label: "SUCCESS", color: "#00875A"},
	"note":    {icon: "✍️", label: "NOTE", color: "#6554C0"},
	"tip":     {icon: "💡", label: "TIP", color: "#FFAB00"},
}

// panelDisplayRenderer renders panels as Markdown blockquotes with an icon
// header. Edit-mode uses the :::type fenced-div syntax which Glamour leaks
// as visible ":::info" lines — display mode swaps to a blockquote so the
// terminal output stays clean.
type panelDisplayRenderer struct{}

// NewPanelDisplayRenderer returns a display-mode renderer for panel nodes.
func NewPanelDisplayRenderer() adf.Renderer {
	return &panelDisplayRenderer{}
}

func (r *panelDisplayRenderer) ToMarkdown(node adf.Node, context adf.ConversionContext) (adf.RenderResult, error) {
	builder := convresult.NewRenderResultBuilder(adf.StandardMarkdown)

	meta := r.lookupMeta(node)

	body, err := r.renderBody(node, context)
	if err != nil {
		return adf.RenderResult{}, err
	}

	header := `<span style="color: ` + meta.color + `">` + meta.icon + " **" + meta.label + "**</span>"

	var out strings.Builder
	out.WriteString("> " + header + "\n")
	out.WriteString(">\n")
	for _, line := range strings.Split(body, "\n") {
		if line == "" {
			out.WriteString(">\n")
			continue
		}
		out.WriteString("> " + line + "\n")
	}
	out.WriteString("\n")

	builder.AppendContent(out.String())
	builder.IncrementConverted()
	return builder.Build(), nil
}

func (r *panelDisplayRenderer) FromMarkdown(_ []string, _ int, _ adf.ConversionContext) (adf.Node, int, error) {
	return adf.Node{}, 0, fmt.Errorf("panel display renderer is read-only")
}

func (r *panelDisplayRenderer) lookupMeta(node adf.Node) panelDisplayMeta {
	panelType := "info"
	if node.Attrs != nil {
		if pt, ok := node.Attrs["panelType"].(string); ok && pt != "" {
			panelType = pt
		}
	}
	if meta, ok := panelDisplayTable[panelType]; ok {
		return meta
	}
	return panelDisplayTable["info"]
}

// renderBody renders the panel's children through the active registry and
// trims trailing newlines so the blockquote-prefixing loop sees stable input.
func (r *panelDisplayRenderer) renderBody(node adf.Node, context adf.ConversionContext) (string, error) {
	var b strings.Builder
	for _, child := range node.Content {
		childRenderer, _ := context.Registry.Lookup(adf.NodeType(child.Type))
		if childRenderer == nil {
			continue
		}
		childResult, err := childRenderer.ToMarkdown(child, context)
		if err != nil {
			return "", fmt.Errorf("converting panel child %s: %w", child.Type, err)
		}
		b.WriteString(childResult.Content)
	}
	return strings.TrimRight(b.String(), "\n"), nil
}
