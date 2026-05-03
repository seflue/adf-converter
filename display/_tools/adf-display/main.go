// Command adf-display renders an ADF JSON document for terminal viewing.
// By default it pipes the document through display.Render — adf-converter's
// display-mode Markdown stage followed by Glamour. The --md flag stops at
// the Markdown stage, useful for diffing renderer output during development.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/seflue/adf-converter/adf"
	"github.com/seflue/adf-converter/adf/defaults"
	"github.com/seflue/adf-converter/display"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "adf-display:", err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("adf-display", flag.ContinueOnError)
	fs.SetOutput(stderr)

	mdOnly := fs.Bool("md", false, "print raw display Markdown instead of running Glamour")
	style := fs.String("style", "auto", "Glamour style name (e.g. auto, dark, light, dracula, notty)")
	wrap := fs.Int("wrap", 80, "Glamour word-wrap column (ignored with --md)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	src, err := readSource(fs.Arg(0), stdin)
	if err != nil {
		return err
	}

	doc, err := adf.ParseFromString(string(src))
	if err != nil {
		return fmt.Errorf("parse ADF: %w", err)
	}

	if *mdOnly {
		conv := defaults.NewDisplayConverter()
		md, _, err := conv.ToMarkdown(*doc)
		if err != nil {
			return fmt.Errorf("render display markdown: %w", err)
		}
		_, err = io.WriteString(stdout, md)
		return err
	}

	out, err := display.Render(doc,
		display.WithStyle(*style),
		display.WithWordWrap(*wrap),
	)
	if err != nil {
		return fmt.Errorf("display render: %w", err)
	}
	_, err = io.WriteString(stdout, out)
	return err
}

// readSource returns the bytes from path or stdin. Empty path means stdin
// — the convention from `cat foo.json | adf-display` style usage.
func readSource(path string, stdin io.Reader) ([]byte, error) {
	if path == "" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(path)
}
