.PHONY: gen-mocks test lint-new swagger

gen-mocks:
	docker run --rm -v "$(CURDIR):/src" -w /src golang:1.26.6 go run github.com/vektra/mockery/v2@v2.53.5 --config .mockery.yaml

test:
	go test ./...

# Vet and formatting checks cover all packages in this new project.
lint-new:
	go vet ./...
	@test -z "$$(gofmt -l $$(git ls-files '*.go') $$(git ls-files --others --exclude-standard '*.go'))" || (echo 'Run gofmt on Go sources'; exit 1)

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.8.1 init
