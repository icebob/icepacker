default: build

get:
	go get -t ./...

build:

build-all:


install:

ci:

test:
	go test -v -timeout 60s -race ./...

cli-test:


cover:
	go test -parallel 4 -coverprofile=coverage.out ./src
	go tool cover -html=coverage.out

vet:
	go vet ./...


clean:
	@echo "Clearing releases"
	@rm -rf release