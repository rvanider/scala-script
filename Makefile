
default: scala-script-darwin scala-script-linux

all: clean default

libs:
	./build.sh

tests:
	./test.sh

scala-script-darwin: *.go
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o scala-script-darwin -ldflags '-s -w -extldflags "-static"'

scala-script-linux: *.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o scala-script-linux -ldflags '-s -w -extldflags "-static"'

clean:
	go clean
	rm -f scala-script-linux
	rm -f scala-script-darwin
	./clean.sh
