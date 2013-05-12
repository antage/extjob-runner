DEPS = $(wildcard src/**/*.go)

all: bin/videoconverter

bin/videoconverter: $(DEPS)
	GOPATH=$(shell pwd) go build -o $@ cmd

.PHONY: clean
clean:
	rm -f bin/videoconverter
