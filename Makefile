VERSION ?= $(shell git describe --tags)

IMAGE = yieldr/vulcand-ingress
PKG = github.com/yieldr/vulcand-ingress
PKGS = $(shell go list ./... | grep -v /vendor/)

LDFLAGS = "-s -w -X code.yieldr.com/vulcand-ingress/pkg/version.Version=$(VERSION)"

OS ?= darwin
ARCH ?= amd64

build:
	@CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -o bin/vulcand-ingress-$(OS)-$(ARCH) -a -tags netgo -ldflags $(LDFLAGS)

test:
	@go test $(PKGS)

lint:
	@for pkg in $(PKGS) ; do golint $$pkg ; done

vet:
	@go vet $(PKGS)

docker-image:
	@docker build -t $(IMAGE):$(VERSION) .
	@docker tag $(IMAGE):$(VERSION) $(IMAGE):latest
	@echo " ---> $(IMAGE):$(VERSION)\n ---> $(IMAGE):latest"

docker-push:
	@docker push $(IMAGE):$(VERSION)
	@docker push $(IMAGE):latest
