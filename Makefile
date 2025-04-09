BINARY_NAME=insta-gif.exe

build:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)
