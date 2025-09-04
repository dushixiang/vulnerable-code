CGO_ENABLED=0 GOOS=linux GOARCH=mips go build -ldflags '-s -w' -o uploader main.go
upx uploader