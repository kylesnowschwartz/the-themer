# the-themer development commands

# Run all checks: vet + tests
check: vet test

# Build the binary
build:
    go build -o the-themer .

# Run go vet
vet:
    go vet ./...

# Run tests
test:
    go test ./... -v

# Run tests without verbose output
test-quiet:
    go test ./...

# Generate a theme (usage: just generate testdata/bleu.toml /tmp/out)
generate input *args="":
    go run . generate --input {{input}} {{args}}

# Run palette audit (APCA + OKLCH) across all themes
audit:
    uv run scripts/contrast-audit.py themes/*/palette.toml

# Run palette audit for a specific theme
audit-theme theme:
    uv run scripts/contrast-audit.py themes/{{theme}}/palette.toml

# Preview palette colors in terminal
preview *args:
    uv run scripts/preview-palette.py {{args}}

# Preview a specific theme
preview-theme theme:
    uv run scripts/preview-palette.py themes/{{theme}}/palette.toml

# Clean build artifacts
clean:
    rm -f the-themer
