LDFLAGS := -ldflags="-s -w"

.PHONY: build clean release tag

build:
	go build $(LDFLAGS) -o bsky-orbit .

clean:
	rm -f bsky-orbit bsky-orbit-*

release: clean
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bsky-orbit-darwin-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bsky-orbit-darwin-amd64 .
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bsky-orbit-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bsky-orbit-linux-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bsky-orbit-windows-amd64.exe .

tag:
	@if [ -z "$(v)" ]; then echo "Usage: make tag v=0.1.0"; exit 1; fi
	git tag v$(v)
	git push origin v$(v)
