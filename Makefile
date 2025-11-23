.PHONY: lint fmt ci generate test
lint:
	golangci-lint run

fmt:
	gofumpt -l -w .

ci: fmt lint
