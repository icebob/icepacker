set GOOS=windows

set GOARCH=386
go build -o release\ipack.exe cmd\ipack\main.go

set GOARCH=amd64
go build -o release\ipack-x64.exe cmd\ipack\main.go

set GOOS=linux

set GOARCH=386
go build -o release\ipack cmd\ipack\main.go

set GOARCH=amd64
go build -o release\ipack-x64 cmd\ipack\main.go

