# adf-converter

[![CI](https://github.com/seflue/adf-converter/actions/workflows/ci.yml/badge.svg)](https://github.com/seflue/adf-converter/actions/workflows/ci.yml)

Lossless ADF <-> Markdown roundtrip library for Go.

## Why

Atlassian Document Format ([spec](https://developer.atlassian.com/cloud/jira/platform/apis/document/structure/))
is a rich JSON format used by Jira and Confluence. Editing it as Markdown in
a terminal is convenient, but naive conversion loses node types and metadata.
This library converts ADF to Markdown and back with full roundtrip fidelity:
non-editable nodes are preserved as placeholders instead of being silently
downgraded.

## The Promise

> If you pass a valid ADF document through `ToMarkdown` and then
> `FromMarkdown`, you get back a semantically equivalent ADF document.
> No node type is silently changed or lost.

See [ARCHITECTURE.md](ARCHITECTURE.md#2-roundtrip-mechanics) for how the conversion is implemented.

## Status

Early development, pre-v1.0, API may change.

## Install

```
go get github.com/seflue/adf-converter
```

## Modules

This repository ships two Go modules. Pick the one that matches your use
case — they layer cleanly, you can also import both.

- **`github.com/seflue/adf-converter`** — the pure ADF↔Markdown library.
  No Glamour, no terminal styling, no ANSI. Use this for editing,
  roundtrips, or any consumer that handles its own rendering.
- **`github.com/seflue/adf-converter/display`** — composes
  adf-converter with [Glamour](https://github.com/charmbracelet/glamour)
  into a single `display.Render` call for themed terminal output.
  Glamour and its transitive dependencies live here, not in the main
  module.

The split is intentional. See [ARCHITECTURE.md §5.3](ARCHITECTURE.md#53-display-mode-vs-roundtrip-mode)
for why display mode is a separate registry composition (no mode flag
in the hot path) and [§7.4](ARCHITECTURE.md#74-display-registry-curated-overrides-for-read-only-rendering)
for how consumers can override individual display renderers.

## Quickstart

ADF↔Markdown roundtrip with the main library:

```go
package main

import (
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
)

func main() {
	doc, err := adf.ParseFromString(`{
		"version": 1,
		"type": "doc",
		"content": [{"type": "paragraph", "content": [{"type": "text", "text": "Hello"}]}]
	}`)
	if err != nil {
		panic(err)
	}

	c := defaults.NewDefaultConverter()

	md, session, err := c.ToMarkdown(*doc)
	if err != nil {
		panic(err)
	}
	fmt.Println(md)

	restored, _, err := c.FromMarkdown(md, session)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", restored)
}
```

## Display Markdown (no terminal styling)

Plain display Markdown straight from the library — no Glamour, no
roundtrip placeholders. Read-only mode. Consumers with their own
renderer (HTML, PDF, alternative terminal library) start here:

```go
package main

import (
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
)

func main() {
	doc, err := adf.ParseFromString(`{
		"version": 1,
		"type": "doc",
		"content": [{"type": "paragraph", "content": [{"type": "text", "text": "Hello"}]}]
	}`)
	if err != nil {
		panic(err)
	}

	c := defaults.NewDisplayConverter()
	md, _, err := c.ToMarkdown(*doc)
	if err != nil {
		panic(err)
	}
	fmt.Print(md)
}
```

Output characteristics: panel as blockquote with icon header, mention
as plain `@Name`, inlineCard as `<URL>` autolink, status as `[Text]`,
sub/sup as Unicode where mappable (`H₂O`, `xⁿ`) with ASCII fallback
otherwise (`_{foo}`, `^Z`), `textColor` mark dropped (text preserved).

## Display Quickstart

Themed terminal rendering through the `display/` submodule:

```go
package main

import (
	"fmt"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/display"
)

func main() {
	doc, err := adf.ParseFromString(`{
		"version": 1,
		"type": "doc",
		"content": [{"type": "paragraph", "content": [{"type": "text", "text": "Hello"}]}]
	}`)
	if err != nil {
		panic(err)
	}

	out, err := display.Render(doc,
		display.WithStyle("dark"),
		display.WithWordWrap(80),
	)
	if err != nil {
		panic(err)
	}
	fmt.Print(out)
}
```

Options: `WithStyle(name)` selects a built-in Glamour style (`auto`,
`dark`, `light`, `dracula`, ...). `WithStyleJSON(json)` loads a custom
theme. `WithWordWrap(n)` sets the column width.

## Tools

`display/_tools/adf-display` is a manual verifier for the display
pipeline. The `_tools/` directory uses Go's underscore-prefix
convention so `go build`/`test`/`vet`/`install` skip it — the binary
is a dev helper, not a shipped CLI. Two `just` recipes wrap it:
`just display [PATH]` for Glamour-rendered output and
`just display-md [PATH]` for the raw display Markdown (Glamour
skipped, useful for diffing renderer output).

## License

MIT
