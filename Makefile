.PHONY: fmt vet lint test race vuln check tidy

# Format all Go sources.
fmt:
	gofmt -w .

# Static analysis.
vet:
	go vet ./...

# Full linter (requires golangci-lint v2).
lint:
	golangci-lint run ./...

# Unit tests with coverage.
test:
	go test -covermode=atomic -coverprofile=coverage.out ./...

# Unit tests under the race detector.
race:
	go test -race ./...

# Vulnerability scan (requires the pinned toolchain in go.mod).
vuln:
	govulncheck ./...

# Ensure go.mod/go.sum are tidy.
tidy:
	go mod tidy

# One-shot pre-commit gate mirroring CI.
check: fmt vet lint race vuln
	@echo "All checks passed."
