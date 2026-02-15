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

# Clean build artifacts
clean:
    rm -f the-themer
