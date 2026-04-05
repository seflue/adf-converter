# adf-converter — Go-Library fuer ADF-Roundtrip-Konvertierung

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
