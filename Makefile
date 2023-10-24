all:
	@echo: "Try `make help`"

run:
	go run ./cmd/main.go -c ./examples/config.yaml
