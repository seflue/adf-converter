package display

import (
	"bytes"
	"fmt"
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"github.com/seflue/glamour/v2"
	"github.com/seflue/glamour/v2/ansi"
)

// kindColorSpan is the goldmark NodeKind for ADF text-color spans
// emitted by adf-converter's display-mode renderers.
var kindColorSpan = ast.NewNodeKind("ADFColorSpan")

// colorSpanNode carries the parsed color value through to the Glamour
// element renderer. Its inline child holds the wrapped text segment.
type colorSpanNode struct {
	ast.BaseInline
	color string
}

func (n *colorSpanNode) Kind() ast.NodeKind         { return kindColorSpan }
func (n *colorSpanNode) Dump(src []byte, lvl int)   { ast.DumpHelper(n, src, lvl, nil, nil) }

// colorSpanParser recognises <span style="color: VALUE">…</span> as
// emitted by renderTextColorMarkEdit. It is whitespace-tolerant on the
// "color:" attribute so display- and edit-mode markdown both parse.
type colorSpanParser struct{}

func (colorSpanParser) Trigger() []byte { return []byte{'<'} }

func (colorSpanParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, segment := block.PeekLine()
	openPrefix := []byte(`<span style="color:`)
	if !bytes.HasPrefix(line, openPrefix) {
		return nil
	}
	rest := line[len(openPrefix):]
	// Tolerate optional whitespace after "color:" so both `color: red` and
	// `color:red` parse.
	skipped := 0
	for skipped < len(rest) && (rest[skipped] == ' ' || rest[skipped] == '\t') {
		skipped++
	}
	rest = rest[skipped:]

	endColor := bytes.IndexByte(rest, '"')
	if endColor <= 0 {
		return nil
	}
	color := string(rest[:endColor])
	rest = rest[endColor:]
	if !bytes.HasPrefix(rest, []byte(`">`)) {
		return nil
	}
	rest = rest[2:]

	endTag := []byte(`</span>`)
	endIdx := bytes.Index(rest, endTag)
	if endIdx < 0 {
		return nil
	}

	innerStart := segment.Start + len(openPrefix) + skipped + endColor + 2
	innerStop := innerStart + endIdx
	totalLen := len(openPrefix) + skipped + endColor + 2 + endIdx + len(endTag)

	n := &colorSpanNode{color: color}
	n.AppendChild(n, ast.NewTextSegment(text.NewSegment(innerStart, innerStop)))
	block.Advance(totalLen)
	return n
}

// colorSpanExtender registers the inline parser at a priority that
// outranks goldmark's default RawHTML parser, so the span is captured
// as our node before the HTML fallback (which would strip it via
// bluemonday) sees it.
type colorSpanExtender struct{}

func (colorSpanExtender) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(colorSpanParser{}, 99),
	))
}

// colorSpanElement applies the parsed color to its children via
// Glamour's StyleOverrider hook so nested marks (bold, em, …) keep
// rendering normally and merely inherit the color.
type colorSpanElement struct {
	color    string
	children []ansi.ElementRenderer
}

func (e *colorSpanElement) Render(w io.Writer, ctx ansi.RenderContext) error {
	style := ansi.StylePrimitive{Color: &e.color}
	for _, child := range e.children {
		if r, ok := child.(ansi.StyleOverriderElementRenderer); ok {
			if err := r.StyleOverrideRender(w, ctx, style); err != nil {
				return fmt.Errorf("color span: %w", err)
			}
			continue
		}
		if err := child.Render(w, ctx); err != nil {
			return fmt.Errorf("color span: %w", err)
		}
	}
	return nil
}

func colorSpanFactory(node ast.Node, source []byte, tr *ansi.ANSIRenderer) ansi.Element {
	n := node.(*colorSpanNode)
	var children []ansi.ElementRenderer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		children = append(children, tr.NewElement(c, source).Renderer)
	}
	return ansi.Element{
		Renderer: &colorSpanElement{color: n.color, children: children},
	}
}

// colorSpanCustomRenderer returns the glamour option that wires the
// parser extension and renderer factory together.
func colorSpanCustomRenderer() glamour.TermRendererOption {
	return glamour.WithCustomRenderer(
		colorSpanExtender{},
		map[ast.NodeKind]ansi.ElementFactory{
			kindColorSpan: {
				OwnsChildren: true,
				Build:        colorSpanFactory,
			},
		},
	)
}
