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

	"github.com/seflue/adf-converter/adf_types"
	"github.com/seflue/adf-converter/converter/defaults"
)

func main() {
	doc, err := adf_types.ParseFromString(`{
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

	result, err := c.FromMarkdown(md, session)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", result.Document)
}
```

## License

MIT
