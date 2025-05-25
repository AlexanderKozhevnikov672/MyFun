download:
	@echo "Downloading"
	go mod download

build:
	@echo "Building"
	go build -o bin/main cmd/main.go

run: build
	@echo "Running"
	./bin/main
