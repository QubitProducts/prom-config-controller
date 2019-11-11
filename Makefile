
.PHONY: deploy

all: prom-config-controller

update-codegen: 
	./hack/update-codegen.sh

update-crd:
	echo "don't run this, you have to edit the CRD afterward at present"
	go run sigs.k8s.io/controller-tools/cmd/controller-gen crd:trivialVersions=true paths="./..." output:crd:artifacts:config=helm/templates

prom-config-controller: go.mod go.sum $(shell find . -name "*.go")
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

