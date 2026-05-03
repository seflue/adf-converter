# adf-converter — Go-Library fuer ADF-Roundtrip-Konvertierung

# Tooling selbstkontainiert gegen die lokale go.mod fahren,
# unabhaengig vom Parent-Workspace (atlassian-tools/go.work).
export GOWORK := "off"

# Default: alles pruefen
default: check

# Alle Tests ausfuehren
test *NAME:
    go test ./... {{ if NAME != "" { "-run " + NAME } else { "" } }}

# Kompilier-Check
build:
    go build ./...

# Linting (read-only)
lint:
    golangci-lint run

# Code formatieren
fmt:
    gofmt -w .

# Auto-Fix (Formatting + Lint-Fixes)
fix:
    golangci-lint run --fix

# Alles pruefen (fail fast: billigstes zuerst)
check: build lint test

# Display-Render eines ADF-JSON-Fixtures durch Glamour ins Terminal.
# Tool lebt im display/_tools-Verzeichnis (Underscore-Prefix → von
# go build/test/vet/install automatisch ignoriert, Konvention fuer
# Dev-Helper im Lib-Repo). -C wechselt das go-Modul ohne shell-cwd.
display PATH="adf/defaults/testdata/display-sample.json":
    go -C display run ./_tools/adf-display ../{{PATH}}

# Wie display, aber MD-Output statt Glamour-Render (Debug)
display-md PATH="adf/defaults/testdata/display-sample.json":
    go -C display run ./_tools/adf-display --md ../{{PATH}}

# Release tagging und GitHub Release erstellen
release VERSION:
    @git diff-index --quiet HEAD -- || { echo "working tree not clean"; exit 1; }
    just check
    git tag -a {{VERSION}} -m "Release {{VERSION}}"
    git push origin {{VERSION}}
    gh release create {{VERSION}} --generate-notes
