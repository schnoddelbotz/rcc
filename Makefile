
export CGO_ENABLED=0
CMD := rcc

all: $(CMD)

$(CMD): *.go */*.go
	go build

install:
	go install

git-setup:
	grep 'make lint test' .git/hooks/pre-push || echo 'make lint test' >> .git/hooks/pre-push
	chmod +x .git/hooks/pre-push

lint:
	go fix -diff ./...
	golangci-lint run

test:
	go test -race

cross: lint test
	GOARCH=arm64 GOOS=darwin  go build -ldflags '-w -s' -o rcc_darwin_arm64
	GOARCH=amd64 GOOS=linux   go build -ldflags '-w -s' -o rcc_linux_amd64
	GOARCH=amd64 GOOS=windows go build -ldflags '-w -s' -o rcc_windows_amd64

coverage:
	go test -race -coverprofile cover.out ./...
	go tool cover -func cover.out

coverage_html: coverage
	go tool cover -html=cover.out

clean:
	go clean -testcache
	rm -rf $(CMD) $(CMD)-output.* cover.out rcc_*
