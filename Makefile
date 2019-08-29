.PHONY: test
test:
	go test -coverprofile cover.out -race ./...
	go tool cover -html=cover.out -o cover.html
