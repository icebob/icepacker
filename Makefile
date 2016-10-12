SHELL=C:/Windows/System32/cmd.exe

# Get the current full sha from git
GITSHA:=$(shell git rev-parse HEAD)

# Get the current local branch name from git (if we can, this may be blank)
GITBRANCH:=$(shell git symbolic-ref --short HEAD)

LDFLAGS:="-X main.GitCommit=${GITSHA}"

default: get test build install 

get:
	go get -t ./...
	go get github.com/smartystreets/goconvey
	go get github.com/mitchellh/gox

build: clean
	@gox -os="windows linux" -arch="386 amd64" -ldflags ${LDFLAGS} -output="releases/{{.OS}}-{{.Arch}}/{{.Dir}}" ./cmd/ipack

build-all: clean
	@gox -output="releases/{{.OS}}-{{.Arch}}/{{.Dir}}" ./cmd/ipack

clean:
	@rm -rf releases

install:
	@go install ./cmd/ipack

run:
	@go run ./cmd/ipack/main.go

ci:
	@goconvey

test:
	@go test -v -timeout 60s -race ./...

cover:
	@go test -parallel 4 -coverprofile=coverage.out ./src
	@go tool cover -html=coverage.out

vet:
	@go vet ./...

test-cli: install
	@test.cmd

