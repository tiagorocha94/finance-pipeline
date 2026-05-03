.PHONY: run serve test vet

run:
	go run ./cmd/cli/main.go

serve:
	go run ./cmd/server/main.go

test:
	go test ./...

vet:
	go vet ./...