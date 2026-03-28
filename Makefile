VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS  = -ldflags "-X github.com/csepant/whatchanged/cmd.Version=$(VERSION) -X github.com/csepant/whatchanged/cmd.Commit=$(COMMIT) -X github.com/csepant/whatchanged/cmd.BuildDate=$(DATE)"

.PHONY: build install test lint clean

build:
	go build $(LDFLAGS) -o whatchanged .

install:
	go install $(LDFLAGS) .

test:
	go test ./... -v

lint:
	golangci-lint run

clean:
	rm -f whatchanged
