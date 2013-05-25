DEPS = $(wildcard src/**/*.go)
TARGET = extjob-runner

all: bin/$(TARGET)

bin/$(TARGET): $(DEPS)
	GOPATH=$(shell pwd) go build -o $@ cmd

.PHONY: clean
clean:
	rm -f bin/$(TARGET)
