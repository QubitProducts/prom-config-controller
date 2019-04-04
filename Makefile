
.PHONY: deploy

all: prom-config-controller

prom-config-controller: go.mod go.sum $(shell find . -name "*.go")
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

