go test -parallel 4 -coverprofile=coverage.out .\...
go tool cover -html=coverage.out