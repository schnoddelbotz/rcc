
CMD := retrospective-code-coverage

all: $(CMD)

$(CMD): *.go */*.go
	go build

coverage:
	go test -race -coverprofile cover.out ./...
	go tool cover -func cover.out

coverage_html: coverage
	go tool cover -html=cover.out

clean:
	go clean -testcache
	rm -rf $(CMD)
