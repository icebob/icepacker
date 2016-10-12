SHELL=C:/Windows/System32/cmd.exe

# Get the current full sha from git
GITSHA:=$(shell git rev-parse HEAD)

# Get the current local branch name from git (if we can, this may be blank)
GITBRANCH:=$(shell git symbolic-ref --short HEAD)

# BUILD_TIME:=`date +%FT%T%z`

# LDFLAGS for build
LDFLAGS=-ldflags="-X main.GitCommit=${GITSHA}"
# LDFLAGS=-ldflags="-X main.GitCommit=${GITSHA} -X main.BuildTime=${BUILD_TIME}"

# Build path style
BUILD_OUTPUT=-output="releases/{{.OS}}-{{.Arch}}/{{.Dir}}"


default: get test build install 

get:
	go get -t ./...
	go get github.com/smartystreets/goconvey
	go get github.com/franciscocpg/gox

build: clean
	@gox -os="windows linux" -arch="386 amd64" ${LDFLAGS} ${BUILD_OUTPUT} ./cmd/ipack

build-all: clean
	@gox ${LDFLAGS} ${BUILD_OUTPUT} ./cmd/ipack

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

