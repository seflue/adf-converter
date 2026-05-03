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

// kindUnderline is the goldmark NodeKind for ADF underline marks
// emitted by adf-converter as bare <u>…</u> tags.
var kindUnderline = ast.NewNodeKind("ADFUnderline")

// underlineNode wraps the inline content carried by a <u>…</u> tag.
// It carries no attributes — adf-converter never emits attributes on
// the underline tag.
type underlineNode struct {
	ast.BaseInline
}

func (n *underlineNode) Kind() ast.NodeKind       { return kindUnderline }
func (n *underlineNode) Dump(src []byte, lvl int) { ast.DumpHelper(n, src, lvl, nil, nil) }

// underlineParser recognises bare <u>…</u> as emitted by the underline
// mark renderer in adf/elements/text.go. It accepts no attributes —
// any variation falls through to goldmark's default RawHTML handler
// (which bluemonday strips downstream).
//
// Limitation: the inner content is captured as a single raw text
// segment, so nested inline marks (e.g. <strong> inside <u>, or <u>
// inside a textColor span) do not re-parse. Same constraint as the
// ColorSpan extension. Composition is tracked in backlog ac-0132.
type underlineParser struct{}

func (underlineParser) Trigger() []byte { return []byte{'<'} }

func (underlineParser) Parse(_ ast.Node, block text.Reader, _ parser.Context) ast.Node {
	line, segment := block.PeekLine()
	openTag := []byte(`<u>`)
	if !bytes.HasPrefix(line, openTag) {
		return nil
	}
	rest := line[len(openTag):]

	endTag := []byte(`</u>`)
	endIdx := bytes.Index(rest, endTag)
	if endIdx < 0 {
		return nil
	}

	innerStart := segment.Start + len(openTag)
	innerStop := innerStart + endIdx
	totalLen := len(openTag) + endIdx + len(endTag)

	n := &underlineNode{}
	if endIdx > 0 {
		n.AppendChild(n, ast.NewTextSegment(text.NewSegment(innerStart, innerStop)))
	}
	block.Advance(totalLen)
	return n
}

// underlineExtender registers the inline parser at a priority that
// outranks goldmark's default RawHTML parser, mirroring colorSpanExtender.
type underlineExtender struct{}

func (underlineExtender) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(underlineParser{}, 99),
	))
}

// underlineElement applies an underline style to its children via
// Glamour's StyleOverrider hook so nested marks (bold, em, color) keep
// rendering and merely inherit the underline attribute.
type underlineElement struct {
	children []ansi.ElementRenderer
}

func (e *underlineElement) Render(w io.Writer, ctx ansi.RenderContext) error {
	on := true
	style := ansi.StylePrimitive{Underline: &on}
	for _, child := range e.children {
		if r, ok := child.(ansi.StyleOverriderElementRenderer); ok {
			if err := r.StyleOverrideRender(w, ctx, style); err != nil {
				return fmt.Errorf("underline: %w", err)
			}
			continue
		}
		if err := child.Render(w, ctx); err != nil {
			return fmt.Errorf("underline: %w", err)
		}
	}
	return nil
}

func underlineFactory(node ast.Node, source []byte, tr *ansi.ANSIRenderer) ansi.Element {
	n := node.(*underlineNode)
	var children []ansi.ElementRenderer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		children = append(children, tr.NewElement(c, source).Renderer)
	}
	return ansi.Element{
		Renderer: &underlineElement{children: children},
	}
}

// underlineCustomRenderer returns the glamour option that wires the
// parser extension and renderer factory together.
func underlineCustomRenderer() glamour.TermRendererOption {
	return glamour.WithCustomRenderer(
		underlineExtender{},
		map[ast.NodeKind]ansi.ElementFactory{
			kindUnderline: {
				OwnsChildren: true,
				Build:        underlineFactory,
			},
		},
	)
}
