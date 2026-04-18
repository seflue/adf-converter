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

# Release tagging und GitHub Release erstellen
release VERSION:
    @git diff-index --quiet HEAD -- || { echo "working tree not clean"; exit 1; }
    just check
    git tag -a {{VERSION}} -m "Release {{VERSION}}"
    git push origin {{VERSION}}
    gh release create {{VERSION}} --generate-notes
