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

## Quickstart

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

## License

MIT
