
CMD := retrospective-code-coverage

all: $(CMD)

$(CMD): *.go */*.go
	go build

clean:
	go clean -testcache
	rm -rf $(CMD)
