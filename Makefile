OUT=bin
DIST=dist
default: *.go scala-script-darwin scala-script-linux

all: clean default

libs:
	./build.sh

tests:
	./test.sh

dist: default
	mkdir -p $(DIST)
	tar cJf $(DIST)/scala-script_$(VERSION)_linux_amd64.tar.xz -C $(OUT)/linux .
	tar cJf $(DIST)/scala-script_$(VERSION)_darwin_amd64.tar.xz -C $(OUT)/darwin .

upload:
	github-release upload \
		-t v$(VERSION) \
		-n scala-script_$(VERSION)_darwin_amd64.tar.xz \
		-f $(DIST)/scala-script_$(VERSION)_darwin_amd64.tar.xz
	github-release upload \
		-t v$(VERSION) \
		-n scala-script_$(VERSION)_linux_amd64.tar.xz \
		-f $(DIST)/scala-script_$(VERSION)_linux_amd64.tar.xz

scala-script-darwin: *.go $(OUT)/darwin/scala-script

$(OUT)/darwin/scala-script: *.go
	mkdir -p $(OUT)/darwin
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o $(OUT)/darwin/scala-script -ldflags "-X main.VERSION=$(VERSION) -s -w -extldflags \"-static\""

scala-script-linux: *.go $(OUT)/linux/scala-script

$(OUT)/linux/scala-script: *.go
	mkdir -p $(OUT)/linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(OUT)/linux/scala-script -ldflags "-X main.VERSION=$(VERSION) -s -w -extldflags \"-static\""

clean:
	go clean
	rm -rf $(OUT)
	rm -rf $(DIST)
	./clean.sh
