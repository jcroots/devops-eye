BINARY = devops-eye
CMD = ./cmd/devops-eye

.PHONY: default build test lint clean run

default: build

build:
	go build -o $(BINARY) $(CMD)

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f $(BINARY)

run: build
	./$(BINARY)
