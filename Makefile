OUTDIR=build/

NOW=$(shell date +%Y%m%d%H%M%S)
GIT_REV=$(shell git rev-parse --short HEAD)
VERSION ?= $(shell \
    if [ -z "`git status --porcelain`" ]; then \
		echo ${GIT_REV}; \
    else \
        echo `whoami`-`git rev-parse --abbrev-ref HEAD`-${GIT_REV}-$(NOW) | sed 's/[^0-9a-zA-Z_\.-]/_/g'; \
	fi)

all: test build

build:
	go build -o ${OUTDIR} -ldflags "-X main.VERSION=${VERSION}"

run:
	go run


test:
	go test


.PHONY: build test run
